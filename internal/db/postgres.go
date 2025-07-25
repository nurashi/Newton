package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/Newton/internal/config"
)



var Pool *pgxpool.Pool

func Connect() error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.App.Database.User,
		config.App.Database.Password,
		config.App.Database.Host,
		config.App.Database.Port,
		config.App.Database.Name,
	)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	var err error 
	Pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("ERROR: failed to connect db: %v", err)
	}

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ERROR: ping db: %w", err)
	}

	return nil
}