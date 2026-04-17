package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var offsetDefault uint64 = 100
var rangeTimeDefault uint64 = 3600
var concurrencyDefault int = 2
var fromDefault uint64 = 0
var rpcRateLimitDefault float64 = 1

type Config struct {
	PostgresDSN       string
	RedisDSN          string
	RpcHTTP           string
	RpcWS             string
	Port              string
	PlaygroundEnabled bool
	RangeMaxTime      uint64
	OffsetMax         uint64
	RpcRateLimit      float64
	From              uint64
	ConcurrencyF      int
}

func loadEnv(path string) (map[string]string, error) {
	envFile, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	defer envFile.Close()
	env, err := godotenv.Parse(envFile)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func getenv(env map[string]string, key string) string {
	if v, ok := env[key]; ok {
		return v
	}
	return os.Getenv(key)
}

func Load(path string) (*Config, error) {
	var rangeMaxTime, offsetMax, from uint64
	var concurrencyF int
	var rpcRateLimit float64

	env, err := loadEnv(path)
	if err != nil {
		return nil, err
	}

	pgEnabled := getenv(env, "PLAYGROUND_ENABLED") == "true"
	if getenv(env, "MAX_TIME") != "" {
		rangeMaxTime, err = strconv.ParseUint(getenv(env, "MAX_TIME"), 10, 64)
		if err != nil {
			return nil, err
		}
	} else {
		rangeMaxTime = rangeTimeDefault
	}

	if getenv(env, "MAX_OFFSET") != "" {
		offsetMax, err = strconv.ParseUint(getenv(env, "MAX_OFFSET"), 10, 64)
		if err != nil {
			return nil, err
		}
	} else {
		offsetMax = offsetDefault
	}

	if getenv(env, "FROM") != "" {
		from, err = strconv.ParseUint(getenv(env, "FROM"), 10, 64)
		if err != nil {
			return nil, err
		}
	} else {
		from = fromDefault
	}

	if getenv(env, "RPC_RATE_LIMIT") != "" {
		rpcRateLimit, err = strconv.ParseFloat(getenv(env, "RPC_RATE_LIMIT"), 64)
		if err != nil {
			return nil, err
		}
	} else {
		rpcRateLimit = rpcRateLimitDefault
	}

	if getenv(env, "CONCURRENCY_FACTOR") != "" {
		concurrencyF, err = strconv.Atoi(getenv(env, "CONCURRENCY_FACTOR"))
		if err != nil {
			return nil, err
		}
	} else {
		concurrencyF = concurrencyDefault
	}

	return &Config{
		PostgresDSN:       getenv(env, "POSTGRES_DSN"),
		RedisDSN:          getenv(env, "REDIS_DSN"),
		RpcHTTP:           getenv(env, "RPC_HTTP"),
		RpcWS:             getenv(env, "RPC_WS"),
		Port:              getenv(env, "PORT"),
		PlaygroundEnabled: pgEnabled,
		RangeMaxTime:      rangeMaxTime,
		OffsetMax:         offsetMax,
		RpcRateLimit:      rpcRateLimit,
		From:              from,
		ConcurrencyF:      concurrencyF,
	}, nil
}
