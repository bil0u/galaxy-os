package cron

import (
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
