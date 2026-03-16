package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github/ijusttookadnatest/evm-indexer/internal/config"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	service "github/ijusttookadnatest/evm-indexer/internal/core/services"
	repository "github/ijusttookadnatest/evm-indexer/internal/repository/db"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/graphql"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/rest"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/ws"
	"github/ijusttookadnatest/evm-indexer/internal/server"

	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context) error {
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}
	
	g, ctx := errgroup.WithContext(ctx)
	indexerStreams := domain.IndexerStreams{
		Block:  make(chan any, 10),
		Txs:    make(chan any, 10),
		Events: make(chan any, 10),
	}

	db, err := repository.New(cfg.PostgresDSN)
	if err != nil {
		return err
	}
	if err := repository.RunUpMigrations(db) ; err != nil {
		return err
	}

	queryRepo := repository.NewQueryRepository(db)
	queryService := service.NewQueryService(queryRepo, cfg.OffsetMax, cfg.RangeMaxTime)
	
	wsHandler := ws.NewRouter(ctx, indexerStreams)
	restHandler := rest.NewRouter(queryService)
	graphqHandler := graphql.NewRouter(queryService, cfg.PlaygroundEnabled)
	server := server.NewHTTPServer(restHandler, wsHandler, graphqHandler, cfg.Port)

	g.Go(func() error {
		return server.Run(ctx)
	})

	slog.Info("server running", "port", cfg.Port)

	return g.Wait()
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
