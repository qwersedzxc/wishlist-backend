package commands

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/KaoriEl/golang-boilerplate/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

const migrationsDir = "./migrations"

// openDB открывает соединение с базой данных через pgx stdlib-драйвер.
func openDB(cfg *config.Config, l *slog.Logger) *sql.DB {
	db, err := sql.Open("pgx", cfg.DBConnectionString)
	if err != nil {
		l.Error("migrate: не удалось открыть соединение с БД", "error", err)
		os.Exit(1)
	}

	return db
}

// NewMigrateUpCmd создаёт команду migrate:up — применить все pending миграции.
func NewMigrateUpCmd(_ context.Context, cfg *config.Config, l *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:up",
		Short: "Применить все миграции",
		Run: func(_ *cobra.Command, _ []string) {
			db := openDB(cfg, l)
			defer db.Close()

			goose.SetLogger(goose.NopLogger())
			if err := goose.SetDialect("postgres"); err != nil {
				l.Error("migrate:up set dialect", "error", err)
				os.Exit(1)
			}

			if err := goose.Up(db, migrationsDir); err != nil {
				l.Error("migrate:up failed", "error", err)
				os.Exit(1)
			}

			l.Info("migrate:up: миграции успешно применены")
			fmt.Println("✅ Migrations applied")
		},
	}
}

// NewMigrateDownCmd создаёт команду migrate:down — откатить последнюю миграцию.
func NewMigrateDownCmd(_ context.Context, cfg *config.Config, l *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:down",
		Short: "Откатить последнюю миграцию",
		Run: func(_ *cobra.Command, _ []string) {
			db := openDB(cfg, l)
			defer db.Close()

			goose.SetLogger(goose.NopLogger())
			if err := goose.SetDialect("postgres"); err != nil {
				l.Error("migrate:down set dialect", "error", err)
				os.Exit(1)
			}

			if err := goose.Down(db, migrationsDir); err != nil {
				l.Error("migrate:down failed", "error", err)
				os.Exit(1)
			}

			l.Info("migrate:down: миграция откачена")
			fmt.Println("✅ Migration rolled back")
		},
	}
}

// NewMigrateStatusCmd создаёт команду migrate:status — показать статус всех миграций.
func NewMigrateStatusCmd(_ context.Context, cfg *config.Config, l *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:status",
		Short: "Статус миграций",
		Run: func(_ *cobra.Command, _ []string) {
			db := openDB(cfg, l)
			defer db.Close()

			goose.SetLogger(goose.NopLogger())
			if err := goose.SetDialect("postgres"); err != nil {
				l.Error("migrate:status set dialect", "error", err)
				os.Exit(1)
			}

			if err := goose.Status(db, migrationsDir); err != nil {
				l.Error("migrate:status failed", "error", err)
				os.Exit(1)
			}
		},
	}
}

// NewMigrateCreateCmd создаёт команду migrate:create — создать новый файл миграции.
func NewMigrateCreateCmd(_ context.Context, _ *config.Config, l *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:create [name]",
		Short: "Создать новый файл миграции",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			if err := goose.Create(nil, migrationsDir, args[0], "sql"); err != nil {
				l.Error("migrate:create failed", "error", err)
				os.Exit(1)
			}

			l.Info("migrate:create: файл миграции создан", "name", args[0])
			fmt.Printf("✅ Migration created: %s\n", args[0])
		},
	}
}
