package main

import (
	"github/ijusttookadnatest/indexer-evm/internal/config"
	service "github/ijusttookadnatest/indexer-evm/internal/core/services"
	"github/ijusttookadnatest/indexer-evm/internal/handlers/graphql"
	"github/ijusttookadnatest/indexer-evm/internal/handlers/rest"
	"github/ijusttookadnatest/indexer-evm/internal/handlers/ws"
	repository "github/ijusttookadnatest/indexer-evm/internal/repository/db"
	"github/ijusttookadnatest/indexer-evm/internal/server"
	"net/http"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		return
	}

	db, err := repository.New(cfg.PostgresDSN)
	if err != nil {
		return
	}

	queryRepo := repository.NewQueryRepository(db)
	queryService := service.NewQueryService(queryRepo, cfg.OffsetMax, cfg.RangeMaxTime)
	handlers := []http.Handler{
		ws.NewRouter(),
		rest.NewRouter(queryService),
		graphql.NewRouter(queryService, cfg.PlaygroundEnabled),
	}

	server := server.NewHTTPServer(handlers, cfg.Port)

	server.Run()
}
