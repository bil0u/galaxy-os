package cron

import (
	"sync"

	"github.com/robfig/cron/v3"
)

var (
	Cron *cron.Cron
)

func InitCron() *cron.Cron {
	sync.OnceFunc(func() {
		Cron = cron.New(cron.WithLogger(cronLogger{}), cron.WithChain(
			cron.Recover(cron.DefaultLogger),
		))
	})()

	return Cron
}

func StartCron() {
	Cron.Start()
}

func StopCron() {
	closeCtx := Cron.Stop()
	<-closeCtx.Done()
}
