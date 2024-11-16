package cron

import (
	"context"
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

func StartCron(ctx context.Context) {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		Cron.Start()
	}()

	select {
	case <-ctx.Done():
		wg.Done()
		closeCtx := Cron.Stop()
		<-closeCtx.Done()
	default:
		wg.Wait()
	}
}
