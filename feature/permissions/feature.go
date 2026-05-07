package permissions

import (
	"context"

	"github.com/bil0u/galaxy-os/internal/core"
)

// Permissions logs the permissions of the bot for each guild.
var Feature = &permissions{}

type permissions struct{}

type permissionsConfig struct{}

func (c permissionsConfig) Validate() error { return nil }

func (f *permissions) Name() string            { return "permissions" }
func (f *permissions) Scope() core.Scope       { return core.BotScope }
func (f *permissions) Needs() []core.ServiceID { return nil }

func (f *permissions) Setup(deps core.Deps) error {
	return nil
}

func (f *permissions) Start(ctx context.Context) error { return nil }
func (f *permissions) Stop(ctx context.Context) error  { return nil }
