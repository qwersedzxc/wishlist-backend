package commands

import (
	"context"
	"log/slog"

	"github.com/KaoriEl/golang-boilerplate/internal/config"
	"github.com/KaoriEl/golang-boilerplate/internal/database"
	"github.com/spf13/cobra"
)

// All возвращает список всех cobra-команд с инжектированными зависимостями.
func All(ctx context.Context, cfg *config.Config, log *slog.Logger, db *database.Database) []*cobra.Command {
	return []*cobra.Command{
		NewHealthzCmd(ctx, db, log),
		NewMigrateUpCmd(ctx, cfg, log),
		NewMigrateDownCmd(ctx, cfg, log),
		NewMigrateStatusCmd(ctx, cfg, log),
		NewMigrateCreateCmd(ctx, cfg, log),
	}
}
