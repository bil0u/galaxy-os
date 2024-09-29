package sdk

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// ------------
// LOCALIZATION
// ------------

type LocalizedString map[discord.Locale]string

func (str LocalizedString) String(locale discord.Locale) string {
	if val, ok := str[locale]; ok {
		return val
	}
	if val, ok := str[discord.LocaleEnglishUS]; ok {
		return val
	}
	slog.Warn("No localization found for locale", slog.Any("locale", locale))
	return fmt.Sprintf("No localization found for locale %s", locale)
}

// -----------
// PERMISSIONS
// -----------

// CheckBotPermissions checks the bot's permissions in a specific guild
func CheckBotPermissions(client bot.Client, guildID snowflake.ID) (discord.Permissions, error) {
	guilds := client.Rest()
	// Fetch the bot's member information in the guild
	member, err := guilds.GetMember(guildID, client.ID())
	if err != nil {
		slog.Error("Failed to fetch bot member info for guild %s: %v", guildID.String(), err)
		return 0, err
	}
	// Fetch the guild roles
	roles, err := guilds.GetRoles(guildID)
	if err != nil {
		slog.Error("Failed to fetch roles for guild %s: %v", guildID.String(), err)
		return 0, err
	}
	// Calculate the bot's permissions
	var botPermissions discord.Permissions
	for _, roleID := range member.RoleIDs {
		for _, role := range roles {
			if role.ID == roleID {
				botPermissions |= role.Permissions
			}
		}
	}
	// Check if the bot has administrator permissions
	if botPermissions.Has(discord.PermissionAdministrator) {
		botPermissions = discord.PermissionsAll
	}
	return botPermissions, nil
}

func LogPermissions(client bot.Client, cfg BotConfig) {
	// Create the bot client if not provided
	if client == nil {
		dummyClient, err := NewBotClient(cfg.Token, *NewBotParts())
		if err != nil {
			slog.Error("Error creating client: %v", slog.Any("err", err))
			return
		}
		client = *dummyClient
	}
	// Checking permissions for each Guild
	for _, guildID := range cfg.Guilds {
		permissions, err := CheckBotPermissions(client, guildID)
		if err != nil {
			slog.Error("Error checking bot permissions: %v", slog.Any("err", err))
		}

		slog.Info("Bot permissions in guild %s: %v\n", guildID.String(), permissions)
	}
}

// -----
// ROLES
// -----

// AssignRoleToBot assigns a specific role to the bot in a guild
func AssignRolesToBot(client bot.Client, guildID snowflake.ID, roleIDs []snowflake.ID) error {
	guilds := client.Rest()

	// Add the role to the bot
	for _, roleID := range roleIDs {
		err := guilds.AddMemberRole(guildID, client.ApplicationID(), roleID)
		if err != nil {
			slog.Error("Failed to assign role %s to bot in guild %s: %v", roleID.String(), guildID.String(), err)
			return err
		}
		slog.Info("Successfully assigned role %s to bot in guild %s\n", roleID.String(), guildID.String())
	}
	return nil
}

// FetchRoles retrieves all roles from the bot's dev guilds and returns them as a slice of discord.Role
func FetchRoles(client bot.Client, Guilds []snowflake.ID) ([]discord.Role, error) {
	guilds := client.Rest()
	var roles []discord.Role
	for _, guildID := range Guilds {
		slog.Info(fmt.Sprintf("Fetching roles for guild %s", guildID))
		guildRoles, err := guilds.GetRoles(guildID)
		if err != nil {
			return nil, err
		}
		roles = append(roles, guildRoles...)
	}
	return roles, nil
}
