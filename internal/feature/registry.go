package feature

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/disgoorg/snowflake/v2"
	"github.com/spf13/viper"
)

// FeatureRegistry resolves feature configurations for guilds.
type FeatureRegistry struct {
	features      Set
	botConfig     *config.Bot
	guildsConfigs *config.GuildMap
}

func (r *FeatureRegistry) guildFeaturesRaw(guildID snowflake.ID) *viper.Viper {
	if guildID == 0 {
		return r.botConfig.FeaturesDefs
	}

	config := r.guildsConfigs.Get(guildID)
	if config == nil {
		return nil
	}

	return config.FeaturesDefs
}

// configRaw returns the raw feature configuration for a specific guild and feature key.
func (r *FeatureRegistry) configRaw(guildID snowflake.ID, featureKey string) (*viper.Viper, error) {
	slog.Debug(fmt.Sprintf("Getting feature configuration for guild '%d' and feature '%s'", guildID, featureKey))
	guildFeatures := r.guildFeaturesRaw(guildID)
	if guildFeatures == nil {
		return nil, fmt.Errorf("no feature config exists for guild '%d'", guildID)
	}
	if !guildFeatures.IsSet(featureKey) {
		return nil, fmt.Errorf("no config defined for feature with key '%s' in guild '%d'", featureKey, guildID)
	}

	return guildFeatures.Sub(featureKey), nil
}

// FeatureConfig returns a typed feature configuration for a specific guild and feature key.
func (r *FeatureRegistry) FeatureConfig(guildID snowflake.ID, featureKey string) (Config, error) {
	feature := r.features.WithKey(featureKey)
	if feature == nil {
		return nil, fmt.Errorf("feature not found in the feature set")
	}
	cfg, err := r.configRaw(guildID, featureKey)
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

// FeatureConfigFor returns a typed feature configuration by config type.
func (r *FeatureRegistry) FeatureConfigFor(guildID snowflake.ID, cfgType reflect.Type) (Config, error) {
	for _, f := range r.features {
		if f.cfgType == cfgType {
			return r.FeatureConfig(guildID, f.Key)
		}
	}
	return nil, fmt.Errorf("feature not found in the feature set")
}

// NewRegistry creates a new FeatureRegistry.
func NewRegistry(fs Set, botCfg *config.Bot, guilds *config.GuildMap) *FeatureRegistry {
	return &FeatureRegistry{
		features:      fs,
		botConfig:     botCfg,
		guildsConfigs: guilds,
	}
}
