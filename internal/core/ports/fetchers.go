package ports

import (
	"context"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
)

type Fetcher interface {
	FetchBlock(id uint64) (domain.BlockTxsEvents, error)
	GetLastBlockId() (uint64,error)
	Subscribe(ctx context.Context, c chan<- uint64) error
}