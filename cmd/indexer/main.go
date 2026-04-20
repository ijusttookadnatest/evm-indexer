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
	service "github/ijusttookadnatest/evm-indexer/internal/core/services"
	"github/ijusttookadnatest/evm-indexer/internal/fetcher"
	"github/ijusttookadnatest/evm-indexer/internal/prometheus"
	"github/ijusttookadnatest/evm-indexer/internal/pubsub"
	repository "github/ijusttookadnatest/evm-indexer/internal/repository/db"
)

func run(ctx context.Context, reindex bool) error {
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}

	db, err := repository.New(cfg.PostgresDSN)
	if err != nil {
		return err
	}

	redis, err := pubsub.New(cfg.RedisDSN)
	if err != nil {
		return err
	}

	reg := prometheus.NewRegistry()
	metrics := prometheus.NewIndexerMetrics(reg)
	prometheusServer := prometheus.NewPrometheusServer(reg, "2112")
	go prometheus.RunPrometheusServer(ctx, prometheusServer)

	pubsub := pubsub.NewRedisPubSub(redis)
	indexerRepo := repository.NewIndexerRepository(db)

	if reindex {
		if err := indexerRepo.ResetBackfillCursor(ctx); err != nil {
			return err
		}
	}

	fetcher, err := fetcher.NewFetcher(cfg.RpcHTTP, cfg.RpcWS, cfg.RpcRateLimit)
	if err != nil {
		return err
	}
	indexerService := service.NewIndexerService(indexerRepo, fetcher, pubsub, metrics)

	return indexerService.Run(ctx, cfg.From, cfg.ConcurrencyF)
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
