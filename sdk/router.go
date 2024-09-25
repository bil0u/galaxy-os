package sdk

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

type RouterInitializer func(*Bot, *handler.Mux) []discord.ApplicationCommandCreate
