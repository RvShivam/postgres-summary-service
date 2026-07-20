package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the minimal interface the repository needs from *pgxpool.Pool.
// Defining it locally means both *pgxpool.Pool (production) and pgxmock
// (tests) satisfy it — no dependency on the pgxmock package in production code.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Compile-time guarantee: *pgxpool.Pool must satisfy DB.
var _ DB = (*pgxpool.Pool)(nil)

type PostgresRepository struct {
	pool DB
}

func NewPostgresRepository(pool DB) Repository {
	return &PostgresRepository{pool: pool}
}
