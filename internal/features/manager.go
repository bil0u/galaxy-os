package features

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/spf13/viper"
)

// Manager manages a feature set.
type Manager struct {
	features      Set
	botConfig     *config.Bot
	guildsConfigs *config.GuildMap
}

func (fm *Manager) getFeaturesDefs(guildID snowflake.ID) []*Feature {
	if guildID == 0 {
		return fm.features.Bot()
	}
	return fm.features.Guild()
}

func (m *Manager) guildFeaturesRaw(guildID snowflake.ID) *viper.Viper {
	if guildID == 0 {
		return m.botConfig.FeaturesDefs
	}

	config := m.guildsConfigs.Get(guildID)
	if config == nil {
		return nil
	}

	return config.FeaturesDefs
}

// ConfigRaw returns the raw feature configuration for a specific guild and feature key.
func (m *Manager) ConfigRaw(guildID snowflake.ID, featureKey string) (*viper.Viper, error) {
	slog.Debug(fmt.Sprintf("Getting feature configuration for guild '%d' and feature '%s'", guildID, featureKey))
	guildFeatures := m.guildFeaturesRaw(guildID)
	if guildFeatures == nil {
		return nil, fmt.Errorf("no feature config exists for guild '%d'", guildID)
	}
	if !guildFeatures.IsSet(featureKey) {
		return nil, fmt.Errorf("no config defined for feature with key '%s' in guild '%d'", featureKey, guildID)
	}

	return guildFeatures.Sub(featureKey), nil
}

// FeatureConfig returns a new feature configuration for a specific guild and feature key.
func (m *Manager) FeatureConfig(guildID snowflake.ID, featureKey string) (Config, error) {
	feature := m.features.WithKey(featureKey)
	if feature == nil {
		return nil, fmt.Errorf("feature not found in the feature set")
	}
	cfg, err := m.ConfigRaw(guildID, featureKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature configuration: %w", err)
	}
	newCfg := reflect.Zero(feature.cfgType).Interface().(Config)
	if err := cfg.Unmarshal(&newCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature configuration: %w", err)
	}
	featureConfig := newCfg
	return featureConfig, featureConfig.Validate()
}

// FeatureConfigFor tries to get a feature configuration for a specific guild and type.
func (m *Manager) FeatureConfigFor(guildID snowflake.ID, cfgType reflect.Type) (Config, error) {
	for _, f := range m.features {
		if f.cfgType == cfgType {
			return m.FeatureConfig(guildID, f.Key)
		}
	}
	return nil, fmt.Errorf("feature not found in the feature set")
}

// SetupFeatures sets up the features, checking for validity first.
func (m *Manager) SetupFeatures(deps SetupDeps) error {
	var errs []error
	for _, feature := range m.features {
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

// SyncCommands syncs the commands to the Discord API.
func (m *Manager) SyncCommands(client bot.Client, guildsIDs []snowflake.ID) error {
	commandsToSync := []discord.ApplicationCommandCreate{}

	for _, f := range m.features {
		commandsToSync = append(commandsToSync, f.CommandsToSync()...)
	}

	if len(commandsToSync) == 0 {
		slog.Info("No commands to sync")
		return nil
	}

	slog.Info("Syncing commands to Discord API...")

	if err := handler.SyncCommands(client, commandsToSync, guildsIDs); err != nil {
		return fmt.Errorf("failed to sync commands: %w", err)
	}

	slog.Info("Commands successfully synced")
	return nil
}

// NewManager creates a new feature manager.
func NewManager(opts ...ManagerOption) (*Manager, error) {
	m := &Manager{}

	for _, opt := range opts {
		opt(m)
	}

	return m, nil
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithSet sets the feature set for the Manager.
func WithSet(fs Set) ManagerOption {
	return func(m *Manager) {
		m.features = fs
	}
}

// WithBotConfig sets the bot configuration for the Manager.
func WithBotConfig(cfg *config.Bot) ManagerOption {
	return func(m *Manager) {
		m.botConfig = cfg
	}
}

// WithGuildsConfigs sets the guilds configurations for the Manager.
func WithGuildsConfigs(cfgs *config.GuildMap) ManagerOption {
	return func(m *Manager) {
		m.guildsConfigs = cfgs
	}
}
