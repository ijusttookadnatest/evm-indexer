package ports

import "github/ijusttookadnatest/indexer-evm/internal/core/domain"

type QueryRepository interface {
	GetBlockByHash(hash string) (*domain.Block, error)
	GetBlockById(id uint64) (*domain.Block, error)
	GetBlocksByRangeId(from, to uint64) ([]domain.Block, error)
	GetBlocksByRangeTime(from, to uint64) ([]domain.Block, error)

	GetTransactionsByFilter(filter domain.TransactionFilter) ([]domain.Transaction, error)
	GetTransactionsByBatchBlocksId(blocksId []uint64) ([]domain.Transaction, error)

	GetEventsByFilter(filter domain.EventFilter) ([]domain.Event, error)
	GetEventByTxHashLogIndex(hash string, logIndex int) (*domain.Event, error)
	GetEventsByBatchTxsHash(txsHash []string) ([]domain.Event, error)
}

type IndexerRepository interface {
	Create(block domain.Block, txs []domain.Transaction, events []domain.Event) error
	Delete(blockId int) error
	GetBackfillCursor() (uint64, error)
	UpdateBackfillCursor(blockId uint64) error
}
