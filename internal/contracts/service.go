package contracts

import "context"

// ServiceID identifies a shared service that features can declare as a dependency.
type ServiceID int

const (
	CronService ServiceID = iota
	OAuthService
	SQLService
	EmailService
)

// Service is the lifecycle interface for shared services.
// Logger is NOT a Service — it is infrastructure initialized unconditionally.
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Health(ctx context.Context) Health
	Stop(ctx context.Context) error
}

// Health reports the current state of a service.
type Health struct {
	Name    string
	Status  HealthStatus
	Details map[string]string
}

// HealthStatus represents the operational state of a component.
type HealthStatus int

const (
	StatusUp HealthStatus = iota
	StatusDegraded
	StatusDown
)

// ServiceRegistry collects feature declarations, starts services in
// dependency order, and exposes aggregate health.
type ServiceRegistry interface {
	Require(ids ...ServiceID)
	StartAll(ctx context.Context) error
	StopAll(ctx context.Context) error
	Health(ctx context.Context) []Health
}

// CronScheduler is the feature-facing interface for cron job scheduling.
// Populated in Deps only when a feature declares CronService in Needs().
type CronScheduler interface{}

// OAuthProvider is the feature-facing interface for OAuth operations.
// Populated in Deps only when a feature declares OAuthService in Needs().
type OAuthProvider interface{}
