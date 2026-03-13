package db

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store holds a connection pool to PostgreSQL.
type Store struct {
	Pool *pgxpool.Pool
}

// New opens a connection pool and runs all pending migrations.
func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	s := &Store{Pool: pool}
	if err := s.runMigrations(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) runMigrations(ctx context.Context) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("db: read migrations dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return fmt.Errorf("db: read migration %s: %w", e.Name(), err)
		}
		if _, err := s.Pool.Exec(ctx, string(data)); err != nil {
			return fmt.Errorf("db: exec migration %s: %w", e.Name(), err)
		}
	}
	return nil
}

// Close closes the connection pool.
func (s *Store) Close() {
	s.Pool.Close()
}
