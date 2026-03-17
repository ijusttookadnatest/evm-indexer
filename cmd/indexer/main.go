package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github/ijusttookadnatest/evm-indexer/internal/config"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	service "github/ijusttookadnatest/evm-indexer/internal/core/services"
	"github/ijusttookadnatest/evm-indexer/internal/fetcher"
	"github/ijusttookadnatest/evm-indexer/internal/pubsub"
	repository "github/ijusttookadnatest/evm-indexer/internal/repository/db"

	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context, reindex bool) error {
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
	redis, err := pubsub.New(cfg.RedisDSN)
	if err != nil {
		return err
	}

	pubsub := pubsub.NewRedisPubSub(redis)
	indexerRepo := repository.NewIndexerRepository(db)

	if reindex {
		if err := indexerRepo.ResetBackfillCursor(); err != nil {
			return err
		}
	}

	fetcher, err := fetcher.NewFetcher(cfg.RpcHTTP, cfg.RpcWS, cfg.RpcRateLimit)
	if err != nil {
		return err
	}
	indexerService := service.NewIndexerService(indexerRepo, fetcher, pubsub)

	g.Go(func() error {
		return indexerService.Run(ctx, cfg.From, cfg.ConcurrencyF)
	})

	return g.Wait()
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	reindex := flag.Bool("reindex", false, "reset backfill cursor to re-index from scratch")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, *reindex); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
