package utils

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

func RoleFromAppCommandRole(perm discord.ApplicationCommandPermissionRole, guildRoles []discord.Role) (discord.Role, error) {
	for _, role := range guildRoles {
		if role.ID == perm.RoleID {
			return role, nil
		}
	}
	return discord.Role{}, fmt.Errorf("role with ID %s not found", perm.RoleID.String())
}

func MemberPermissionsFromRoles(guildMember discord.Member, guildRoles []discord.Role) (discord.Permissions, error) {
	// Fetch the user's roles
	for _, roleID := range guildMember.RoleIDs {
		// Fetch the role permissions
		for _, role := range guildRoles {
			if role.ID == roleID {
				return role.Permissions, nil
			}
		}
	}
	return 0, fmt.Errorf("user %s not found", guildMember.User.ID.String())
}

func GuildChannelFromAppCommandChannel(perm discord.ApplicationCommandPermissionChannel, guildChannels []discord.GuildChannel) (discord.GuildChannel, error) {

	for _, channel := range guildChannels {
		if channel.ID() == perm.ChannelID {
			return channel, nil
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
				acpRole := perm.(discord.ApplicationCommandPermissionRole)
				role, err := RoleFromAppCommandRole(acpRole, guildRoles)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch role permissions for guild %s: %v", guildID.String(), err)
				}
				rolePermissions = role.Permissions
			case discord.ApplicationCommandPermissionTypeUser:
				acpUser := perm.(discord.ApplicationCommandPermissionUser)
				// Fetch the user permissions
				guildMember, err := guilds.GetMember(guildID, acpUser.UserID)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch guild member for guild %s: %v", guildID.String(), err)
				}
				userPermissions, err = MemberPermissionsFromRoles(*guildMember, guildRoles)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch user permissions for guild %s: %v", guildID.String(), err)
				}

			case discord.ApplicationCommandPermissionTypeChannel:
				// Fetch the channel permissions
				acpChannel := perm.(discord.ApplicationCommandPermissionChannel)

				guildChannels, err := guilds.GetGuildChannels(guildID)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch guild channels for guild %s: %v", guildID.String(), err)
				}
				channel, err := GuildChannelFromAppCommandChannel(acpChannel, guildChannels)
				// No errors means the channel was found
				if err == nil {
					channelPermissions[acpChannel.ChannelID] = channel.PermissionOverwrites()
				}
			default:
				slog.Warn("Unknown permission type", slog.Any("type", perm.Type()))
			}
		}

	}
	return rolePermissions, userPermissions, channelPermissions, nil
}

func LogPermissions(client bot.Client, guildsID []snowflake.ID) {
	// Checking permissions for each Guild
	for _, guildID := range guildsID {
		rolePerms, userPerms, channelOverwrites, err := CheckBotPermissions(client, guildID)
		if err != nil {
			slog.Error("Error checking bot permissions:", slog.Any("err", err))
		}
		slog.Info(fmt.Sprintf("[BOT PERMISSIONS - GUILD '%s']:", guildID.String()))
		slog.Info(fmt.Sprintf("- Role permissions:\n%v", rolePerms.String()))
		slog.Info(fmt.Sprintf("- User permissions:\n%v", userPerms.String()))
		slog.Info("- Channel overwrites:")
		for channelID, overwrites := range channelOverwrites {
			slog.Info(fmt.Sprintf("  - <Channel '%s'>:", channelID.String()))
			for _, overwrite := range overwrites {
				switch overwrite.Type() {
				case discord.PermissionOverwriteTypeRole:
					roleOverwrite := overwrite.(discord.RolePermissionOverwrite)
					slog.Info(fmt.Sprintf("    > Role '%s':", roleOverwrite.RoleID.String()))
					slog.Info(fmt.Sprintf("      - Allow: %v", roleOverwrite.Allow.String()))
					slog.Info(fmt.Sprintf("      - Deny: %v", roleOverwrite.Deny.String()))
				case discord.PermissionOverwriteTypeMember:
					memberOverwrite := overwrite.(discord.MemberPermissionOverwrite)
					slog.Info(fmt.Sprintf("    > User '%s':", memberOverwrite.UserID.String()))
					slog.Info(fmt.Sprintf("      - Allow: %v", memberOverwrite.Allow.String()))
					slog.Info(fmt.Sprintf("      - Deny: %v", memberOverwrite.Deny.String()))
				}
			}
		}
	}
}
