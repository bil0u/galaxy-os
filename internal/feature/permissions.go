package feature

import (
	"context"

	"github.com/bil0u/galaxy-os/internal/platform"
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
