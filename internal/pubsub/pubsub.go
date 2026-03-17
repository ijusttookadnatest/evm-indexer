package pubsub

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type RedisPubSub struct {
	client *redis.Client
}

func NewRedisPubSub(client *redis.Client) *RedisPubSub {
	return &RedisPubSub{client:client}
}

func (r *RedisPubSub) Subscribe(ctx context.Context, topic string) (<-chan[]byte, error) {
	sub := r.client.Subscribe(ctx, topic)
	
	if _, err := sub.Receive(ctx) ; err != nil {
		return nil, err
	}

	out := make(chan[]byte, 10)
	
	go func() {
		ch := sub.Channel()
		defer close(out)
		defer sub.Close()

		for {
			select {
			case <-ctx.Done(): {
				slog.Error("pubsub subscribe: context cancelled", "err", ctx.Err())
				return
			}
			case msg := <-ch: {
				if msg == nil {
					return
				}
				out <- []byte(msg.Payload)
			}
			}
		}
	}()

	return out, nil
}

func (r *RedisPubSub) Publish(ctx context.Context, topic string, payload []byte) error {
	return r.client.Publish(ctx, topic, payload).Err()
}