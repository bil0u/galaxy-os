package features

import (
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/pkg"
	"github.com/bil0u/galaxy-os/pkg/utils"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func init() {
	pkg.RegisterFeature[SuspiciousInterwiewFeature]("suspicious_interview")
}

type SuspiciousInterwiewFeature struct {
	Enabled             bool            `toml:"enabled"`
	DetectRoles         []snowflake.ID  `toml:"detect_roles"`
	Questions           utils.Interview `toml:"questions"`
	IfSuccess           snowflake.ID    `toml:"if_success"`
	IfFailure           snowflake.ID    `toml:"if_failure"`
	ClearAfterInterview bool            `toml:"clear_after_interview"`
}

func (f SuspiciousInterwiewFeature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Suspicious Role Interview",
		discord.LocaleFrench:    "Entretien des rôles suspects",
	}
}

func (f SuspiciousInterwiewFeature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Interview users with suspicious roles with a set of questions, and assign them a role based on their answers",
		discord.LocaleFrench:    "Interviewer les utilisateurs avec des rôles suspects, et leur attribuer un rôle en fonction de leurs réponses",
	}
}

func (f SuspiciousInterwiewFeature) IsEnabled() bool {
	return f.Enabled
}
func (f SuspiciousInterwiewFeature) IsProperlyConfigured() error {
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

	err := f.Questions.IsProperlyConfigured()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (f SuspiciousInterwiewFeature) Setup(bot *pkg.Bot) error {

	mainLogic := func(member discord.Member) {
		// Gettting the guild config from the bot
		guildConfig, err := bot.Config.GetGuildConfig(member.GuildID)
		if err != nil {
			slog.Error("failed to get guild config: %w", slog.Any("err", err))
			return
		}

		// Getting the feature definition from the guild config
		feature, err := pkg.GetFeature[*SuspiciousInterwiewFeature](guildConfig.Features)
		if err != nil {
			slog.Warn("feature not found in guild config", slog.Any("err", err))
			return
		}

		// Check if the feature is enabled
		if !feature.IsEnabled() {
			slog.Warn("feature is not enabled")
			return
		}

		// Check if the feature is properly configured
		err = feature.IsProperlyConfigured()
		if err != nil {
			slog.Error("feature is not properly configured", slog.Any("err", err))
			return
		}

		preferedLocale := discord.LocaleEnglishUS

		// Get the guild locale
		guild, err := bot.Client.Rest().GetGuild(member.GuildID, false)
		if err == nil {
			preferedLocale = discord.Locale(guild.PreferredLocale)
		}

		// Inspect each role of the user
		for _, roleID := range member.RoleIDs {

			// Check if the role is in the list of suspicious roles
			for _, susRole := range feature.DetectRoles {
				if roleID == susRole {
					slog.Info("User has a suspicious role", slog.Any("userID", member.User.ID), slog.Any("roleID", roleID))

					// Execute the interview
					err := feature.ExecuteInterview(bot, member, preferedLocale)
					if err != nil {
						slog.Error("failed to make interview", slog.Any("err", err))
						return
					}
				}

			}
		}
	}

	bot.Client.AddEventListeners(
		disbot.NewListenerFunc(func(event *events.GuildMemberUpdate) {
			mainLogic(event.Member)
		}),
		disbot.NewListenerFunc(func(event *events.GuildMemberJoin) {
			mainLogic(event.Member)
		}),
	)

	return nil
}

var welcomeMessage = pkg.LocalizedString{
	discord.LocaleEnglishUS: "Hello %s!\n\nBefore you join our ship, we need to ask you a few questions to determine your role. You will be asked a series of questions, please answer them truthfully.\n\nAre you ready?",
	discord.LocaleFrench:    "Bonjour %s!\n\nAvant de rejoindre notre vaisseau, nous devons te poser quelques questions pour déterminer votre rôle. Tu seras invité à répondre à une série de questions, merci de répondre honnêtement.\n\nEs-tu prêt?",
}

func (f SuspiciousInterwiewFeature) ExecuteInterview(bot *pkg.Bot, member discord.Member, locale discord.Locale) error {
	slog.Info(fmt.Sprintf("Running '%s' for user '%s' from guild '%s'", f.Name()[discord.LocaleEnglishUS], member.EffectiveName(), member.GuildID.String()))

	restClient := bot.Client.Rest()
	userDMChannel, err := restClient.CreateDMChannel(member.User.ID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel for user '%s' of guild '%s': %w", member.EffectiveName(), member.GuildID.String(), err)
	}

	// Create the welcome message
	welcomeMsgCreate := discord.NewMessageCreateBuilder().SetContent(welcomeMessage[locale]).Build()

	restClient.CreateMessage(userDMChannel.ID(), welcomeMsgCreate)

	// Processing each question
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
