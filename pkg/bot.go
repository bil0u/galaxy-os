package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/oauth2"
	"github.com/disgoorg/paginator"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron/v3"
)

type Bot struct {
	Name              string
	Version           string
	Commit            string
	Config            Config
	Client            bot.Client
	OAuthClient       oauth2.Client
	Router            *handler.Mux
	Paginator         *paginator.Manager
	Cron              *cron.Cron
	CommandsToSync    []discord.ApplicationCommandCreate
	AvailableFeatures BotFeatureSet
}

// NewBot creates a new bot instance
func NewBot(client bot.Client, cfg Config, name string, version string, commit string) *Bot {
	return &Bot{
		Name:              name,
		Version:           version,
		Commit:            commit,
		Config:            cfg,
		Client:            client,
		Router:            handler.New(),
		Paginator:         paginator.New(),
		Cron:              nil,
		AvailableFeatures: nil,
	}
}

// SetupBot sets up the bot with the provided parts
func (b *Bot) Setup(guilds []snowflake.ID, features BotFeatureSet) error {

	b.AvailableFeatures = features

	slog.Info("Setting up bot...")

	// Add default listeners
	b.Client.AddEventListeners(b.Paginator)

	// For each guild, create the default config, and override it using a local toml config if it exists
	for _, guildID := range guilds {
		guildConfig := NewGuildConfig(guildID, b.Name, b.AvailableFeatures)
		b.Config.Guilds = append(b.Config.Guilds, guildConfig)
	}

	// Validate the configuration
	if errs := b.Config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid configuration: %v", errs)
	}

	// Start each feature using the feature kit
	for _, feature := range features {
		if err := feature.Setup(b); err != nil {
			return fmt.Errorf("failed to setup feature '%s': %w", feature.Name(), err)
		}
	}

	return nil
}

// AddCommandsToSync adds commands to the list of commands to sync
func (b *Bot) AddCommandsToSync(commands ...discord.ApplicationCommandCreate) {
	b.CommandsToSync = append(b.CommandsToSync, commands...)
}

// syncCommands syncs the commands to the Discord API
func (b *Bot) syncCommands() error {

	var errors []error

	slog.Info("Syncing commands to Discord API...")
	// Loop through each guild and sync only the commands that are enabled
	for _, guildCfg := range b.Config.Guilds {

		slog.Info(fmt.Sprintf(" > Guild '%s'", guildCfg.ID.String()))

		// If no commands to sync, skip
		if len(b.CommandsToSync) == 0 {
			slog.Info("   No commands to sync")
			continue
		}

		// Otherwise, sync the commands
		if err := handler.SyncCommands(b.Client, b.CommandsToSync, []snowflake.ID{guildCfg.ID}); err != nil {
			errors = append(errors, err)
		}
		slog.Info("   Commands successfully synced")
	}

	// Return any errors
	if len(errors) > 0 {
		return fmt.Errorf("encountered errors while trying to sync commands: %v", errors)
	}
	return nil
}

// Start starts the bot
func (b *Bot) Start(syncCommands, enableCron, enableOAuth2 bool) error {

	slog.Info(fmt.Sprintf("Starting bot '%s' ...", b.Name))

	// Sync commands
	if syncCommands {
		if err := b.syncCommands(); err != nil {
			return fmt.Errorf("failed to sync commands: %w", err)
		}
	}

	// Deferring client close
	slog.Info("Deferring client close")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Client.Close(ctx)
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if b.Client.HasShardManager() {
		slog.Info("Opening gateway using shard manager")
		if err := b.Client.OpenShardManager(ctx); err != nil {
			return fmt.Errorf("failed to open gateway through shard manager: %w", err)
		}
	} else {
		slog.Info("Opening gateway")
		if err := b.Client.OpenGateway(ctx); err != nil {
			return fmt.Errorf("failed to open gateway: %w", err)
		}
	}

	if enableCron {
		slog.Info("Starting Cron")
		b.Cron.Start()
	}

	// Wait for signal to shutdown
	slog.Info("Bot is running 🚀\nPress CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
	b.Stop()
	return nil
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.Cron.Stop()
}
