package service

import (
	"context"
	"sync"

	"github.com/bil0u/galaxy-os/internal/contracts"
)

// Aggregator implements contracts.HealthAggregator with worst-of semantics.
type Aggregator struct {
	mu       sync.RWMutex
	checkers map[string]func(ctx context.Context) contracts.Health
}

// NewAggregator creates an empty Aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{
		checkers: make(map[string]func(ctx context.Context) contracts.Health),
	}
}

// Register adds a named health checker.
func (a *Aggregator) Register(name string, checker func(ctx context.Context) contracts.Health) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.checkers[name] = checker
}

// Check returns health from all registered checkers.
func (a *Aggregator) Check(ctx context.Context) []contracts.Health {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := make([]contracts.Health, 0, len(a.checkers))
	for _, checker := range a.checkers {
		results = append(results, checker(ctx))
	}
	return results
}

// Status returns the worst health status across all checkers.
// An empty aggregator returns StatusUp.
func (a *Aggregator) Status(ctx context.Context) contracts.HealthStatus {
	checks := a.Check(ctx)
	worst := contracts.StatusUp
	for _, h := range checks {
		if h.Status > worst {
			worst = h.Status
		}
	}
	return worst
}
