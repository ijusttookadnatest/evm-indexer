package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var offsetDefault uint64 = 100
var rangeTimeDefault uint64 = 3600
var concurrencyDefault int = 2
var fromDefault uint64 = 0

type Config struct {
	PostgresDSN string
	RedisDSN    string
	Rpc     	string
	Port		string
	PlaygroundEnabled bool
	RangeMaxTime uint64
	OffsetMax	uint64
	From		uint64
	ConcurrencyF int
}


func loadEnv(path string) (map[string]string, error) {
	envFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer envFile.Close()
	env, err := godotenv.Parse(envFile)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func Load(path string) (*Config,error) {
	var rangeMaxTime, offsetMax, from uint64
	var concurrencyF int

	env, err := loadEnv(path)
	if err != nil {
		return nil, err
	}

	pgEnabled := env["PLAYGROUND_ENABLED"] == "true"
	if env["MAX_TIME"] != "" {
		rangeMaxTime, err = strconv.ParseUint(env["MAX_TIME"], 10, 64)
		if err != nil {
			return nil, err
		}
	} else {
		rangeMaxTime = rangeTimeDefault
	}

	if env["MAX_OFFSET"] != "" {
		offsetMax, err = strconv.ParseUint(env["MAX_OFFSET"], 10, 64)
		if err != nil {
			return nil, err
		}
	} else {
		offsetMax = offsetDefault
	}

	if env["FROM"] != "" {
		from, err = strconv.ParseUint(env["FROM"], 10, 64)
		if err != nil {
			return nil, err
		}
	} else {
		from = fromDefault
	}

	if env["CONCURRENCY_FACTOR"] != "" {
		concurrencyF, err = strconv.Atoi(env["CONCURRENCY_FACTOR"])
		if err != nil {
			return nil, err
		}
	} else {
		concurrencyF = concurrencyDefault
	}

	return &Config{
		PostgresDSN: fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=disable", env["POSTGRES_USER"], env["POSTGRES_PASSWORD"], env["POSTGRES_HOST"], env["POSTGRES_PORT"], env["POSTGRES_DB"]),
		RedisDSN:    fmt.Sprintf("redis://:%v@%v:%v/%v", env["REDIS_PASSWORD"], env["REDIS_HOST"], env["REDIS_PORT"], env["REDIS_DB"]),
		Rpc: env["RPC_URL"],
		Port: env["PORT"],
		PlaygroundEnabled: pgEnabled,
		RangeMaxTime: rangeMaxTime,
		OffsetMax: offsetMax,
		From: from,
		ConcurrencyF: concurrencyF,
	}, nil
}
