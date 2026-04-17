package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github/ijusttookadnatest/evm-indexer/internal/config"
	service "github/ijusttookadnatest/evm-indexer/internal/core/services"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/graphql"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/rest"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/ws"
	"github/ijusttookadnatest/evm-indexer/internal/prometheus"
	"github/ijusttookadnatest/evm-indexer/internal/pubsub"
	repository "github/ijusttookadnatest/evm-indexer/internal/repository/db"
	"github/ijusttookadnatest/evm-indexer/internal/server"

	"golang.org/x/sync/errgroup"
)

func run(ctx context.Context) error {
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}
	fmt.Println("DEBUG DSN:", cfg.PostgresDSN)

	g, ctx := errgroup.WithContext(ctx)

	db, err := repository.New(cfg.PostgresDSN)
	if err != nil {
		return err
	}
	if err := repository.RunUpMigrations(db); err != nil {
		return err
	}

	redis, err := pubsub.New(cfg.RedisDSN)
	if err != nil {
		return err
	}

	reg := prometheus.NewRegistry()
	metrics := prometheus.NewApiMetrics(reg)
	prometheusServer := prometheus.NewPrometheusServer(reg, "2113")
	go prometheus.RunPrometheusServer(ctx, prometheusServer)

	pubsub := pubsub.NewRedisPubSub(redis)
	queryRepo := repository.NewQueryRepository(db)
	queryService := service.NewQueryService(queryRepo, cfg.OffsetMax, cfg.RangeMaxTime)

	wsHandler, err := ws.NewRouter(ctx, pubsub, metrics)
	if err != nil {
		return err
	}
	restHandler := rest.NewRouter(queryService, metrics)
	graphqHandler := graphql.NewRouter(queryService, cfg.PlaygroundEnabled, metrics)
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
