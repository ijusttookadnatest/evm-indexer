package main

import (
	// "context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github/ijusttookadnatest/indexer-evm/config"
	service "github/ijusttookadnatest/indexer-evm/core/services"
	"github/ijusttookadnatest/indexer-evm/db"
	"github/ijusttookadnatest/indexer-evm/listener"
	"github/ijusttookadnatest/indexer-evm/processor"
	"github/ijusttookadnatest/indexer-evm/redis"
	repository "github/ijusttookadnatest/indexer-evm/repository/db"

	"github.com/joho/godotenv"
)

func main() {
	cfg , err := config.Load(".env")
	if err != nil {
		return
	}
	db, err := ConnectDatabase(cfg.PostgresDSN)
	if err != nil {
		return
	}
	repo := repository.NewQueryRepo(db)
	service := service.NewQueryService(repo)
	server := NewServer(cfg.ListenAddr, service)

	if err := server.server.ListenAndServe(); err != nil {
	 log.Fatalf("Server error: %v", err)
	}
}