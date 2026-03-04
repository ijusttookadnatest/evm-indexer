package ports

import "github/ijusttookadnatest/indexer-evm/internal/core/domain"

type Backfiller interface {
	FetchBlock(id uint64) (domain.BlockTxsEvents, error)
	GetLastBlockId() (uint64,error)
}

