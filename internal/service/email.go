package service

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

var emailClient *sesv2.Client

var (
	emailSender  = ""
	emailCharset = ""
)

type AWSEmail struct {
	Client  *sesv2.Client
	subject string
	emails  []string
	content *string
}

func InitEmail(ctx context.Context) (*sesv2.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
		config.WithSharedConfigProfile("serendipe"),
	)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	emailClient = sesv2.NewFromConfig(cfg)
	return emailClient, nil
}

func EmailClient() *sesv2.Client {
	return emailClient
}

func (mail *AWSEmail) Send(ctx context.Context) error {
	svc := mail.Client

	sender := emailSender
	chartset := emailCharset

	_, err := svc.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &sender,
		Destination: &types.Destination{
			ToAddresses: mail.emails,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    &mail.subject,
					Charset: &chartset,
				},
				Body: &types.Body{
					Text: &types.Content{
						Data:    mail.content,
						Charset: &chartset,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
