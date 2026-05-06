package platform

import (
	"context"
	"fmt"
	"log/slog"

	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

// Stage represents a single step in the boot pipeline.
// Stages run in slice order; Requires/Provides are validated before execution.
type Stage struct {
	Name     string
	Requires []string // stage names this depends on
	Provides []string // capabilities this stage provides
	Run      func(ctx context.Context, state *BootState) error
}

// ConfigBundle holds the raw config objects loaded during the config stage.
// Fields are typed as any until config types are lifted into the platform package.
type ConfigBundle struct {
	Global any
	Bot    any
	Log    any
}

// BootState accumulates resources across stages.
// Each stage reads from and writes to the fields it owns.
type BootState struct {
	Obs      *Observability
	Spec     BotSpec
	Config   *ConfigBundle
	Client   *disbot.Client
	Router   *handler.Mux
	Guilds   GuildManager
	GuildIDs []snowflake.ID
	Services ServiceRegistry
	Features []Feature
}

// Run executes stages in order after validating the dependency graph.
// On error it logs the failing stage and returns immediately.
func Run(ctx context.Context, spec BotSpec, stages []Stage) (*BootState, error) {
	if err := validateStages(stages); err != nil {
		return nil, fmt.Errorf("validating boot stages: %w", err)
	}

	state := &BootState{
		Spec:   spec,
		Config: &ConfigBundle{},
	}

	for _, s := range stages {
		slog.Info("boot: running stage", slog.String("stage", s.Name))
		if err := s.Run(ctx, state); err != nil {
			slog.Error("boot: stage failed", slog.String("stage", s.Name), slog.Any("error", err))
			return state, fmt.Errorf("stage %q: %w", s.Name, err)
		}
	}

	return state, nil
}

// Shutdown tears down resources in reverse order:
// features (reverse setup order), services, then the Discord gateway client.
// It respects the context deadline for graceful shutdown.
func Shutdown(ctx context.Context, state *BootState) error {
	var errs []error

	// Stop features in reverse setup order.
	for i := len(state.Features) - 1; i >= 0; i-- {
		f := state.Features[i]
		slog.Info("shutdown: stopping feature", slog.String("feature", f.Name()))
		if err := f.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stopping feature %q: %w", f.Name(), err))
		}
	}

	// Stop services via registry.
	if state.Services != nil {
		slog.Info("shutdown: stopping services")
		if err := state.Services.StopAll(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stopping services: %w", err))
		}
	}

	// Close the Discord gateway client.
	if state.Client != nil {
		slog.Info("shutdown: closing gateway")
		state.Client.Close(ctx)
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// validateStages checks that every Requires reference is satisfied by a
// preceding stage's Provides.
func validateStages(stages []Stage) error {
	provided := make(map[string]bool)
	for _, s := range stages {
		for _, req := range s.Requires {
			if !provided[req] {
				return fmt.Errorf("stage %q requires %q, but no preceding stage provides it", s.Name, req)
			}
		}
		for _, p := range s.Provides {
			provided[p] = true
		}
	}
	return nil
}
