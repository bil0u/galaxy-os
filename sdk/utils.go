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

func getRolePermissions(perm discord.ApplicationCommandPermissionRole, guildRoles []discord.Role) (discord.Permissions, error) {
	for _, role := range guildRoles {
		if role.ID == perm.RoleID {
			return role.Permissions, nil
		}
	}
	return 0, fmt.Errorf("role %s not found", perm.RoleID.String())
}

func getUserPermissions(perm discord.ApplicationCommandPermissionUser, guildMember discord.Member, guildRoles []discord.Role) (discord.Permissions, error) {
	// Fetch the user's roles
	for _, roleID := range guildMember.RoleIDs {
		// Fetch the role permissions
		for _, role := range guildRoles {
			if role.ID == roleID {
				return role.Permissions, nil
			}
		}
	}
	return 0, fmt.Errorf("user %s not found", perm.UserID.String())
}

func getGuildChannelPermissions(perm discord.ApplicationCommandPermissionChannel, guildChannels []discord.GuildChannel) (discord.PermissionOverwrites, error) {

	for _, channel := range guildChannels {
		if channel.ID() == perm.ChannelID {
			return channel.PermissionOverwrites(), nil
		}
	}
	return nil, fmt.Errorf("guild channel %s not found", perm.ChannelID.String())
}

// CheckBotPermissions checks the bot's permissions in a specific guild
func CheckBotPermissions(client bot.Client, guildID snowflake.ID) (discord.Permissions, discord.Permissions, map[snowflake.ID]discord.PermissionOverwrites, error) {
	guilds := client.Rest()

	appId := client.ApplicationID()

	// Fetch the bot's information in the guild
	guildCommandsPermissions, err := guilds.GetGuildCommandsPermissions(appId, guildID)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to fetch bot info for guild %s: %v", guildID.String(), err)
	}

	// Fetch the guild roles
	guildRoles, err := guilds.GetRoles(guildID)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to fetch guildRoles for guild %s: %v", guildID.String(), err)
	}

	var rolePermissions, userPermissions discord.Permissions
	channelPermissions := map[snowflake.ID]discord.PermissionOverwrites{}

	// Convert the permissions to a single value
	for _, permGroup := range guildCommandsPermissions {
		for _, perm := range permGroup.Permissions {
			switch perm.Type() {
			case discord.ApplicationCommandPermissionTypeRole:
				rolePerm := perm.(discord.ApplicationCommandPermissionRole)
				rolePermissions, err = getRolePermissions(rolePerm, guildRoles)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch role permissions for guild %s: %v", guildID.String(), err)
				}
			case discord.ApplicationCommandPermissionTypeUser:
				userPerm := perm.(discord.ApplicationCommandPermissionUser)
				// Fetch the user permissions
				guildMember, err := guilds.GetMember(guildID, userPerm.UserID)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch guild member for guild %s: %v", guildID.String(), err)
				}
				userPermissions, err = getUserPermissions(userPerm, *guildMember, guildRoles)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch user permissions for guild %s: %v", guildID.String(), err)
				}

			case discord.ApplicationCommandPermissionTypeChannel:
				// Fetch the channel permissions
				channelPerm := perm.(discord.ApplicationCommandPermissionChannel)

				guildChannels, err := guilds.GetGuildChannels(guildID)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch guild channels for guild %s: %v", guildID.String(), err)
				}
				channelOverwrites, err := getGuildChannelPermissions(channelPerm, guildChannels)
				// No errors means the channel was found
				if err == nil {
					channelPermissions[channelPerm.ChannelID] = channelOverwrites
				}
			default:
				slog.Warn("Unknown permission type", slog.Any("type", perm.Type()))
			}
		}

	}
	return rolePermissions, userPermissions, channelPermissions, nil
}

func LogPermissions(client bot.Client, cfg BotConfig) {
	// Create the bot client if not provided
	if client == nil {
		dummyClient, err := NewBotClient(cfg.Token, *NewBotParts())
		if err != nil {
			slog.Error("Error creating client", slog.Any("err", err))
			return
		}
		client = *dummyClient
	}
	// Checking permissions for each Guild
	for _, guildID := range cfg.Guilds {
		rolePerms, userPerms, channelOverwrites, err := CheckBotPermissions(client, guildID)
		if err != nil {
			slog.Error("Error checking bot permissions:", slog.Any("err", err))
		}
		slog.Info(fmt.Sprintf("[BOT PERMISSIONS - GUILD '%s']:\n", guildID.String()))
		slog.Info(fmt.Sprintf("- Role permissions:\n%v\n", rolePerms.String()))
		slog.Info(fmt.Sprintf("- User permissions:\n%v\n", userPerms.String()))
		slog.Info("- Channel overwrites:\n")
		for channelID, overwrites := range channelOverwrites {
			slog.Info(fmt.Sprintf("  - <Channel '%s'>:\n", channelID.String()))
			for _, overwrite := range overwrites {
				switch overwrite.Type() {
				case discord.PermissionOverwriteTypeRole:
					roleOverwrite := overwrite.(discord.RolePermissionOverwrite)
					slog.Info(fmt.Sprintf("    > Role '%s':\n", roleOverwrite.RoleID.String()))
					slog.Info(fmt.Sprintf("      - Allow: %v\n", roleOverwrite.Allow.String()))
					slog.Info(fmt.Sprintf("      - Deny: %v\n", roleOverwrite.Deny.String()))
				case discord.PermissionOverwriteTypeMember:
					memberOverwrite := overwrite.(discord.MemberPermissionOverwrite)
					slog.Info(fmt.Sprintf("    > User '%s':\n", memberOverwrite.UserID.String()))
					slog.Info(fmt.Sprintf("      - Allow: %v\n", memberOverwrite.Allow.String()))
					slog.Info(fmt.Sprintf("      - Deny: %v\n", memberOverwrite.Deny.String()))
				}
			}
		}
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
			return fmt.Errorf("failed to assign role %s to bot in guild %s: %v", roleID.String(), guildID.String(), err)
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

// FetchMembers retrieves all members from the bot's dev guilds and returns them as a slice of discord.Member

func FetchMembers(client bot.Client, guildID snowflake.ID) ([]discord.Member, error) {
	guilds := client.Rest()
	// Fetch the guild users
	guildMembers := []discord.Member{}
	chunkSize := 100
	for i := 0; i < len(guildMembers); i += chunkSize {
		end := i + chunkSize
		if end > len(guildMembers) {
			end = len(guildMembers)
		}
		members, err := guilds.GetMembers(guildID, chunkSize, guildMembers[end].User.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guildMembers for guild %s: %v", guildID.String(), err)
		}
		guildMembers = append(guildMembers, members...)
	}
	return guildMembers, nil
}
