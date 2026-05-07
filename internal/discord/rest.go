package discord

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// RestAdapter wraps disgo's rest.Rest and satisfies platform.RestClient.
// Context is forwarded to disgo via rest.WithCtx.
type RestAdapter struct {
	rest rest.Rest
}

// NewRestAdapter returns a RestAdapter wrapping the given disgo REST client.
func NewRestAdapter(r rest.Rest) *RestAdapter {
	return &RestAdapter{rest: r}
}

func (a *RestAdapter) withCtx(ctx context.Context, extra []rest.RequestOpt) []rest.RequestOpt {
	return append([]rest.RequestOpt{rest.WithCtx(ctx)}, extra...)
}

func (a *RestAdapter) AddMemberRole(ctx context.Context, guildID, userID, roleID snowflake.ID, opts ...rest.RequestOpt) error {
	return a.rest.AddMemberRole(guildID, userID, roleID, a.withCtx(ctx, opts)...)
}

func (a *RestAdapter) RemoveMemberRole(ctx context.Context, guildID, userID, roleID snowflake.ID, opts ...rest.RequestOpt) error {
	return a.rest.RemoveMemberRole(guildID, userID, roleID, a.withCtx(ctx, opts)...)
}

func (a *RestAdapter) CreateMessage(ctx context.Context, channelID snowflake.ID, create discord.MessageCreate, opts ...rest.RequestOpt) (*discord.Message, error) {
	return a.rest.CreateMessage(channelID, create, a.withCtx(ctx, opts)...)
}

func (a *RestAdapter) UpdateMessage(ctx context.Context, channelID, messageID snowflake.ID, update discord.MessageUpdate, opts ...rest.RequestOpt) (*discord.Message, error) {
	return a.rest.UpdateMessage(channelID, messageID, update, a.withCtx(ctx, opts)...)
}

func (a *RestAdapter) CreateDMChannel(ctx context.Context, userID snowflake.ID, opts ...rest.RequestOpt) (*discord.DMChannel, error) {
	return a.rest.CreateDMChannel(userID, a.withCtx(ctx, opts)...)
}

func (a *RestAdapter) GetMember(ctx context.Context, guildID, userID snowflake.ID, opts ...rest.RequestOpt) (*discord.Member, error) {
	return a.rest.GetMember(guildID, userID, a.withCtx(ctx, opts)...)
}

func (a *RestAdapter) GetRole(ctx context.Context, guildID, roleID snowflake.ID, opts ...rest.RequestOpt) (*discord.Role, error) {
	return a.rest.GetRole(guildID, roleID, a.withCtx(ctx, opts)...)
}
