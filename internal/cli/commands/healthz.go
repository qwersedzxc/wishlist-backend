package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/KaoriEl/golang-boilerplate/internal/database"
	"github.com/spf13/cobra"
)

// NewHealthzCmd создаёт команду healthz с инжектированными зависимостями.
func NewHealthzCmd(ctx context.Context, db *database.Database, log *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "healthz",
		Short: "Проверить состояние приложения",
		Long:  "Пингует базу данных и выводит OK или завершается с кодом 1",
		Run: func(_ *cobra.Command, _ []string) {
			if err := db.Pool.Ping(ctx); err != nil {
				log.Error("healthz: database ping failed", "error", err)
				fmt.Fprintln(os.Stderr, "FAIL:", err)
				os.Exit(1)
			}
			fmt.Println("OK")
		},
	}
}
