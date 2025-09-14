package db

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

// Connect creates a pgx pool using database.connection_string (or DB_CONNECTION alias).
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	cs := viper.GetString("database.connection_string")
	if cs == "" {
		return nil, fmt.Errorf("database.connection_string is empty")
	}

	cfg, err := pgxpool.ParseConfig(cs)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	// Basic pool tuning; can be extended
	cfg.MaxConns = 10
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}
	log.Info("connected to postgres")
	return pool, nil
}
