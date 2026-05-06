package feature

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/platform"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// Permissions logs the permissions of the bot for each guild.
var Permissions = &permissions{}

type permissions struct{}

type permissionsConfig struct{}

func (c permissionsConfig) Validate() error { return nil }

func (f *permissions) Name() string               { return "permissions" }
func (f *permissions) Scope() platform.Scope       { return platform.BotScope }
func (f *permissions) Needs() []platform.ServiceID { return nil }

func (f *permissions) Setup(deps platform.Deps) error {
	return nil
}

func (f *permissions) Start(ctx context.Context) error { return nil }
func (f *permissions) Stop(ctx context.Context) error  { return nil }

// --- Utility functions used by cmd/bot.go ---

func roleFromAppCommandRole(perm discord.ApplicationCommandPermissionRole, guildRoles []discord.Role) (discord.Role, error) {
	for _, role := range guildRoles {
		if role.ID == perm.RoleID {
			return role, nil
		}
	}
	return discord.Role{}, fmt.Errorf("role with ID %s not found", perm.RoleID.String())
}

func memberPermissionsFromRoles(guildMember discord.Member, guildRoles []discord.Role) discord.Permissions {
	var combined discord.Permissions
	for _, roleID := range guildMember.RoleIDs {
		for _, role := range guildRoles {
			if role.ID == roleID {
				combined |= role.Permissions
				break
			}
		}
	}
	return combined
}

func guildChannelFromAppCommandChannel(perm discord.ApplicationCommandPermissionChannel, guildChannels []discord.GuildChannel) (discord.GuildChannel, error) {

	for _, channel := range guildChannels {
		if channel.ID() == perm.ChannelID {
			return channel, nil
		}
	}
	return nil, fmt.Errorf("guild channel %s not found", perm.ChannelID.String())
}

func checkBotPermissions(restClient rest.Rest, applicationID snowflake.ID, guildID snowflake.ID) (discord.Permissions, discord.Permissions, map[snowflake.ID]discord.PermissionOverwrites, error) {
	guildCommandsPermissions, err := restClient.GetGuildCommandsPermissions(applicationID, guildID)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to fetch bot info for guild %s: %v", guildID.String(), err)
	}

	guildRoles, err := restClient.GetRoles(guildID)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to fetch guildRoles for guild %s: %v", guildID.String(), err)
	}

	var rolePermissions, userPermissions discord.Permissions
	channelPermissions := map[snowflake.ID]discord.PermissionOverwrites{}

	for _, permGroup := range guildCommandsPermissions {
		for _, perm := range permGroup.Permissions {
			switch perm.Type() {
			case discord.ApplicationCommandPermissionTypeRole:
				acpRole := perm.(discord.ApplicationCommandPermissionRole)
				role, err := roleFromAppCommandRole(acpRole, guildRoles)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch role permissions for guild %s: %v", guildID.String(), err)
				}
				rolePermissions |= role.Permissions
			case discord.ApplicationCommandPermissionTypeUser:
				acpUser := perm.(discord.ApplicationCommandPermissionUser)
				guildMember, err := restClient.GetMember(guildID, acpUser.UserID)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch guild member for guild %s: %v", guildID.String(), err)
				}
				userPermissions = memberPermissionsFromRoles(*guildMember, guildRoles)

			case discord.ApplicationCommandPermissionTypeChannel:
				acpChannel := perm.(discord.ApplicationCommandPermissionChannel)

				guildChannels, err := restClient.GetGuildChannels(guildID)
				if err != nil {
					return 0, 0, nil, fmt.Errorf("failed to fetch guild channels for guild %s: %v", guildID.String(), err)
				}
				channel, err := guildChannelFromAppCommandChannel(acpChannel, guildChannels)
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

// LogPermissions logs the bot's permissions for each guild.
// Called from cmd/bot.go — uses raw rest.Rest and old config types.
func LogPermissions(restClient rest.Rest, botCfg *config.Bot, guilds *config.GuildMap, devGuildsOnly bool) {
	for _, guildID := range guilds.IDs(devGuildsOnly) {
		rolePerms, userPerms, channelOverwrites, err := checkBotPermissions(restClient, botCfg.ApplicationID, guildID)
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
