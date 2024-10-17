package sdk

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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
	Router            *handler.Mux
	Paginator         *paginator.Manager
	Cron              *cron.Cron
	CommandsToSync    []discord.ApplicationCommandCreate
	AvailableFeatures BotFeatureSet
}

// NewBot creates a new bot instance
func NewBot(client bot.Client, cfg Config, name string, version string, commit string) *Bot {
	return &Bot{
		Name:      name,
		Version:   version,
		Commit:    commit,
		Config:    cfg,
		Client:    client,
		Router:    handler.New(),
		Paginator: paginator.New(),
		Cron: cron.New(
			cron.WithLogger(
				cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags)))),
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

	// Setup each feature using the feature kit
	for _, feature := range features {
		if err := feature.Setup(b); err != nil {
			return fmt.Errorf("failed to setup feature '%s': %w", feature.Name(), err)
		}
	}

	return nil
}

func (b *Bot) AddCommandsToSync(commands ...discord.ApplicationCommandCreate) {
	b.CommandsToSync = append(b.CommandsToSync, commands...)
}

func (b *Bot) SyncCommands() error {

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

func (b *Bot) Start() error {

	slog.Info(fmt.Sprintf("Starting bot '%s' ...", b.Name))

	// Deferring client close
	slog.Info("Deferring client close")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Client.Close(ctx)
	}()

	// Open gateway
	slog.Info("Opening gateway")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if b.Client.HasShardManager() {
		// if err := b.Client.OpenGateway(ctx); err != nil {
		if err := b.Client.OpenShardManager(ctx); err != nil {
			return fmt.Errorf("failed to open gateway through shard manager: %w", err)
		}
	} else {
		if err := b.Client.OpenGateway(ctx); err != nil {
			return fmt.Errorf("failed to open gateway: %w", err)
		}
	}

	// Start cron
	slog.Info("Starting cron")
	b.Cron.Start()

	// Wait for signal to shutdown
	slog.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
	b.Cron.Stop()
	return nil
}
