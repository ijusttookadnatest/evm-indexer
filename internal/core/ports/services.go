package ports

import (
	"context"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
)

type QueryService interface {
	GetBlockByHash(ctx context.Context, hash string, tx bool) (*domain.BlockTxs, error)
	GetBlockById(ctx context.Context, id uint64, tx bool) (*domain.BlockTxs, error)
	GetBlocksWithOffset(ctx context.Context, fromId, offset uint64, tx bool) ([]domain.BlockTxs, error)
	GetBlocksByRangeTime(ctx context.Context, from, to uint64, tx bool) ([]domain.BlockTxs, error)

	GetTransactionsByFilter(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error)
	GetTransactionsByBatchBlocksId(ctx context.Context, blockIDs []uint64, tx bool) (map[uint64][]domain.Transaction, error)

	GetEventsByFilter(ctx context.Context, filter domain.EventFilter) ([]domain.Event, error)
	GetEventByTxHashLogIndex(ctx context.Context, hash string, logIndex int) (*domain.Event, error)
	GetEventsByBatchTxsHash(ctx context.Context, txsHash []string) (map[string][]domain.Event, error)
}

type IndexerService interface {
	Backfill(from uint64, concurrencyF int) error
	ForwardFill() error
	// Delete(blockId int) error
}



