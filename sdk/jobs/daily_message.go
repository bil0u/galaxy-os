package jobs

import (
	"log/slog"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/enums"
	"github.com/bil0u/galaxy-os/sdk/utils"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron/v3"
)

var dailyMessageTemplate = utils.LocalizedString{
	discord.LocaleEnglishUS: "Hello %s! This is your daily message.",
	discord.LocaleFrench:    "Bonjour %s! Ceci est votre message quotidien.",
}

func DailyMessageJob(b *sdk.Bot, guildID snowflake.ID) cron.FuncJob {

	// guildConfig := b.Cfg.GetGuildConfig(guildID)

	return func() {

		slog.Info("Running daily message job", slog.Any("guildID", guildID))

		restClient := b.Client.Rest()

		// Get the guild
		guild, err := restClient.GetGuild(guildID, false)
		if err != nil {
			slog.Error("failed to get guild: %w", slog.Any("err", err))
			return
		}

		// targetChannelID := guild.PublicUpdatesChannelID
		targetChannelID := enums.GuildChannelLaboratoire.ID()

		// Send the message
		_, err = restClient.CreateMessage(targetChannelID, discord.NewMessageCreateBuilder().
			SetContentf(dailyMessageTemplate[discord.Locale(guild.PreferredLocale)], guild.Name).
			Build(),
		)
		if err != nil {
			slog.Error("failed to send message: %w", slog.Any("err", err))
		}
	}
}
