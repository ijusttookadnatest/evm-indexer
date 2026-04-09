package ports

import (
	"context"

	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
)

type QueryRepository interface {
	GetBlockByHash(ctx context.Context, hash string) (*domain.Block, error)
	GetBlockById(ctx context.Context, id uint64) (*domain.Block, error)
	GetBlocksByRangeId(ctx context.Context, from, to uint64) ([]domain.Block, error)
	GetBlocksByRangeTime(ctx context.Context, from, to uint64) ([]domain.Block, error)

	GetTransactionsByFilter(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error)
	GetTransactionsByBatchBlocksId(ctx context.Context, blocksId []uint64) ([]domain.Transaction, error)

	GetEventsByFilter(ctx context.Context, filter domain.EventFilter) ([]domain.Event, error)
	GetEventByTxHashLogIndex(ctx context.Context, hash string, logIndex int) (*domain.Event, error)
	GetEventsByBatchTxsHash(ctx context.Context, txsHash []string) ([]domain.Event, error)
}

type IndexerRepository interface {
	Create(ctx context.Context, block domain.Block, txs []domain.Transaction, events []domain.Event) error
	BulkCreate(ctx context.Context, items []domain.BlockTxsEvents) error
	Delete(ctx context.Context, blockId uint64) error
	GetBlockById(ctx context.Context, id uint64) (*domain.Block, error)
	GetMaxIndexedBlock(ctx context.Context) (uint64, error)
	GetBackfillCursor(ctx context.Context) (uint64, error)
	UpdateBackfillCursor(ctx context.Context, blockId uint64) error
	ResetBackfillCursor(ctx context.Context) error
	GetBalancefillCursor(ctx context.Context) (uint64, error)
	UpdateBalancefillCursor(ctx context.Context, blockId uint64) error
	ResetBalancefillCursor(ctx context.Context) error
	BatchUpsertBalance(ctx context.Context, entries []domain.BalanceEntry) error
	GetLogsByTopic(ctx context.Context, filter domain.LogFilter) ([]domain.Log, error)
}
