package config

import (
	"fmt"
	"strconv"
)

type Config struct {
	PostgresDSN string
	RedisDSN    string
	Rpc     	string
	LastBlocks  int
}

var config *Config

func Init(env map[string]string) error {
	block, err := strconv.Atoi(env["LAST_BLOCKS"])
	if err != nil {
		return err
	}
	config = &Config{
		PostgresDSN: fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=disable", env["POSTGRES_USER"], env["POSTGRES_PASSWORD"], env["POSTGRES_HOST"], env["POSTGRES_PORT"], env["POSTGRES_DB"]),
		RedisDSN:    fmt.Sprintf("redis://:%v@%v:%v/%v", env["REDIS_PASSWORD"], env["REDIS_HOST"], env["REDIS_PORT"], env["REDIS_DB"]),
		Rpc: env["RPC_URL"],
		LastBlocks: block,
	}
	return nil
}

func Get() *Config {
	return config
}
