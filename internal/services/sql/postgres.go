package sql

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

var (
	PgPool *pgxpool.Pool
)

func Init(ctx context.Context, pgURL string) (*pgxpool.Pool, error) {
	sync.OnceFunc(func() {
		dbpool, err := pgxpool.New(ctx, pgURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
			os.Exit(1)
		}

		dbpool.Config().AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
			pgxUUID.Register(conn.TypeMap())

			return nil
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
			os.Exit(1)
		}

		PgPool = dbpool
	})()

	return PgPool, nil
}
