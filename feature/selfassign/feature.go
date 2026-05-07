package selfassign

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/bil0u/galaxy-os/internal/core"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

var Feature = &selfAssign{}

type selfAssign struct{}

type selfAssignConfig struct {
	Enabled         bool
	ClearRolesFirst bool
	Roles           []snowflake.ID
}

func (c selfAssignConfig) Validate() error {
	if len(c.Roles) == 0 {
		return fmt.Errorf("roles are required")
	}
	return nil
}

func (f *selfAssign) Name() string            { return "self_assign" }
func (f *selfAssign) Scope() core.Scope       { return core.GuildScope }
func (f *selfAssign) Needs() []core.ServiceID { return nil }

func (f *selfAssign) Setup(deps core.Deps) error {
	guildID := deps.GuildID
	logger := deps.Logger
	configs := deps.Configs
	gw := deps.Gateway
	restClient := deps.Rest

	gw.AddEventListeners(disbot.NewListenerFunc(func(_ *events.Ready) {
		ctx := context.Background()

		cfg, err := core.ResolveGuild[selfAssignConfig](configs, guildID)
		if err != nil {
			logger.Error("resolving config", slog.Any("error", err))
			return
		}
		if !cfg.Enabled {
			return
		}
		if err := cfg.Validate(); err != nil {
			logger.Error("invalid config", slog.Any("error", err))
			return
		}

		botID := gw.ID()

		if cfg.ClearRolesFirst {
			clearRoles(ctx, logger, restClient, guildID, botID)
		}
		assignRoles(ctx, logger, restClient, guildID, botID, cfg.Roles)
	}))

	return nil
}

func (f *selfAssign) Start(_ context.Context) error { return nil }
func (f *selfAssign) Stop(_ context.Context) error  { return nil }

func clearRoles(ctx context.Context, logger *slog.Logger, restClient core.RestClient, guildID, botID snowflake.ID) {
	member, err := restClient.GetMember(ctx, guildID, botID)
	if err != nil {
		logger.Error("getting bot member", slog.Any("error", err))
		return
	}

	for _, roleID := range member.RoleIDs {
		role, err := restClient.GetRole(ctx, guildID, roleID)
		if err != nil {
			logger.Error("getting role", slog.Any("error", err))
			continue
		}
		if role.Managed {
			continue
		}
		if err := restClient.RemoveMemberRole(ctx, guildID, botID, roleID); err != nil {
			logger.Error("removing role", slog.String("role", role.Name), slog.Any("error", err))
			continue
		}
		logger.Info("removed role", slog.String("role", role.Name))
	}
}

func assignRoles(ctx context.Context, logger *slog.Logger, restClient core.RestClient, guildID, botID snowflake.ID, roles []snowflake.ID) {
	member, err := restClient.GetMember(ctx, guildID, botID)
	if err != nil {
		logger.Error("getting bot member", slog.Any("error", err))
		return
	}

	for _, roleID := range roles {
		if slices.Contains(member.RoleIDs, roleID) {
			continue
		}
		if err := restClient.AddMemberRole(ctx, guildID, botID, roleID); err != nil {
			logger.Error("assigning role", slog.String("role", roleID.String()), slog.Any("error", err))
			continue
		}
		logger.Info("assigned role", slog.String("role", roleID.String()))
	}
}
