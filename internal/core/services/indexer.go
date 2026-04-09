package service

import (
	"context"
	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	"github/ijusttookadnatest/evm-indexer/internal/prometheus"

	"golang.org/x/sync/errgroup"
)

type IndexerService struct {
	repo           ports.IndexerRepository
	fetcher        ports.Fetcher
	pubsub		   ports.RedisPubSub
	metrics 	   *prometheus.IndexerMetrics
}

func NewIndexerService(repo ports.IndexerRepository, fetcher ports.Fetcher, pubsub ports.RedisPubSub, metrics *prometheus.IndexerMetrics) *IndexerService {
	return &IndexerService{repo:repo, fetcher:fetcher, pubsub:pubsub, metrics:metrics}
}

func (i *IndexerService) Run(ctx context.Context, from uint64, concurrencyF int) error {
	parentCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(parentCtx)
	backfillChan := make(chan struct{}, 1)

	g.Go(func() error {
		i.metrics.ForwardfillIsSyncing.Inc()
		err := i.forwardfill(ctx)
		if err != nil {
			i.metrics.ForwardfillError.Inc()
		}
		i.metrics.ForwardfillIsSyncing.Dec()
		return err
	})
	targetId, err := i.fetcher.GetLastBlockId()
	if err != nil {
		cancel()
		g.Wait()
		return err
	}
	g.Go(func() error {
		i.metrics.BackfillIsSyncing.Inc()
		err := i.backfill(ctx, from, targetId, concurrencyF, backfillChan)
		if err != nil {
			i.metrics.BackfillError.Inc()
		}
		i.metrics.BackfillIsSyncing.Dec()
		return err
	})

	g.Go(func() error {
		select {
		case <-backfillChan: {
			i.metrics.BalancefillIsSyncing.Inc()
			err := i.balancefill(ctx, 1000, 96)
			if err != nil {
				i.metrics.BalancefillError.Inc()
			}
			i.metrics.BalancefillIsSyncing.Dec()
			return err
		}
		case <-ctx.Done(): return ctx.Err()
		}
	})
	
	return g.Wait()
}
