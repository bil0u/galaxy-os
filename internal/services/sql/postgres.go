package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

var pool *pgxpool.Pool

func Init(ctx context.Context, pgURL string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(pgURL)
	if err != nil {
		return nil, fmt.Errorf("parsing connection pool config: %w", err)
	}

	poolCfg.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	pool = dbpool
	return pool, nil
}

func Pool() *pgxpool.Pool {
	return pool
}
