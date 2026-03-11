package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github/ijusttookadnatest/evm-indexer/internal/config"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	service "github/ijusttookadnatest/evm-indexer/internal/core/services"
	"github/ijusttookadnatest/evm-indexer/internal/fetcher"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/graphql"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/rest"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/ws"
	repository "github/ijusttookadnatest/evm-indexer/internal/repository/db"
	"github/ijusttookadnatest/evm-indexer/internal/server"

	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context) error {
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}

	indexerStreams := domain.IndexerStreams{
		Block:  make(chan any, 10),
		Txs:    make(chan any, 10),
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

	queryRepo := repository.NewQueryRepository(db)
	queryService := service.NewQueryService(queryRepo, cfg.OffsetMax, cfg.RangeMaxTime)
	handlers := []http.Handler{
		ws.NewRouter(indexerStreams),
		rest.NewRouter(queryService),
		graphql.NewRouter(queryService, cfg.PlaygroundEnabled),
	}
	server := server.NewHTTPServer(handlers, cfg.Port)

	g, context := errgroup.WithContext(ctx)
	g.Go(func() error {
		return indexerService.Run(context, cfg.From, cfg.ConcurrencyF)
	})

	g.Go(func() error {
		return server.Run(context)
	})

	return g.Wait()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
