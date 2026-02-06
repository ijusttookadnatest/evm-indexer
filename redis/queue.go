package redis

import (
	"context"
	"fmt"
)
	
func Producer (ctx context.Context, data []byte) error {
	redis, _ := Get()
	_, err := redis.LPush(ctx, "queue", data).Result()
	if err != nil {
		return err
	}
	return nil
}

func Consumer(ctx context.Context) error {
	redis, _ := Get()
	res, err := redis.BRPop(ctx, 0, "queue").Result()
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}