package ports

import "context"

type RedisPubSub interface {
	Subscribe(ctx context.Context, topic string) (<-chan[]byte, error)
	Publish(ctx context.Context, topic string, payload []byte) error
}