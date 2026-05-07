package suspiciousinterview

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/core"
	"github.com/bil0u/galaxy-os/internal/i18n"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
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

// --- Interview execution logic ---

var suspiciousWelcomeMessage = i18n.Text{
	discord.LocaleEnglishUS: "Hello %s!\n\nBefore you join our ship, we need to ask you a few questions to determine your role. You will be asked a series of questions, please answer them truthfully.\n\nAre you ready?",
	discord.LocaleFrench:    "Bonjour %s!\n\nAvant de rejoindre notre vaisseau, nous devons te poser quelques questions pour déterminer votre rôle. Tu seras invité à répondre à une série de questions, merci de répondre honnêtement.\n\nEs-tu prêt?",
}

func (cfg suspiciousInterviewConfig) executeInterview(restClient rest.Rest, member discord.Member, l discord.Locale) error {

	userDMChannel, err := restClient.CreateDMChannel(member.User.ID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel for user '%s' of guild '%s': %w", member.EffectiveName(), member.GuildID.String(), err)
	}

	welcomeMsgCreate := discord.NewMessageCreate().WithContent(suspiciousWelcomeMessage[l])

	restClient.CreateMessage(userDMChannel.ID(), welcomeMsgCreate)

	for _, question := range cfg.Questions {

		slog.Info(fmt.Sprintf("Asking question '%s'", question.Question))

		// Ask the question
		// err := restClient.CreateMessage(userDMChannel.ID, discord.NewMessageCreate().WithContent(question.Question))
		// if err != nil {
		// 	return fmt.Errorf("failed to send question: %w", err)
		// }

		// Wait for the user to answer

	}
	return nil
}
