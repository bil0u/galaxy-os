package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/contracts"
)

// Registry implements contracts.ServiceRegistry with demand-driven activation.
// Services are registered as available, then features declare what they need
// via Require. Only required services are started.
type Registry struct {
	available map[contracts.ServiceID]contracts.Service
	required  map[contracts.ServiceID]bool
	started   []contracts.Service
	byID      map[contracts.ServiceID]contracts.Service // started services for lookup
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		available: make(map[contracts.ServiceID]contracts.Service),
		required:  make(map[contracts.ServiceID]bool),
		byID:      make(map[contracts.ServiceID]contracts.Service),
	}
}

// Register adds a service as available for activation.
func (r *Registry) Register(id contracts.ServiceID, svc contracts.Service) {
	r.available[id] = svc
}

// Require marks services as needed. Called once per feature with its Needs().
func (r *Registry) Require(ids ...contracts.ServiceID) {
	for _, id := range ids {
		r.required[id] = true
	}
}

// StartAll starts only the required services that were registered.
// Services are started in registration order; any failure is fatal.
func (r *Registry) StartAll(ctx context.Context) error {
	// Iterate in a deterministic order based on ServiceID constants.
	order := []contracts.ServiceID{
		contracts.SQLService,
		contracts.EmailService,
		contracts.CronService,
		contracts.OAuthService,
	}

	for _, id := range order {
		if !r.required[id] {
			continue
		}
		svc, ok := r.available[id]
		if !ok {
			return fmt.Errorf("service %d is required but not registered", id)
		}
		slog.Info("starting service", slog.String("service", svc.Name()))
		if err := svc.Start(ctx); err != nil {
			return fmt.Errorf("starting service %q: %w", svc.Name(), err)
		}
		r.started = append(r.started, svc)
		r.byID[id] = svc
	}
	return nil
}

// StopAll stops started services in reverse order.
func (r *Registry) StopAll(ctx context.Context) error {
	var errs []error
	for i := len(r.started) - 1; i >= 0; i-- {
		svc := r.started[i]
		slog.Info("stopping service", slog.String("service", svc.Name()))
		if err := svc.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stopping service %q: %w", svc.Name(), err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("stopping services: %v", errs)
	}
	return nil
}

// Health returns health for all started services.
func (r *Registry) Health(ctx context.Context) []contracts.Health {
	result := make([]contracts.Health, 0, len(r.started))
	for _, svc := range r.started {
		result = append(result, svc.Health(ctx))
	}
	return result
}

// Service returns a started service by ID, or nil if not started.
func (r *Registry) Service(id contracts.ServiceID) contracts.Service {
	return r.byID[id]
}
