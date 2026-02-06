package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"github/ijusttookadnatest/indexer-evm/config"
)

var initCache = sync.OnceValues(func() (*redis.Client,error) {
	ctx := context.Background()

	opt, err := redis.ParseURL(config.Get().RedisDSN)
	if err != nil {
		return nil, err
	}
	cache := redis.NewClient(opt)

    _, err = cache.Ping(ctx).Result()
    if err != nil {
        panic(err)
    }

	fmt.Println("Connection to redis successfull")
	return cache, nil
})

func Get() (*redis.Client,error) {
	return initCache()
}