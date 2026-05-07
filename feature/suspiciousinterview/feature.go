package suspiciousinterview

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/core"
	"github.com/disgoorg/snowflake/v2"
)

// SuspiciousInterview interviews users with suspicious roles and assigns
// a role based on their answers.
var Feature = &suspiciousInterview{}

type suspiciousInterview struct {
	logger  *slog.Logger
	configs core.ConfigProvider
	guildID snowflake.ID
}

type suspiciousInterviewConfig struct {
	Enabled             bool
	DetectRoles         []snowflake.ID
	Questions           Interview
	IfSuccess           snowflake.ID
	IfFailure           snowflake.ID
	ClearAfterInterview bool
}

func (c suspiciousInterviewConfig) Validate() error {
	var errs []error
	if len(c.DetectRoles) == 0 {
		errs = append(errs, fmt.Errorf("detect roles are required"))
	}
	if c.IfSuccess == 0 {
		errs = append(errs, fmt.Errorf("if_success role is required"))
	}
	if c.IfFailure == 0 {
		errs = append(errs, fmt.Errorf("if_failure role is required"))
	}
	if len(c.Questions) == 0 {
		errs = append(errs, fmt.Errorf("questions are required"))
	}

	_, err := c.Questions.IsValid()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (f *suspiciousInterview) Name() string            { return "suspicious_interview" }
func (f *suspiciousInterview) Scope() core.Scope       { return core.GuildScope }
func (f *suspiciousInterview) Needs() []core.ServiceID { return nil }

func (f *suspiciousInterview) Setup(deps core.Deps) error {
	f.logger = deps.Logger
	f.configs = deps.Configs
	f.guildID = deps.GuildID

	// TODO: suspicious interview requires gateway client access
	// (bot.NewListenerFunc for events.GuildMemberUpdate and events.GuildMemberJoin)
	// and REST methods not on core.RestClient (GetGuild).
	// Wire when gateway/client access is added to Deps.

	return nil
}

func (f *suspiciousInterview) Start(ctx context.Context) error { return nil }
func (f *suspiciousInterview) Stop(ctx context.Context) error  { return nil }

// TODO: implement interview execution flow.
// When gateway client access is available in Deps:
// 1. Listen for GuildMemberJoin/GuildMemberUpdate events
// 2. Check if the member has any of the configured DetectRoles
// 3. Open a DM channel and send the welcome message (localized)
// 4. Walk through each question from config, send it, wait for the user's reply
// 5. Evaluate answers and assign IfSuccess or IfFailure role accordingly
// 6. If ClearAfterInterview is set, remove the original DetectRoles from the member
