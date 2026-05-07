package contracts

import "context"

// Bus enables optional coordination between bots.
// Default is NoBus — all operations succeed silently.
type Bus interface {
	Publish(ctx context.Context, topic string, payload any) error
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, payload any)) error
}

// NoBus is the default Bus implementation that silently succeeds.
var NoBus Bus = noBus{}

type noBus struct{}

func (noBus) Publish(context.Context, string, any) error { return nil }
func (noBus) Subscribe(context.Context, string, func(context.Context, any)) error { return nil }
