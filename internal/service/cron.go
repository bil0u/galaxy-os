package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/contracts"
	"github.com/robfig/cron/v3"
)

// CronService wraps robfig/cron as a contracts.Service and contracts.CronScheduler.
type CronService struct {
	scheduler *cron.Cron
}

// NewCronService creates a CronService. Call Start to begin scheduling.
func NewCronService() *CronService {
	return &CronService{}
}

func (s *CronService) Name() string { return "cron" }

func (s *CronService) Start(_ context.Context) error {
	s.scheduler = cron.New(cron.WithLogger(cronLogger{}), cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))
	s.scheduler.Start()
	return nil
}

func (s *CronService) Health(_ context.Context) contracts.Health {
	status := contracts.StatusDown
	if s.scheduler != nil {
		status = contracts.StatusUp
	}
	return contracts.Health{
		Name:   s.Name(),
		Status: status,
	}
}

func (s *CronService) Stop(_ context.Context) error {
	if s.scheduler == nil {
		return nil
	}
	closeCtx := s.scheduler.Stop()
	<-closeCtx.Done()
	return nil
}

// AddFunc adds a cron job. Features type-assert CronScheduler to *CronService.
func (s *CronService) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return s.scheduler.AddFunc(spec, cmd)
}

// Remove removes a scheduled cron entry.
func (s *CronService) Remove(id cron.EntryID) {
	s.scheduler.Remove(id)
}

type cronLogger struct{}

func (cl cronLogger) Info(msg string, keysAndValues ...any) {
	slog.Info(fmt.Sprintf("Cron: %s", msg), keysAndValues...)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...any) {
	slog.Error(fmt.Sprintf("Cron: %s: %v", msg, err), keysAndValues...)
}
