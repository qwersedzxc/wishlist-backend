package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/qwersedzxc/wishlist-backend/internal/app"
	"github.com/qwersedzxc/wishlist-backend/internal/config"
	"github.com/qwersedzxc/wishlist-backend/internal/database"
	"github.com/qwersedzxc/wishlist-backend/internal/logger"
)

// @title           Wishlist App API
// @version         1.0
// @description     API для управления списками желаний.
// @host            localhost:8081
// @BasePath        /api/v1.
func main() {
	cfg := config.MustLoad()
	l := logger.NewLogger("golang-boilerplate", cfg.Environment, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.New(ctx, cfg.DBConnectionString)
	if err != nil {
		l.Error("database connection failed", "error", err)
		stop()

		return
	}

	defer db.Close()

	go func() {
		if err := app.Run(ctx, cfg, db, l); err != nil {
			l.Error("app.Run error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	l.Info("shutdown signal received")
}
