package selfassign

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/bil0u/galaxy-os/internal/core"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// SelfAssign assigns configured roles to the bot on startup.
var Feature = &selfAssign{}

type selfAssign struct {
	logger  *slog.Logger
	configs core.ConfigProvider
	guildID snowflake.ID
}

type selfAssignConfig struct {
	Enabled         bool
	ClearRolesFirst bool
	Roles           []snowflake.ID
}

func (c selfAssignConfig) Validate() error {
	var errs []error
	if len(c.Roles) == 0 {
		errs = append(errs, fmt.Errorf("roles are required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (f *selfAssign) Name() string            { return "self_assign" }
func (f *selfAssign) Scope() core.Scope       { return core.GuildScope }
func (f *selfAssign) Needs() []core.ServiceID { return nil }

func (f *selfAssign) Setup(deps core.Deps) error {
	f.logger = deps.Logger
	f.configs = deps.Configs
	f.guildID = deps.GuildID

	// TODO: self-assign requires gateway client access (bot.NewListenerFunc for events.Ready)
	// and REST methods not on core.RestClient (GetCurrentUser, GetRole, AddMemberRole,
	// RemoveMemberRole, GetMember). The event listener and guild iteration loop are removed;
	// the framework calls Setup per guild. Wire when gateway/client access is added to Deps.

	return nil
}

func (f *selfAssign) Start(ctx context.Context) error { return nil }
func (f *selfAssign) Stop(ctx context.Context) error  { return nil }

// --- Utility functions for role management ---

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
