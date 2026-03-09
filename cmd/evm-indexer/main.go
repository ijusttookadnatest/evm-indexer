package main

import (
	"fmt"
	"net/http"
	"os"
	
	"github/ijusttookadnatest/indexer-evm/internal/config"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
	service "github/ijusttookadnatest/indexer-evm/internal/core/services"
	"github/ijusttookadnatest/indexer-evm/internal/fetcher"
	"github/ijusttookadnatest/indexer-evm/internal/handlers/graphql"
	"github/ijusttookadnatest/indexer-evm/internal/handlers/rest"
	"github/ijusttookadnatest/indexer-evm/internal/handlers/ws"
	repository "github/ijusttookadnatest/indexer-evm/internal/repository/db"
	"github/ijusttookadnatest/indexer-evm/internal/server"
)

func run() error {
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}
	
	indexerStreams := domain.IndexerStreams {
		Block: make(chan any, 10),
		Txs: make(chan any, 10),
		Events: make(chan any, 10),
	}
	
	db, err := repository.New(cfg.PostgresDSN)
	if err != nil {
		return err
	}
	
	indexerRepo := repository.NewIndexerRepository(db)
	fetcher, err := fetcher.NewFetcher(cfg.Rpc)
	if err != nil {
		return err
	}
	indexerService := service.NewIndexerService(indexerRepo, fetcher, indexerStreams)
	indexerService.Run(cfg.From, cfg.ConcurrencyF)
	
	queryRepo := repository.NewQueryRepository(db)
	queryService := service.NewQueryService(queryRepo, cfg.OffsetMax, cfg.RangeMaxTime)
	handlers := []http.Handler{
		ws.NewRouter(indexerStreams),
		rest.NewRouter(queryService),
		graphql.NewRouter(queryService, cfg.PlaygroundEnabled),
	}
	server := server.NewHTTPServer(handlers, cfg.Port)
	server.Run()
	return nil
}

func main() {
	if err := run() ; err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
