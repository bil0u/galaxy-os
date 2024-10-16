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

	"github.com/bil0u/galaxy-os/sdk/utils"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
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
	AvailableFeatures BotFeatureSet
}

// NewBotClient creates a new bot client, with the provided token and parts
func NewBotClient(token string, intents []gateway.Intents, caches []cache.Flags) (bot.Client, error) {
	return disgo.New(token,
		bot.WithGatewayConfigOpts(gateway.WithIntents(intents...)),
		bot.WithCacheConfigOpts(cache.WithCaches(caches...)),
	)
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
func (b *Bot) Setup(features BotFeatureSet) error {

	b.AvailableFeatures = features

	slog.Info("Setting up bot...")

	// Add default listeners
	b.Client.AddEventListeners(b.Paginator)

	// Fetching each guild the bot is in
	botGuilds, err := b.Client.Rest().GetCurrentUserGuilds("", 0, 0, 0, true)
	if err != nil {
		return fmt.Errorf("failed to fetch bot guilds: %w", err)
	}

	// For each guild, create the default config, and override it using a local toml config if it exists
	for _, botGuild := range botGuilds {
		guildConfig := NewGuildConfig(botGuild.ID, b.Name, b.AvailableFeatures)
		b.Config.Guilds = append(b.Config.Guilds, guildConfig)
	}

	// Validate the configuration
	if errs := b.Config.Validate(); len(errs) > 0 {
		return fmt.Errorf("invalid configuration: %v", errs)
	}

	// Setup each feature
	for _, feature := range features {
		if err := feature.Setup(b); err != nil {
			return fmt.Errorf("failed to setup feature '%s': %w", feature.Name(), err)
		}
	}

	return nil
}

func (b *Bot) SyncCommands() error {

	var errors []error

	slog.Info("Syncing commands to Discord API...")
	// Loop through each guild and sync only the commands that are enabled
	for _, guildCfg := range b.Config.Guilds {

		slog.Info(fmt.Sprintf(" > Guild '%s'", guildCfg.ID.String()))

		// Loop through each feature and get the commands to sync
		var syncCommands []discord.ApplicationCommandCreate
		for _, feature := range guildCfg.Features.GetFeatures(true) {
			cmds := feature.CommandsCreate()
			if cmds != nil {
				slog.Info(fmt.Sprintf("   - Commands from feature '%s' will be synced", feature.Name()))
				syncCommands = append(syncCommands, cmds...)
			}
		}

		// If no commands to sync, skip
		if len(syncCommands) == 0 {
			slog.Info("   No commands to sync")
			continue
		}

		// Otherwise, sync the commands
		if err := handler.SyncCommands(b.Client, syncCommands, []snowflake.ID{guildCfg.ID}); err != nil {
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

func (b *Bot) LogPermissions(devGuildsOnly bool) {
	// Checking permissions for each Guild
	for _, guildID := range b.Config.GetGuildsIDs(devGuildsOnly) {
		rolePerms, userPerms, channelOverwrites, err := utils.CheckBotPermissions(b.Client, guildID)
		if err != nil {
			slog.Error("Error checking bot permissions:", slog.Any("err", err))
		}
		slog.Info(fmt.Sprintf("[BOT PERMISSIONS - GUILD '%s']:", guildID.String()))
		slog.Info(fmt.Sprintf("- Role permissions:\n%v", rolePerms.String()))
		slog.Info(fmt.Sprintf("- User permissions:\n%v", userPerms.String()))
		slog.Info("- Channel overwrites:")
		for channelID, overwrites := range channelOverwrites {
			slog.Info(fmt.Sprintf("  - <Channel '%s'>:", channelID.String()))
			for _, overwrite := range overwrites {
				switch overwrite.Type() {
				case discord.PermissionOverwriteTypeRole:
					roleOverwrite := overwrite.(discord.RolePermissionOverwrite)
					slog.Info(fmt.Sprintf("    > Role '%s':", roleOverwrite.RoleID.String()))
					slog.Info(fmt.Sprintf("      - Allow: %v", roleOverwrite.Allow.String()))
					slog.Info(fmt.Sprintf("      - Deny: %v", roleOverwrite.Deny.String()))
				case discord.PermissionOverwriteTypeMember:
					memberOverwrite := overwrite.(discord.MemberPermissionOverwrite)
					slog.Info(fmt.Sprintf("    > User '%s':", memberOverwrite.UserID.String()))
					slog.Info(fmt.Sprintf("      - Allow: %v", memberOverwrite.Allow.String()))
					slog.Info(fmt.Sprintf("      - Deny: %v", memberOverwrite.Deny.String()))
				}
			}
		}
	}
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
	if err := b.Client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("failed to open gateway: %w", err)
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
