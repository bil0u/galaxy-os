package suspicious_interview

import (
	"log/slog"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func OnSuspicousRoleAssigned(b *sdk.Bot) bot.EventListener {

	return bot.NewListenerFunc(func(event *events.GuildMemberUpdate) {

		guildConfig, err := b.Config.GetGuildConfig(event.GuildID)
		if err != nil {
			slog.Error("failed to get guild config: %w", slog.Any("err", err))
			return
		}

		feature, _ := sdk.GetFeature[*SuspiciousInterwiewFeature](guildConfig.Features)

		if !feature.Enabled {
			return
		}

		for _, roleID := range event.Member.RoleIDs {
			slog.Info("Role ID: ", slog.Any("roleID", roleID))

			for _, susRole := range feature.DetectRoles {
				if roleID == susRole {
					// TODO: Implement the suspicious role interview
					return
				}
			}

		}
	})
}
