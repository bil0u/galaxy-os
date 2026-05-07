package suspiciousinterview

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/core"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

var Feature = &suspiciousInterview{}

type suspiciousInterview struct{}

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
	guildID := deps.GuildID
	logger := deps.Logger
	configs := deps.Configs
	gw := deps.Gateway

	hasSuspiciousRole := func(member *events.GenericGuildMember) bool {
		cfg, err := core.ResolveGuild[suspiciousInterviewConfig](configs, guildID)
		if err != nil || !cfg.Enabled {
			return false
		}
		for _, roleID := range member.Member.RoleIDs {
			for _, detectID := range cfg.DetectRoles {
				if roleID == detectID {
					return true
				}
			}
		}
		return false
	}

	gw.AddEventListeners(disbot.NewListenerFunc(func(e *events.GuildMemberJoin) {
		if e.GuildID != guildID {
			return
		}
		if !hasSuspiciousRole(e.GenericGuildMember) {
			return
		}
		logger.Info("suspicious member joined", slog.String("user", e.Member.User.Username))
		// TODO: implement interview execution flow
		// 1. Open DM channel with the member
		// 2. Send localized welcome message
		// 3. Walk through each question, wait for reply
		// 4. Evaluate answers, assign IfSuccess or IfFailure role
		// 5. If ClearAfterInterview, remove DetectRoles
	}))

	gw.AddEventListeners(disbot.NewListenerFunc(func(e *events.GuildMemberUpdate) {
		if e.GuildID != guildID {
			return
		}
		if !hasSuspiciousRole(e.GenericGuildMember) {
			return
		}
		logger.Info("member gained suspicious role", slog.String("user", e.Member.User.Username))
		// TODO: same interview flow as GuildMemberJoin above
	}))

	return nil
}

func (f *suspiciousInterview) Start(_ context.Context) error { return nil }
func (f *suspiciousInterview) Stop(_ context.Context) error  { return nil }
