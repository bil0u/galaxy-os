package core

import (
	"context"
	"log/slog"
)

// Observability is initialized at boot stage 0.
// Tracer and Meter are placeholders (any) until OTel is integrated.
type Observability struct {
	Logger *slog.Logger
	Tracer any
	Meter  any
}

// For returns an Observability scoped to a named component.
func (o *Observability) For(name string) *Observability {
	return &Observability{
		Logger: o.Logger.With("component", name),
		Tracer: o.Tracer,
		Meter:  o.Meter,
	}
}

// HealthAggregator collects health from all registered components.
type HealthAggregator interface {
	Register(name string, checker func(ctx context.Context) Health)
	Check(ctx context.Context) []Health
	Status(ctx context.Context) HealthStatus
}
