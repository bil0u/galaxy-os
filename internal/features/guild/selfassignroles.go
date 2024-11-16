package guild_features

import (
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/services"
	"github.com/bil0u/galaxy-os/internal/utils"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

var SelfAssignRolesFeature = features.New[SelfAssignRolesConfig](
	setupSelfAssignRolesFeature,
	features.WithName(utils.LocalizedString{
		discord.LocaleEnglishUS: "Self assign roles",
		discord.LocaleFrench:    "Auto-attribution de rôles",
	}),
	features.WithDescription(utils.LocalizedString{
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

func setupSelfAssignRolesFeature() error {

	client := services.GetClient()

	(*client).AddEventListeners(disbot.NewListenerFunc(selfAssignRoles))
	return nil
}

// selfAssignRoles returns an event listener that assigns roles to the bot for each guild
func selfAssignRoles(_ *events.Ready) {
	// Looping through each guild to assign roles
	for _, guildID := range config.Guilds.IDs(config.Global.Development) {

		// Getting feature config for guild
		cfg, _ := features.GetConfig[SelfAssignRolesConfig](guildID)

		if !cfg.Enabled {
			slog.Warn(fmt.Sprintf("Feature 'SelfAssignRoles' is disabled for guild '%s'", guildID))
			continue
		}

		existingRoles, err := getAssignedRoles(guildID)
		if err != nil {
			slog.Error("Failed to get assigned roles", slog.Any("err", err))
			continue
		}

		// Removing roles if set in config
		if cfg.ClearRolesFirst {
			slog.Info(fmt.Sprintf("Clearing roles for guild '%s'", guildID))
			if err := removeBotRoles(guildID, existingRoles); err != nil {
				slog.Error("Failed to remove roles", slog.Any("err", err))
			}
		}

		// Then assign roles to the bot
		if err := assignBotRoles(guildID, cfg.Roles, existingRoles); err != nil {
			slog.Error("Failed to assign roles to bot", slog.Any("err", err))
		}
	}
}

// getAssignedRoles returns the roles assigned to the bot in the provided guild
func getAssignedRoles(guildID snowflake.ID) ([]snowflake.ID, error) {
	restClient := services.GetRestClient()

	botUser, err := restClient.GetMember(guildID, config.Bot.ApplicationID)
	if err != nil {
		return nil, err
	}

	return botUser.RoleIDs, nil
}

// RemoveBotRoles removes the provided roles from the bot in the provided guild
func removeBotRoles(guildID snowflake.ID, roles []snowflake.ID) error {

	restClient := services.GetRestClient()

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
func assignBotRoles(guildID snowflake.ID, roles, existingRoles []snowflake.ID) error {

	restClient := services.GetRestClient()
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
