package email

import (
	"context"
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

var (
	Client *sesv2.Client
)

func Init(ctx context.Context) *sesv2.Client {
	sync.OnceFunc(func() {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion("eu-west-1"),
			config.WithSharedConfigProfile("serendipe"),
		)
		if err != nil {
			slog.Error("Error loading AWS config.", slog.Any("error", err))
			panic(err)
		}

		Client = sesv2.NewFromConfig(cfg)
	})()

	return Client
}
