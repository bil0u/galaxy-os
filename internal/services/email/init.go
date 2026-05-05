package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

var client *sesv2.Client

func Init(ctx context.Context) (*sesv2.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
		config.WithSharedConfigProfile("serendipe"),
	)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	client = sesv2.NewFromConfig(cfg)
	return client, nil
}

func Client() *sesv2.Client {
	return client
}
