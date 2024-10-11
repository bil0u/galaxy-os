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

	"github.com/bil0u/galaxy-os/sdk/enums"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/paginator"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron/v3"
)

func NewBot(cfg Config, name string, version string, commit string) *Bot {
	return &Bot{
		Name:      name,
		Version:   version,
		Commit:    commit,
		Cfg:       cfg,
		Client:    nil,
		Paginator: paginator.New(),
		Cron: cron.New(
			cron.WithLogger(
				cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags)))),
	}
}

func NewBotClient(token string, parts BotParts) (*bot.Client, error) {
	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(gateway.WithIntents(parts.Intents...)),
		bot.WithCacheConfigOpts(cache.WithCaches(parts.Caches...)),
	)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

type Bot struct {
	Name      string
	Version   string
	Commit    string
	Cfg       Config
	Client    bot.Client
	Paginator *paginator.Manager
	Cron      *cron.Cron
}

// SetupBot sets up the bot with the provided parts
func (b *Bot) SetupBot(parts BotParts) error {

	// Add default listeners
	b.Client.AddEventListeners(b.Paginator)
	b.Client.AddEventListeners(bot.NewListenerFunc(b.OnReady))

	// Create router and register it as an event listener
	router := parts.CreateRouter(b)
	b.Client.AddEventListeners(router)

	// Create bot listeners
	listeners := parts.CreateListeners(b)
	b.Client.AddEventListeners(listeners...)

	// Registering each cron jobs for each guild, and adding the guild timezone if set
	for guildID, guildCfg := range b.Cfg.Guilds {

		timezonePrefix := ""
		// If a timezone is set for the guild, we need to prefix the cron job with the timezone
		if guildCfg.Timezone != "" {
			timezonePrefix = fmt.Sprintf("CRON_TZ=%s ", guildCfg.Timezone)
		}

		for _, cronJob := range parts.CronJobs {
			job := cronJob.Creator(b, guildID)
			_, err := b.Cron.AddFunc(fmt.Sprintf("%s%s", timezonePrefix, cronJob.Schedule), job)
			if err != nil {
				slog.Error("Failed to add cron job", slog.Any("err", err))
				continue
			}
			slog.Info(fmt.Sprintf("Added cron job '%s' for guild '%s'", cronJob.Name, guildID), slog.Any("schedule", cronJob.Schedule))
		}
	}

	return nil
}

func (b *Bot) GetGuild(guildID snowflake.ID) (*discord.Guild, error) {
	restClient := b.Client.Rest()
	guild, err := restClient.GetGuild(guildID, false)
	if err != nil {
		return nil, err
	}
	return &guild.Guild, nil
}

func (b *Bot) getRoles(guildID snowflake.ID) ([]enums.RoleEnum, error) {
	restClient := b.Client.Rest()

	botUser, err := restClient.GetMember(guildID, b.Client.ApplicationID())
	if err != nil {
		return nil, err
	}

	var roles []enums.RoleEnum
	for _, roleID := range botUser.RoleIDs {
		if role := enums.GetRoleEnum(roleID); role.IsValid() {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

func (b *Bot) SelfRemoveRoles(guildID snowflake.ID, roles []enums.RoleEnum) error {

	restClient := b.Client.Rest()

	// Getting user using the bot ID
	botUser, err := restClient.GetCurrentUser("")
	if err != nil {
		return err
	}

	for _, role := range roles {
		r, err := restClient.GetRole(guildID, role.ID())
		if err != nil {
			slog.Error("Failed to get role", slog.Any("err", err))
			continue
		}
		if !r.Managed {
			err := restClient.RemoveMemberRole(guildID, botUser.ID, r.ID)
			if err != nil {
				slog.Error("Failed to remove role", slog.Any("err", err))
				continue
			}
			slog.Info("Successfully removed role", slog.Any("role", r.Name))
		}
	}

	return nil
}

func (b *Bot) SelfAssignRoles(guildID snowflake.ID, roles []enums.RoleEnum) error {

	restClient := b.Client.Rest()

	// Getting user using the bot ID
	botUser, err := restClient.GetCurrentUser("")
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Syncing roles for guild '%s'", guildID), slog.Any("roles", roles))

	for _, role := range roles {
		// Assign each role to the bot

		err := restClient.AddMemberRole(guildID, botUser.ID, role.ID())

		if err != nil {
			slog.Error(fmt.Sprintf("Failed to assign role '%s' to bot:", role.String()), slog.Any("err", err))
		} else {
			slog.Info(fmt.Sprintf("Successfully assigned role '%s' to bot in guild '%s'", role.String(), guildID.String()))
		}
	}

	return nil
}

func (b *Bot) UpdateApplicationInfos(appUpdate discord.ApplicationUpdate) error {

	restClient := b.Client.Rest()
	_, err := restClient.UpdateCurrentApplication(appUpdate)

	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) Start(syncCommands []discord.ApplicationCommandCreate, syncRoles bool) {

	slog.Info(fmt.Sprintf("Starting bot '%s' ...", b.Name))

	// Deferring client close
	slog.Info("Deferring client close")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Client.Close(ctx)
	}()

	syncToGuilds := b.Cfg.GetDevGuildsIDs()
	if len(syncToGuilds) == 0 {
		syncToGuilds = b.Cfg.GetGuildsIDs()
	}

	// Sync roles if needed
	if syncRoles {

		// Assigning roles to the bot
		for _, guildID := range syncToGuilds {

			guildRoles := b.Cfg.GetGuildRoles(guildID)

			slog.Info(fmt.Sprintf("Clearing roles for guild '%s'", guildID))

			// Clearing existing roles except the auto assigned ones
			rolesToClear, err := b.getRoles(guildID)
			if err != nil {
				slog.Error("Failed to get roles", slog.Any("err", err))
			}

			b.SelfRemoveRoles(guildID, rolesToClear)

			if err := b.SelfAssignRoles(guildID, guildRoles); err != nil {
				slog.Error("Failed to assign roles to bot", slog.Any("err", err))
			}
		}

	}

	// Sync commands if needed
	if syncCommands != nil {
		slog.Info("Syncing commands", slog.Any("guilds", syncToGuilds))
		if err := handler.SyncCommands(b.Client, syncCommands, syncToGuilds); err != nil {
			slog.Error("Failed to sync commands", slog.Any("err", err))
		}
	}

	// Open gateway
	slog.Info("Opening gateway")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := b.Client.OpenGateway(ctx); err != nil {
		slog.Error("Failed to open gateway", slog.Any("err", err))
		os.Exit(-1)
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
}

func (b *Bot) OnReady(_ *events.Ready) {
	slog.Info(fmt.Sprintf("Bot '%s' is ready", b.Name))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.Client.SetPresence(ctx, gateway.WithCustomActivity("Loading Kernel..."), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
		slog.Error("Failed to set presence", slog.Any("err", err))
	}
}
