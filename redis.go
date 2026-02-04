package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

var initCache = sync.OnceValues(func() (*redis.Client,error) {
	ctx := context.Background()

	opt, err := redis.ParseURL(config.RedisDSN)
	if err != nil {
		return nil, err
	}
	cache := redis.NewClient(opt)

	err = cache.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		return nil, err
	}
	_, err = cache.Get(ctx, "foo").Result()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connection to cache successfull")
	return cache, nil
})

func GetCache() (*redis.Client,error) {
	return initCache()
}