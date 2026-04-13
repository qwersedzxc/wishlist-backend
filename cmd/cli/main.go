package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/qwersedzxc/wishlist-backend/internal/cli"
	"github.com/qwersedzxc/wishlist-backend/internal/config"
	"github.com/qwersedzxc/wishlist-backend/internal/database"
	"github.com/qwersedzxc/wishlist-backend/internal/logger"
)

func main() {
	cfg := config.MustLoad()
	l := logger.NewLogger("golang-boilerplate-cli", cfg.Environment, cfg.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.New(ctx, cfg.DBConnectionString)
	if err != nil {
		l.Error("database connection failed", "error", err)
		stop()

		return
	}
	defer db.Close()
	cli.Execute(ctx, cfg, l, db)
}
