package cmd

import (
	"github.com/bil0u/galaxy-os/feature/dailymessage"
	"github.com/bil0u/galaxy-os/feature/info"
	"github.com/bil0u/galaxy-os/feature/permissions"
	"github.com/bil0u/galaxy-os/feature/presence"
	"github.com/bil0u/galaxy-os/feature/selfassign"
	"github.com/bil0u/galaxy-os/feature/suspiciousinterview"
	"github.com/bil0u/galaxy-os/feature/testcmd"
	"github.com/bil0u/galaxy-os/internal/core"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
)

type botDef struct {
	features   []core.Feature
	cacheFlags cache.Flags
	intents    gateway.Intents
}

var bots = map[string]botDef{
	"hue": {
		features: []core.Feature{
			info.Feature,
			permissions.Feature,
			testcmd.Feature,
			presence.Feature,
			selfassign.Feature,
			dailymessage.Feature,
			suspiciousinterview.Feature,
		},
		cacheFlags: cache.FlagGuilds | cache.FlagMembers | cache.FlagRoles,
		intents:    gateway.IntentGuilds | gateway.IntentGuildMembers,
	},
	"kevin": {
		features: []core.Feature{
			info.Feature,
			permissions.Feature,
			testcmd.Feature,
			presence.Feature,
			selfassign.Feature,
			dailymessage.Feature,
		},
		cacheFlags: cache.FlagGuilds | cache.FlagRoles,
		intents:    gateway.IntentGuilds,
	},
}
