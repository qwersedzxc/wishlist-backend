package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/KaoriEl/golang-boilerplate/internal/config"
	v1 "github.com/KaoriEl/golang-boilerplate/internal/controller/http/v1"
	"github.com/KaoriEl/golang-boilerplate/internal/database"
	"github.com/KaoriEl/golang-boilerplate/internal/email"
	"github.com/KaoriEl/golang-boilerplate/internal/oauth"
	friendshiprepo "github.com/KaoriEl/golang-boilerplate/internal/repository/friendship"
	rolerepo "github.com/KaoriEl/golang-boilerplate/internal/repository/role"
	userrepo "github.com/KaoriEl/golang-boilerplate/internal/repository/user"
	wishlistrepo "github.com/KaoriEl/golang-boilerplate/internal/repository/wishlist"
	"github.com/KaoriEl/golang-boilerplate/internal/scheduler"
	authuc "github.com/KaoriEl/golang-boilerplate/internal/usecase/auth"
	friendshipuc "github.com/KaoriEl/golang-boilerplate/internal/usecase/friendship"
	wishlistuc "github.com/KaoriEl/golang-boilerplate/internal/usecase/wishlist"
	_ "github.com/jackc/pgx/v5/stdlib" //nolint:revive,nolintlint
	"github.com/pressly/goose/v3"
)

// runMigrations применяет все pending миграции из директории migrationsDir.
func runMigrations(dsn, migrationsDir string, log *slog.Logger) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("runMigrations open: %w", err)
	}
	defer db.Close()

	// Отключаем встроенный логгер goose логируем сами
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("runMigrations set dialect: %w", err)
	}

	current, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("runMigrations get version: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("runMigrations up: %w", err)
	}

	next, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("runMigrations get version after: %w", err)
	}

	if next > current {
		log.Info("migrations applied",
			slog.Int64("from_version", current),
			slog.Int64("to_version", next),
		)
	} else {
		log.Info("migrations: nothing to apply", slog.Int64("version", current))
	}

	return nil
}

// Run запускает HTTP-сервер и блокирует до отмены ctx (graceful shutdown).
func Run(ctx context.Context, cfg *config.Config, db *database.Database, log *slog.Logger) error {
	// Применяем миграции до инициализации репозиториев
	if err := runMigrations(cfg.DBConnectionString, "./migrations", log); err != nil {
		log.Error("migrations failed", "error", err)

		return err
	}

	wishlistRepo := wishlistrepo.New(db.Pool)
	wishlistItemRepo := wishlistrepo.NewItemRepository(db.Pool)
	wishlistUC := wishlistuc.New(wishlistRepo, wishlistItemRepo, log)

	// Инициализация role компонентов
	roleRepo := rolerepo.New(db.Pool, log)

	// Инициализация auth компонентов
	userRepo := userrepo.New(db.Pool)
	authUC := authuc.New(userRepo, roleRepo, cfg.JWTSecret, log)

	// Инициализация friendship компонентов
	friendshipRepo := friendshiprepo.New(db.Pool)
	friendshipUC := friendshipuc.New(friendshipRepo, userRepo, log)

	// Инициализация email сервиса
	emailService := email.New(cfg.SMTP, log)

	oauthProvider, err := oauth.New(cfg.OAuth)
	if err != nil {
		log.Error("failed to init oauth provider", "error", err)

		return err
	}

	router := v1.NewRouter(wishlistUC, authUC, friendshipUC, roleRepo, oauthProvider, cfg.OAuth.Provider, cfg.S3, emailService, log)

	// Запускаем планировщик дней рождения
	birthdayScheduler := scheduler.New(friendshipUC, emailService, log)
	go birthdayScheduler.Start(ctx)

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second, // защита от Slowloris-атаки
	}

	// Запускаем сервер в горутине
	errCh := make(chan error, 1)
	go func() {
		log.Info("starting HTTP server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// Ждём отмены контекста или ошибки сервера
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down HTTP server")
	}

	// Graceful shutdown с таймаутом 10 секунд
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Info("HTTP server stopped")

	return nil
}
