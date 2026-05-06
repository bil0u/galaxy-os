package feature

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/bil0u/galaxy-os/internal/locale"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

var SelfAssignRolesFeature = New[SelfAssignRolesConfig](
	SelfAssignRolesSetup,
	WithType(GuildFeature),
	WithName(locale.Text{
		discord.LocaleEnglishUS: "Self assign roles",
		discord.LocaleFrench:    "Auto-attribution de rôles",
	}),
	WithDescription(locale.Text{
		discord.LocaleEnglishUS: "The bot will assign roles to itself automatically",
		discord.LocaleFrench:    "Le bot s'attribuera des rôles automatiquement",
	}),
)

type SelfAssignRolesConfig struct {
	Enabled         bool
	ClearRolesFirst bool
	Roles           []snowflake.ID
}

func (f SelfAssignRolesConfig) Validate() error {
	var errs []error
	if len(f.Roles) == 0 {
		errs = append(errs, fmt.Errorf("roles are required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func SelfAssignRolesSetup(deps SetupDeps) error {
	client := deps.Bot.Client
	registry := deps.Configs.Features
	botCfg := deps.Configs.Bot
	guilds := deps.Configs.Guilds
	global := deps.Configs.Global

	client.AddEventListeners(disbot.NewListenerFunc(func(_ *events.Ready) {
		restClient := client.Rest
		for _, guildID := range guilds.IDs(global.Development) {
			cfg, err := GetConfigFrom[SelfAssignRolesConfig](registry, guildID)
			if err != nil {
				slog.Error("Failed to get self-assign roles config", slog.Any("err", err))
				continue
			}

			if !cfg.Enabled {
				slog.Warn(fmt.Sprintf("Feature 'SelfAssignRoles' is disabled for guild '%s'", guildID))
				continue
			}

			existingRoles, err := getAssignedRoles(restClient, guildID, botCfg.ApplicationID)
			if err != nil {
				slog.Error("Failed to get assigned roles", slog.Any("err", err))
				continue
			}

			if cfg.ClearRolesFirst {
				slog.Info(fmt.Sprintf("Clearing roles for guild '%s'", guildID))
				if err := removeBotRoles(restClient, guildID, existingRoles); err != nil {
					slog.Error("Failed to remove roles", slog.Any("err", err))
				}
			}

			if err := assignBotRoles(restClient, guildID, cfg.Roles, existingRoles); err != nil {
				slog.Error("Failed to assign roles to bot", slog.Any("err", err))
			}
		}
	}))
	return nil
}

func getAssignedRoles(restClient rest.Rest, guildID snowflake.ID, applicationID snowflake.ID) ([]snowflake.ID, error) {
	botUser, err := restClient.GetMember(guildID, applicationID)
	if err != nil {
		return nil, err
	}
	return botUser.RoleIDs, nil
}

func removeBotRoles(restClient rest.Rest, guildID snowflake.ID, roles []snowflake.ID) error {
	botUser, err := restClient.GetCurrentUser("")
	if err != nil {
		return err
	}

	for _, roleID := range roles {
		r, err := restClient.GetRole(guildID, roleID)
		if err != nil {
			slog.Error("Failed to get role", slog.Any("err", err))
			continue
		}
		if !r.Managed {
			err := restClient.RemoveMemberRole(guildID, botUser.ID, r.ID)
			if err != nil {
				slog.Error("Failed to remove role", slog.Any("err", err))
				continue
			}
			slog.Info("Successfully removed role", slog.Any("role", r.Name))
		}
	}

	return nil
}

func assignBotRoles(restClient rest.Rest, guildID snowflake.ID, roles, existingRoles []snowflake.ID) error {
	botUser, err := restClient.GetCurrentUser("")
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Syncing roles for guild '%s'", guildID), slog.Any("roles", roles))

	for _, roleID := range roles {
		if slices.Contains(existingRoles, roleID) {
			slog.Info(fmt.Sprintf("Role '%s' already assigned to bot in guild '%s'. Skipping.", roleID.String(), guildID.String()))
			continue
		}

		r, err := restClient.GetRole(guildID, roleID)
		if err != nil {
			slog.Error("Failed to get role", slog.Any("err", err))
			continue
		}

		err = restClient.AddMemberRole(guildID, botUser.ID, r.ID)

		if err != nil {
			slog.Error(fmt.Sprintf("Failed to assign role '%s' to bot:", r.Name), slog.Any("err", err))
		} else {
			slog.Info(fmt.Sprintf("Successfully assigned role '%s' to bot in guild '%s'", r.Name, guildID.String()))
		}
	}

	return nil
}
