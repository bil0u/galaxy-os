package service

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/bil0u/galaxy-os/internal/platform"
)

var (
	emailSender  = ""
	emailCharset = ""
)

// EmailService wraps AWS SES as a platform.Service.
type EmailService struct {
	client *sesv2.Client
}

// NewEmailService creates an EmailService.
func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) Name() string { return "email" }

func (s *EmailService) Start(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
		config.WithSharedConfigProfile("serendipe"),
	)
	if err != nil {
		return fmt.Errorf("loading aws config: %w", err)
	}

	s.client = sesv2.NewFromConfig(cfg)
	return nil
}

func (s *EmailService) Health(_ context.Context) platform.Health {
	status := platform.StatusDown
	if s.client != nil {
		status = platform.StatusUp
	}
	return platform.Health{
		Name:   s.Name(),
		Status: status,
	}
}

func (s *EmailService) Stop(_ context.Context) error {
	s.client = nil
	return nil
}

// Client returns the underlying SES client.
func (s *EmailService) Client() *sesv2.Client {
	return s.client
}

// AWSEmail holds the data for sending an email.
type AWSEmail struct {
	Client  *sesv2.Client
	subject string
	emails  []string
	content *string
}

func (mail *AWSEmail) Send(ctx context.Context) error {
	svc := mail.Client

	sender := emailSender
	charset := emailCharset

	_, err := svc.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &sender,
		Destination: &types.Destination{
			ToAddresses: mail.emails,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    &mail.subject,
					Charset: &charset,
				},
				Body: &types.Body{
					Text: &types.Content{
						Data:    mail.content,
						Charset: &charset,
					},
				},
			},
		},
	})
	return err
}
