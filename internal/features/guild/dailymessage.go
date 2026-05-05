package guild_features

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/locale"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

var DailyMessageFeature = features.New[DailyMessageConfig](
	setupDailyMessageFeature,
	features.WithType(features.GuildFeature),
	features.WithLocalizedName(discord.LocaleFrench, "Message du jour"),
	features.WithDescription(locale.Text{
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
	if deps.Shared.Cron == nil {
		slog.Warn("Cron not available, skipping DailyMessage setup")
		return nil
	}

	client := deps.Bot.Client
	registry := deps.Configs.Features
	guilds := deps.Configs.Guilds

	var errs []error
	for guildID, guildConfig := range guilds.All() {

		cfg, err := features.GetConfigFrom[DailyMessageConfig](registry, guildID)
		if err != nil {
			errs = append(errs, fmt.Errorf("feature config not found: %w", err))
			continue
		}

		schedule, err := cfg.getCronSchdule()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get guild schedule: %w", err))
			continue
		}

		if guildConfig.Timezone != "" {
			schedule = fmt.Sprintf("CRON_TZ=%s %s", guildConfig.Timezone, schedule)
		}

		entryID, err := deps.Shared.Cron.AddFunc(schedule, dailyMessageJob(client, registry, guildID))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to add cron job: %w", err))
			continue
		}
		slog.Debug("Added cron job", slog.Any("entryID", entryID))
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to setup feature: %v", errs)
	}
	return nil
}

var dailyMessageTemplate = locale.Text{
	discord.LocaleEnglishUS: "Hello %s! This is your daily message.",
	discord.LocaleFrench:    "Bonjour %s! Ceci est votre message quotidien.",
}

func dailyMessageJob(client bot.Client, registry *features.FeatureRegistry, guildID snowflake.ID) func() {
	return func() {
		slog.Info("Running daily message job", slog.Any("guildID", guildID))

		cfg, err := features.GetConfigFrom[DailyMessageConfig](registry, guildID)
		if err != nil || !cfg.Enabled {
			slog.Warn(fmt.Sprintf("Feature 'DailyMessage' is disabled for guild '%s'", guildID))
			return
		}

		restClient := client.Rest()

		guild, err := restClient.GetGuild(guildID, false)
		if err != nil {
			slog.Error("failed to get guild: %w", slog.Any("err", err))
			return
		}

		_, err = restClient.CreateMessage(cfg.Channel, discord.NewMessageCreateBuilder().
			SetContentf(dailyMessageTemplate.Using(discord.Locale(guild.PreferredLocale)), guild.Name).
			Build(),
		)
		if err != nil {
			slog.Error("failed to send message: %w", slog.Any("err", err))
		}
	}
}
