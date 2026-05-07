package dailymessage

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/bil0u/galaxy-os/internal/contracts"
	"github.com/disgoorg/snowflake/v2"
)

// DailyMessage sends a message every day at a configured time.
var Feature = &dailyMessage{}

type dailyMessage struct {
	logger  *slog.Logger
	rest    contracts.RestClient
	configs contracts.ConfigProvider
	cron    contracts.CronScheduler
	guildID snowflake.ID
}

type dailyMessageConfig struct {
	Enabled bool
	Channel snowflake.ID
	Time    string
}

func (c dailyMessageConfig) Validate() error {
	var errs []error
	if c.Channel == 0 {
		errs = append(errs, fmt.Errorf("channel is required"))
	}
	if c.Time == "" {
		errs = append(errs, fmt.Errorf("time is required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (c dailyMessageConfig) cronSchedule() (string, error) {
	pattern := regexp.MustCompile(`^([0-9]{1,2}):([0-9]{1,2})$`)
	match := pattern.FindStringSubmatch(c.Time)
	if match == nil {
		return "", fmt.Errorf("invalid time format")
	}
	return fmt.Sprintf("%s %s * * *", match[2], match[1]), nil
}

func (f *dailyMessage) Name() string               { return "daily_message" }
func (f *dailyMessage) Scope() contracts.Scope       { return contracts.GuildScope }
func (f *dailyMessage) Needs() []contracts.ServiceID { return []contracts.ServiceID{contracts.CronService} }

func (f *dailyMessage) Setup(deps contracts.Deps) error {
	f.logger = deps.Logger
	f.rest = deps.Rest
	f.configs = deps.Configs
	f.cron = deps.Cron
	f.guildID = deps.GuildID

	// TODO: CronScheduler is currently an empty interface.
	// When it exposes AddFunc(schedule string, cmd func()) (int, error),
	// register the cron job here using cronSchedule() and dailyMessageJob().
	// Previously this iterated all guilds — now the framework calls Setup per guild.

	return nil
}

func (f *dailyMessage) Start(ctx context.Context) error { return nil }
func (f *dailyMessage) Stop(ctx context.Context) error  { return nil }
