package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/feature"
	"github.com/bil0u/galaxy-os/internal/platform"
	"github.com/bil0u/galaxy-os/internal/service"
	"github.com/disgoorg/disgo"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/spf13/cobra"
)

var (
	enableCron   bool
	enableOAuth2 bool
	syncCommands bool
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Main command for the bot",
}

func init() {
	botCmd.AddCommand(startCmd)

	startCmd.Flags().BoolVarP(&enableCron, "cron", "c", false, "Whether to enable cron jobs")
	startCmd.Flags().BoolVarP(&enableOAuth2, "oauth2", "o", false, "Whether to enable oauth2 server")
	startCmd.Flags().BoolVarP(&syncCommands, "sync", "s", false, "Sync slash commands to Discord before starting")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the discord bot",
	RunE:  startBot,
}

type botDef struct {
	features   []platform.Feature
	cacheFlags cache.Flags
	intents    gateway.Intents
}

var bots = map[string]botDef{
	"hue": {
		features: []platform.Feature{
			feature.Info,
			feature.Permissions,
			feature.Test,
			feature.Presence,
			feature.SelfAssign,
			feature.DailyMessage,
			feature.SuspiciousInterview,
		},
		cacheFlags: cache.FlagGuilds | cache.FlagMembers | cache.FlagRoles,
		intents:    gateway.IntentGuilds | gateway.IntentGuildMembers,
	},
	"kevin": {
		features: []platform.Feature{
			feature.Info,
			feature.Permissions,
			feature.Test,
			feature.Presence,
			feature.SelfAssign,
			feature.DailyMessage,
		},
		cacheFlags: cache.FlagGuilds | cache.FlagRoles,
		intents:    gateway.IntentGuilds,
	},
}

// muxRegistrar adapts handler.Mux to platform.Registrar.
// disgo v0.19.3 handler types include a typed data parameter;
// the platform.Registrar signatures omit it for simplicity.
type muxRegistrar struct {
	mux *handler.Mux
}

func (r *muxRegistrar) SlashCommand(path string, h platform.SlashCommandHandler) {
	r.mux.SlashCommand(path, func(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
		return h(e)
	})
}

func (r *muxRegistrar) ButtonComponent(customID string, h platform.ButtonComponentHandler) {
	r.mux.ButtonComponent(customID, func(_ discord.ButtonInteractionData, e *handler.ComponentEvent) error {
		return h(e)
	})
}

func (r *muxRegistrar) Autocomplete(path string, h platform.AutocompleteHandler) {
	r.mux.Autocomplete(path, func(e *handler.AutocompleteEvent) error {
		return h(e)
	})
}

func startBot(cmd *cobra.Command, _ []string) error {
	def, ok := bots[bot]
	if !ok {
		return fmt.Errorf("bot %q not found", bot)
	}
	if len(def.features) == 0 {
		return fmt.Errorf("no features defined for bot %q", bot)
	}

	globalCfg, logCfg, botCfg, err := config.Init(bot)
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	globalCfg.Development = development == "true"
	globalCfg.Version = version
	globalCfg.Commit = commit

	if _, err := service.InitLogger(logCfg.Level, logCfg.Format, logCfg.AddSource); err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}

	botLogger := slog.Default().With("bot", bot)

	client, err := disgo.New(botCfg.Token,
		disbot.WithCacheConfigOpts(cache.WithCaches(def.cacheFlags)),
		disbot.WithGatewayConfigOpts(
			gateway.WithIntents(def.intents),
		),
	)
	if err != nil {
		return fmt.Errorf("building discord client: %w", err)
	}

	ctx := context.Background()

	guilds, err := config.InitGuilds(ctx, client.Rest, bot)
	if err != nil {
		return fmt.Errorf("initializing guilds config: %w", err)
	}

	slog.Debug(fmt.Sprintf("Global configuration: %+v", globalCfg))
	slog.Debug(fmt.Sprintf("Bot configuration: %++v", botCfg))
	slog.Debug(fmt.Sprintf("Log configuration: %+v", logCfg))
	slog.Debug(fmt.Sprintf("Guilds configuration: %+v", guilds))

	router := handler.New()
	client.AddEventListeners(router)

	client.AddEventListeners(disbot.NewListenerFunc(func(_ *events.Resumed) {
		botLogger.Info("gateway reconnected")
	}))

	if enableCron {
		service.InitCron()
	}

	if enableOAuth2 {
		if err := botCfg.ValidateOAuth(); err != nil {
			return fmt.Errorf("oauth2 config: %w", err)
		}
		service.InitOAuth(botCfg.ApplicationID, botCfg.ClientSecret, botCfg.BaseURL)
	}

	registrar := &muxRegistrar{mux: router}

	// Setup features: iterate each feature and call Setup with platform.Deps.
	// BotScope features are set up once. GuildScope features are set up per guild.
	var errs []error
	guildIDs := guilds.IDs(development == "true")
	for _, f := range def.features {
		switch f.Scope() {
		case platform.BotScope:
			deps := platform.Deps{
				Logger:   botLogger.With("feature", f.Name()),
				Commands: registrar,
				BotName:  bot,
				// TODO: wire Locale, Bus, Configs, Env, Rest when implementations exist
			}
			if err := f.Setup(deps); err != nil {
				errs = append(errs, fmt.Errorf("setting up feature %q: %w", f.Name(), err))
			}
		case platform.GuildScope:
			for _, guildID := range guildIDs {
				deps := platform.Deps{
					Logger:   botLogger.With("feature", f.Name(), "guild", guildID),
					Commands: registrar,
					BotName:  bot,
					GuildID:  guildID,
					// TODO: wire Locale, Bus, Configs, Env, Rest, Guilds when implementations exist
				}
				if err := f.Setup(deps); err != nil {
					errs = append(errs, fmt.Errorf("setting up feature %q for guild %s: %w", f.Name(), guildID, err))
				}
			}
		case platform.CrossGuildScope:
			deps := platform.Deps{
				Logger:   botLogger.With("feature", f.Name()),
				Commands: registrar,
				BotName:  bot,
				// TODO: wire Locale, Bus, Configs, Env, Rest, Guilds when implementations exist
			}
			if err := f.Setup(deps); err != nil {
				errs = append(errs, fmt.Errorf("setting up feature %q: %w", f.Name(), err))
			}
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("setting up features: %w", err)
	}

	// TODO: command syncing was previously handled by feature.SyncCommands.
	// Re-implement when features declare their command creates via platform.Feature.
	if syncCommands {
		slog.Warn("command syncing not yet implemented with new feature framework")
	}

	if err := client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("opening gateway: %w", err)
	}

	if enableCron {
		service.StartCron()
	}

	if enableOAuth2 {
		service.StartOAuth()
	}

	defer func() {
		if enableCron {
			service.StopCron()
		}
		withTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client.Close(withTimeout)
	}()

	slog.Info("----- Bot is running 🚀 Press CTRL-C to exit -----")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
	return nil
}
