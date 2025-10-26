package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/Newton/internal/config"
)

var Pool *pgxpool.Pool

func NewPostgresPool(cfg config.PostgreSQL) (*pgxpool.Pool, error) {
	password := url.QueryEscape(cfg.Password)
	user := url.QueryEscape(cfg.User)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, cfg.Host, cfg.Port, cfg.Name)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to parse config: %w", err)
	}

	poolConfig.MaxConns = 30
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to create new pool with config: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ERROR: failed to Ping to postgres: %w", err)
	}

	return pool, nil
}
