package contracts

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// RestClient is the testability seam for Discord REST operations.
// Wraps methods from disgo's rest.Rest that features actually use.
type RestClient interface {
	AddMemberRole(ctx context.Context, guildID, userID, roleID snowflake.ID, opts ...rest.RequestOpt) error
	RemoveMemberRole(ctx context.Context, guildID, userID, roleID snowflake.ID, opts ...rest.RequestOpt) error
	CreateMessage(ctx context.Context, channelID snowflake.ID, create discord.MessageCreate, opts ...rest.RequestOpt) (*discord.Message, error)
	UpdateMessage(ctx context.Context, channelID, messageID snowflake.ID, update discord.MessageUpdate, opts ...rest.RequestOpt) (*discord.Message, error)
	CreateDMChannel(ctx context.Context, userID snowflake.ID, opts ...rest.RequestOpt) (*discord.DMChannel, error)
	GetMember(ctx context.Context, guildID, userID snowflake.ID, opts ...rest.RequestOpt) (*discord.Member, error)
}
