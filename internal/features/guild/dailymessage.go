package guild_features

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/services"
	"github.com/bil0u/galaxy-os/internal/utils"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

var DailyMessageFeature = features.New[DailyMessageConfig](
	setupDailyMessageFeature,
	features.WithType(features.GuildFeature),
	features.WithLocalizedName(discord.LocaleFrench, "Message du jour"),
	features.WithDescription(utils.LocalizedString{
		discord.LocaleEnglishUS: "Send a message every day at a specific time",
		discord.LocaleFrench:    "Envoie un message tous les jours à une heure spécifique",
	}),
)

type DailyMessageConfig struct {
	Enabled bool
	Channel snowflake.ID
	Time    string
}

func (f DailyMessageConfig) Validate() error {
	var errs []error
	if f.Channel == 0 {
		errs = append(errs, fmt.Errorf("channel is required"))
	}
	if f.Time == "" {
		errs = append(errs, fmt.Errorf("time is required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (cfg DailyMessageConfig) getCronSchdule() (string, error) {
	pattern := regexp.MustCompile(`^([0-9]{1,2}):([0-9]{1,2})$`)
	match := pattern.FindStringSubmatch(cfg.Time)
	if match == nil {
		return "", fmt.Errorf("invalid time format")
	}
	return fmt.Sprintf("%s %s * * *", match[2], match[1]), nil
}

func setupDailyMessageFeature(deps features.SetupDeps) error {
	if deps.Cron == nil {
		slog.Warn("Cron not available, skipping DailyMessage setup")
		return nil
	}

	var errs []error
	for guildID, guildConfig := range config.GuildsCfg.All() {

		config, err := features.GetConfig[DailyMessageConfig](guildID)
		if err != nil {
			errs = append(errs, fmt.Errorf("feature config not found: %w", err))
			continue
		}

		// Get the cron schedule for the guild
		schedule, err := config.getCronSchdule()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get guild schedule: %w", err))
			continue
		}

		// Add the timezone if set in the guild config
		if guildConfig.Timezone != "" {
			schedule = fmt.Sprintf("CRON_TZ=%s %s", guildConfig.Timezone, schedule)
		}

		// Add the cron job
		entryID, err := deps.Cron.AddFunc(schedule, DailyMessageJob(guildID))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to add cron job: %w", err))
			continue
		}
		slog.Debug("Added cron job", slog.Any("entryID", entryID))
	}

	// Return any errors
	if len(errs) > 0 {
		return fmt.Errorf("failed to setup feature: %v", errs)
	}
	return nil
}

var dailyMessageTemplate = utils.LocalizedString{
	discord.LocaleEnglishUS: "Hello %s! This is your daily message.",
	discord.LocaleFrench:    "Bonjour %s! Ceci est votre message quotidien.",
}

func DailyMessageJob(guildID snowflake.ID) func() {

	return func() {

		slog.Info("Running daily message job", slog.Any("guildID", guildID))

		config, err := features.GetConfig[DailyMessageConfig](guildID)
		if err != nil || !config.Enabled {
			slog.Warn(fmt.Sprintf("Feature 'DailyMessage' is disabled for guild '%s'", guildID))
			return
		}

		restClient := services.GetRestClient()

		// Get the guild
		guild, err := restClient.GetGuild(guildID, false)
		if err != nil {
			slog.Error("failed to get guild: %w", slog.Any("err", err))
			return
		}

		// Send the message
		_, err = restClient.CreateMessage(config.Channel, discord.NewMessageCreateBuilder().
			SetContentf(dailyMessageTemplate.Using(discord.Locale(guild.PreferredLocale)), guild.Name).
			Build(),
		)
		if err != nil {
			slog.Error("failed to send message: %w", slog.Any("err", err))
		}
	}
}
