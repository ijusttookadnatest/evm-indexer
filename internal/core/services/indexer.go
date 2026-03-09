package service

import (
	"context"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
	"github/ijusttookadnatest/indexer-evm/internal/core/ports"

	"golang.org/x/sync/errgroup"
)

type IndexerService struct {
	repo ports.IndexerRepository
	fetcher ports.Fetcher
	indexerStreams domain.IndexerStreams
}

func NewIndexerService(repo ports.IndexerRepository, fetcher ports.Fetcher, indexerStreams domain.IndexerStreams) *IndexerService {
	return &IndexerService{repo: repo, fetcher:fetcher, indexerStreams:indexerStreams}
}

func (i *IndexerService) Run(from uint64, concurrencyF int) error {
	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(parentCtx)
	g.Go(func() error {
		return i.forwardfill(ctx)
	})
	targetId, err := i.fetcher.GetLastBlockId()
	if err != nil {
		cancel()
		g.Wait()
		return err
	}
	g.Go(func() error {
		return i.backfill(ctx, from, targetId, concurrencyF)
	})

	return g.Wait()
}