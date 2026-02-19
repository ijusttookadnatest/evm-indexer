package main

import (
	"github/ijusttookadnatest/indexer-evm/config"
	service "github/ijusttookadnatest/indexer-evm/internal/core/services"
	repository "github/ijusttookadnatest/indexer-evm/internal/repository/db"
	repository "github/ijusttookadnatest/indexer-evm/repository/db"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		return
	}
	db, err := ConnectDatabase(cfg.PostgresDSN)
	if err != nil {
		return
	}
	queryRepo := repository.NewQueryRepository(db)
	queryService := service.NewQueryService(queryRepo)

}
