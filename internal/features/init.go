package features

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

// SetupFeatures validates and runs the setup function for each feature in the set.
func SetupFeatures(fs Set, deps SetupDeps) error {
	var errs []error
	for _, feature := range fs {
		if err := feature.IsValid(); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := feature.Setup(deps); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// SyncCommands syncs all registered slash commands to the Discord API.
func SyncCommands(fs Set, client bot.Client, guildsIDs []snowflake.ID) error {
	var commandsToSync []discord.ApplicationCommandCreate

	for _, f := range fs {
		commandsToSync = append(commandsToSync, f.CommandsToSync()...)
	}

	if len(commandsToSync) == 0 {
		slog.Info("No commands to sync")
		return nil
	}

	slog.Info("Syncing commands to Discord API...")

	if err := handler.SyncCommands(client, commandsToSync, guildsIDs); err != nil {
		return fmt.Errorf("syncing commands: %w", err)
	}

	slog.Info("Commands successfully synced")
	return nil
}

// GetConfigFrom returns a typed feature configuration for a specific guild, resolved via the registry.
func GetConfigFrom[T Config](r *FeatureRegistry, guildID snowflake.ID) (T, error) {
	var zero T
	cfg, err := r.FeatureConfigFor(guildID, reflect.TypeOf(zero))
	if err != nil {
		return zero, err
	}
	return cfg.(T), nil
}
