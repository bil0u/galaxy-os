package service

import (
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
)

var scheduler *cron.Cron

func InitCron() *cron.Cron {
	scheduler = cron.New(cron.WithLogger(cronLogger{}), cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))
	return scheduler
}

func Scheduler() *cron.Cron {
	return scheduler
}

func StartCron() {
	scheduler.Start()
}

func StopCron() {
	closeCtx := scheduler.Stop()
	<-closeCtx.Done()
}

type cronLogger struct{}

func (cl cronLogger) Info(msg string, keysAndValues ...any) {
	slog.Info(fmt.Sprintf("Cron: %s", msg), keysAndValues...)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...any) {
	slog.Error(fmt.Sprintf("Cron: %s: %v", msg, err), keysAndValues...)
}
