package local

import (
	"context"
	"fmt"

	"github.com/RvShivam/postgres-summary-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {

	ctx := context.Background()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(
		ctx,
		poolConfig,
	)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	return pool, nil
}
