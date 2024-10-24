package features

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/bil0u/galaxy-os/pkg"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron/v3"
)

func init() {
	pkg.RegisterFeature[DailyMessageFeature]("daily_message")
}

type DailyMessageFeature struct {
	Enabled bool         `toml:"enabled"`
	Channel snowflake.ID `toml:"channel"`
	Time    string       `toml:"schedule"`
}

func (f DailyMessageFeature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Daily message",
		discord.LocaleFrench:    "Message du jour",
	}
}

func (f DailyMessageFeature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Send a message every day at a specific time",
		discord.LocaleFrench:    "Envoie un message tous les jours à une heure spécifique",
	}
}

func (f DailyMessageFeature) IsEnabled() bool {
	return f.Enabled
}

func (f DailyMessageFeature) IsProperlyConfigured() error {
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

func (f DailyMessageFeature) Setup(bot *pkg.Bot) error {

	var errs []error
	for _, guildConfig := range bot.Config.Guilds {

		feature, err := pkg.GetFeature[DailyMessageFeature](guildConfig.Features)
		if err != nil {
			// Means feature is not enabled
			continue
		}

		// Get the cron schedule for the guild
		guildSchdule, err := feature.getCronSchdule()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get guild schedule: %w", err))
			continue
		}

		// Add the timezone if set in the guild config
		if guildConfig.Timezone != "" {
			guildSchdule = fmt.Sprintf("CRON_TZ=%s %s", guildConfig.Timezone, guildSchdule)
		}

		// Add the cron job
		entryID, err := bot.Cron.AddFunc(guildSchdule, DailyMessageJob(bot, &guildConfig))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to add cron job: %w", err))
			continue
		}
		slog.Info("Added cron job", slog.Any("entryID", entryID))
	}

	// Return any errors
	if len(errs) > 0 {
		return fmt.Errorf("failed to setup feature: %v", errs)
	}
	return nil
}

func (cfg DailyMessageFeature) getCronSchdule() (string, error) {
	pattern := regexp.MustCompile(`^[0-9]{1,2}:[0-9]{1,2}$`)
	match := pattern.FindStringSubmatch(cfg.Time)
	if match == nil {
		return "", fmt.Errorf("invalid time format")
	}
	return fmt.Sprintf("%s %s * * *", match[2], match[1]), nil
}

var dailyMessageTemplate = pkg.LocalizedString{
	discord.LocaleEnglishUS: "Hello %s! This is your daily message.",
	discord.LocaleFrench:    "Bonjour %s! Ceci est votre message quotidien.",
}

func DailyMessageJob(bot *pkg.Bot, guildConfig *pkg.GuildConfig) cron.FuncJob {

	return func() {

		slog.Info("Running daily message job", slog.Any("guildID", guildConfig.ID))

		feature, err := pkg.GetFeature[DailyMessageFeature](guildConfig.Features)
		if err != nil || !feature.IsEnabled() {
			return
		}

		restClient := bot.Client.Rest()

		// Get the guild
		guild, err := restClient.GetGuild(guildConfig.ID, false)
		if err != nil {
			slog.Error("failed to get guild: %w", slog.Any("err", err))
			return
		}

		// Send the message
		_, err = restClient.CreateMessage(feature.Channel, discord.NewMessageCreateBuilder().
			SetContentf(dailyMessageTemplate[discord.Locale(guild.PreferredLocale)], guild.Name).
			Build(),
		)
		if err != nil {
			slog.Error("failed to send message: %w", slog.Any("err", err))
		}
	}
}
