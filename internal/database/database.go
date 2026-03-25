package database

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database объединяет пул соединений и query-builder с $-плейсхолдерами.
type Database struct {
	Pool *pgxpool.Pool
	Sq   sq.StatementBuilderType
}

// New создаёт пул соединений к PostgreSQL и пингует базу.
func New(ctx context.Context, dsn string) (*Database, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Database{
		Pool: pool,
		Sq:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// Close закрывает пул соединений.
func (db *Database) Close() {
	db.Pool.Close()
}
