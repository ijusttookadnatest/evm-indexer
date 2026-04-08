package ports

import (
	"context"
	"math/big"

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
	Create(block domain.Block, txs []domain.Transaction, events []domain.Event) error
	BulkCreate(items []domain.BlockTxsEvents) error
	Delete(blockId uint64) error
	GetBlockById(ctx context.Context, id uint64) (*domain.Block, error)
	GetBackfillCursor() (uint64, error)
	UpdateBackfillCursor(blockId uint64) error
	ResetBackfillCursor() error
	GetBalancefillCursor() (uint64, error)
	UpdateBalancefillCursor(blockId uint64) error
	ResetBalancefillCursor() error
	UpsertBalance(from, to, token, tokenId string, amount big.Int) error
}
