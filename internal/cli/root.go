package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/qwersedzxc/wishlist-backend/internal/cli/commands"
	"github.com/qwersedzxc/wishlist-backend/internal/config"
	"github.com/qwersedzxc/wishlist-backend/internal/database"
	"github.com/spf13/cobra"
)

// Execute собирает корневую команду с зависимостями и выполняет её.
func Execute(ctx context.Context, cfg *config.Config, log *slog.Logger, db *database.Database) {
	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "Golang Boilerplate CLI",
		Long:  "Утилита командной строки для управления приложением",
	}

	for _, cmd := range commands.All(ctx, cfg, log, db) {
		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
