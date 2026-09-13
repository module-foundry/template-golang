// Package migrations embeds SQL migrations and applies them at startup.
package migrations

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed *.sql
var FS embed.FS

// Run applies all pending migrations. A session-level advisory lock makes it
// safe to run the same set from several application instances concurrently.
func Run(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("create migration locker: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, FS, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("create goose provider: %w", err)
	}
	defer func() { _ = provider.Close() }()

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	for _, result := range results {
		log.Info("migration applied", slog.Int64("version", result.Source.Version), slog.String("path", result.Source.Path))
	}
	return nil
}
