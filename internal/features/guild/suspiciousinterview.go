package guild_features

import (
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/locale"
	"github.com/bil0u/galaxy-os/internal/services"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

type SuspiciousInterviewConfig struct {
	Enabled             bool
	DetectRoles         []snowflake.ID
	Questions           Interview
	IfSuccess           snowflake.ID
	IfFailure           snowflake.ID
	ClearAfterInterview bool
}

func (f SuspiciousInterviewConfig) Validate() error {
	var errs []error
	if len(f.DetectRoles) == 0 {
		errs = append(errs, fmt.Errorf("detect roles are required"))
	}
	if f.IfSuccess == 0 {
		errs = append(errs, fmt.Errorf("if_success role is required"))
	}
	if f.IfFailure == 0 {
		errs = append(errs, fmt.Errorf("if_failure role is required"))
	}
	if len(f.Questions) == 0 {
		errs = append(errs, fmt.Errorf("questions are required"))
	}

	_, err := f.Questions.IsValid()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

var SuspiciousInterviewFeature = features.New[SuspiciousInterviewConfig](
	setupSuspiciousInterviewFeature,
	features.WithType(features.GuildFeature),
	features.WithName(locale.Text{
		discord.LocaleEnglishUS: "Suspicious Role Interview",
		discord.LocaleFrench:    "Entretien des rôles suspects",
	}),
	features.WithDescription(locale.Text{
		discord.LocaleEnglishUS: "Interview users with suspicious roles with a set of questions, and assign them a role based on their answers",
		discord.LocaleFrench:    "Interviewer les utilisateurs avec des rôles suspects, et leur attribuer un rôle en fonction de leurs réponses",
	}),
)

func setupSuspiciousInterviewFeature(deps features.SetupDeps) error {
	mainLogic := func(member discord.Member) {
		cfg, err := features.GetConfig[SuspiciousInterviewConfig](member.GuildID)
		if err != nil || !cfg.Enabled {
			slog.Warn(fmt.Sprintf("Feature 'SuspiciousInterview' is disabled for guild '%s'", member.GuildID))
			return
		}

		err = cfg.Validate()
		if err != nil {
			slog.Error(fmt.Sprintf("feature is not properly configured for guild '%s'", member.GuildID), slog.Any("err", err))
			return
		}

		restClient := services.RestClient()

		preferedLocale := discord.LocaleEnglishUS

		guild, err := restClient.GetGuild(member.GuildID, false)
		if err == nil {
			preferedLocale = discord.Locale(guild.PreferredLocale)
		}

		for _, roleID := range member.RoleIDs {
			for _, susRole := range cfg.DetectRoles {
				if roleID == susRole {
					slog.Info("User has a suspicious role", slog.Any("userID", member.User.ID), slog.Any("roleID", roleID))

					slog.Info(fmt.Sprintf("Running interview for user '%s'", member.EffectiveName()))
					err := cfg.ExecuteInterview(restClient, member, preferedLocale)
					if err != nil {
						slog.Error("failed to make interview", slog.Any("err", err))
						return
					}
				}
			}
		}
	}

	deps.Client.AddEventListeners(
		disbot.NewListenerFunc(func(event *events.GuildMemberUpdate) {
			mainLogic(event.Member)
		}),
		disbot.NewListenerFunc(func(event *events.GuildMemberJoin) {
			mainLogic(event.Member)
		}),
	)

	return nil
}

var welcomeMessage = locale.Text{
	discord.LocaleEnglishUS: "Hello %s!\n\nBefore you join our ship, we need to ask you a few questions to determine your role. You will be asked a series of questions, please answer them truthfully.\n\nAre you ready?",
	discord.LocaleFrench:    "Bonjour %s!\n\nAvant de rejoindre notre vaisseau, nous devons te poser quelques questions pour déterminer votre rôle. Tu seras invité à répondre à une série de questions, merci de répondre honnêtement.\n\nEs-tu prêt?",
}

func (f SuspiciousInterviewConfig) ExecuteInterview(restClient rest.Rest, member discord.Member, l discord.Locale) error {

	userDMChannel, err := restClient.CreateDMChannel(member.User.ID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel for user '%s' of guild '%s': %w", member.EffectiveName(), member.GuildID.String(), err)
	}

	welcomeMsgCreate := discord.NewMessageCreateBuilder().SetContent(welcomeMessage[l]).Build()

	restClient.CreateMessage(userDMChannel.ID(), welcomeMsgCreate)

	for _, question := range f.Questions {

		slog.Info(fmt.Sprintf("Asking question '%s'", question.Question))

		// Ask the question
		// err := restClient.CreateMessage(userDMChannel.ID, discord.NewMessageCreateBuilder().SetContent(question.Question).Build())
		// if err != nil {
		// 	return fmt.Errorf("failed to send question: %w", err)
		// }

		// Wait for the user to answer

	}
	return nil
}
