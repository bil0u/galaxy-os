package service

import (
	"context"
	"fmt"

	"github.com/bil0u/galaxy-os/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

// SQLService wraps pgxpool as a platform.Service.
type SQLService struct {
	pgURL string
	pool  *pgxpool.Pool
}

// NewSQLService creates a SQLService with the given connection URL.
func NewSQLService(pgURL string) *SQLService {
	return &SQLService{pgURL: pgURL}
}

func (s *SQLService) Name() string { return "sql" }

func (s *SQLService) Start(ctx context.Context) error {
	poolCfg, err := pgxpool.ParseConfig(s.pgURL)
	if err != nil {
		return fmt.Errorf("parsing connection pool config: %w", err)
	}

	poolCfg.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("creating connection pool: %w", err)
	}

	s.pool = dbpool
	return nil
}

func (s *SQLService) Health(ctx context.Context) platform.Health {
	h := platform.Health{
		Name:   s.Name(),
		Status: platform.StatusDown,
	}
	if s.pool == nil {
		return h
	}
	if err := s.pool.Ping(ctx); err != nil {
		h.Status = platform.StatusDegraded
		h.Details = map[string]string{"error": err.Error()}
		return h
	}
	h.Status = platform.StatusUp
	return h
}

func (s *SQLService) Stop(_ context.Context) error {
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}

// Pool returns the underlying connection pool.
func (s *SQLService) Pool() *pgxpool.Pool {
	return s.pool
}
