package pubsub

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func New(dsn string) (*redis.Client, error) {
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	conn := redis.NewClient(opt)

	_, err = conn.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connection to redis successfull")
	return conn, nil
}
