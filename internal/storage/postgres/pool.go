package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

// Connect DB wraps pgx pool for future extensions (metrics, health, etc.).
// Connect uses the viper key database.connection_string.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	conn := viper.GetString("database.connection_string")
	if conn == "" {
		return nil, errors.New("missing config: database.connection_string or DB_CONNECTION")
	}
	cfg, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}
	// Reasonable defaults for dev
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pg connect: %w", err)
	}
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}
	log.Info("postgres connected")
	return pool, nil
}
