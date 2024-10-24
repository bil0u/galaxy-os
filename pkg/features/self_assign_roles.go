package features

import (
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/pkg"
	"github.com/bil0u/galaxy-os/pkg/utils"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func init() {
	pkg.RegisterFeature[SelfAssignRolesFeature]("self_assign_roles")
}

type SelfAssignRolesFeature struct {
	Enabled         bool           `toml:"enabled"`
	Roles           []snowflake.ID `toml:"roles"`
	ClearRolesFirst bool           `toml:"clear_roles_first"`
}

func (f SelfAssignRolesFeature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Self assign roles",
		discord.LocaleFrench:    "Auto-attribution de rôles",
	}
}

func (f SelfAssignRolesFeature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "The bot will assign roles to itself automatically",
		discord.LocaleFrench:    "Le bot s'attribuera des rôles automatiquement",
	}
}

func (f SelfAssignRolesFeature) IsEnabled() bool {
	return f.Enabled
}

func (f SelfAssignRolesFeature) IsProperlyConfigured() error {
	var errs []error
	if len(f.Roles) == 0 {
		errs = append(errs, fmt.Errorf("roles are required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (f SelfAssignRolesFeature) Setup(bot *pkg.Bot) error {
	bot.Client.AddEventListeners(selfAssignRoles(bot))
	return nil
}

// selfAssignRoles returns an event listener that assigns roles to the bot for each guild
func selfAssignRoles(bot *pkg.Bot) disbot.EventListener {
	return disbot.NewListenerFunc(func(_ *events.Ready) {

		// Looping through each guild to assign roles
		for _, guildCfg := range bot.Config.Guilds {

			// Getting feature config for guild
			feature, _ := pkg.GetFeature[*SelfAssignRolesFeature](guildCfg.Features)

			if !feature.IsEnabled() {
				continue
			}

			existingRoles, err := getAssignedRoles(bot, guildCfg.ID)
			if err != nil {
				slog.Error("Failed to get assigned roles", slog.Any("err", err))
				continue
			}

			// Removing roles if set in config
			if feature.ClearRolesFirst {
				slog.Info(fmt.Sprintf("Clearing roles for guild '%s'", guildCfg.ID))
				if err := removeBotRoles(bot, guildCfg.ID, existingRoles); err != nil {
					slog.Error("Failed to remove roles", slog.Any("err", err))
				}
			}

			// Then assign roles to the bot
			if err := assignBotRoles(bot, guildCfg.ID, feature.Roles, existingRoles); err != nil {
				slog.Error("Failed to assign roles to bot", slog.Any("err", err))
			}
		}

	})
}

// getAssignedRoles returns the roles assigned to the bot in the provided guild
func getAssignedRoles(bot *pkg.Bot, guildID snowflake.ID) ([]snowflake.ID, error) {
	restClient := bot.Client.Rest()

	botUser, err := restClient.GetMember(guildID, bot.Client.ApplicationID())
	if err != nil {
		return nil, err
	}

	return botUser.RoleIDs, nil
}

// RemoveBotRoles removes the provided roles from the bot in the provided guild
func removeBotRoles(bot *pkg.Bot, guildID snowflake.ID, roles []snowflake.ID) error {

	restClient := bot.Client.Rest()

	// Getting user using the bot ID
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

// assignBotRoles assigns the provided roles to the bot in the provided guild
func assignBotRoles(bot *pkg.Bot, guildID snowflake.ID, roles, existingRoles []snowflake.ID) error {

	restClient := bot.Client.Rest()

	// Getting user using the bot ID
	botUser, err := restClient.GetCurrentUser("")
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Syncing roles for guild '%s'", guildID), slog.Any("roles", roles))

	for _, roleID := range roles {
		// Assign each role to the bot

		if utils.Contains(existingRoles, roleID) {
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
