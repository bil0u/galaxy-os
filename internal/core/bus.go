package core

import "context"

// InnerBus enables optional coordination between features and bots.
// Default is VoidBus — all operations succeed silently.
type InnerBus interface {
	Publish(ctx context.Context, topic string, payload any) error
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, payload any)) error
}

// VoidBus is the default InnerBus implementation that silently succeeds.
var VoidBus InnerBus = voidBus{}

type voidBus struct{}

func (voidBus) Publish(context.Context, string, any) error                          { return nil }
func (voidBus) Subscribe(context.Context, string, func(context.Context, any)) error { return nil }
