package ports

import (
	"context"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
)

type Fetcher interface {
	FetchBlock(ctx context.Context, id uint64) (domain.BlockTxsEvents, error)
	FetchBlockPriority(ctx context.Context, id uint64) (domain.BlockTxsEvents, error)
	GetLastBlockId() (uint64, error)
	Subscribe(ctx context.Context, c chan<- uint64) error
}
