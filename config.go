package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
    PostgresDSN string
    RedisDSN    string
}

var config *Config

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

func initConfig(env map[string]string) {
	config = &Config{
		PostgresDSN: fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=disable", env["POSTGRES_USER"], env["POSTGRES_PASSWORD"], env["POSTGRES_HOST"], env["POSTGRES_PORT"], env["POSTGRES_DB"]),
		RedisDSN: fmt.Sprintf("redis://:%v@%v:%v/%v", env["REDIS_PASSWORD"], env["REDIS_HOST"], env["REDIS_PORT"], env["REDIS_DB"]),
	}
}
