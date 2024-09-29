package generators

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/bot"
)

var (
	All = []GeneratorHandler{
		GenerateRoleEnum,
	}
)

type GeneratorHandler func(client bot.Client, cfg sdk.BotConfig) error
