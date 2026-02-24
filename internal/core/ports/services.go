package ports

import "github/ijusttookadnatest/indexer-evm/internal/core/domain"

type QueryService interface {
	GetBlockByHash(hash string, tx bool) (*domain.BlockTxs, error)
	GetBlockById(id uint64, tx bool) (*domain.BlockTxs, error)
	GetBlocksWithOffset(fromId, offset uint64, tx bool) ([]domain.BlockTxs, error)
	GetBlocksByRangeTime(from, to uint64, tx bool) ([]domain.BlockTxs, error)

	GetTransactionsByFilter(filter domain.TransactionFilter) ([]domain.Transaction, error)
	GetTransactionsByBatchBlocksId(blockIDs []uint64, tx bool) (map[uint64][]domain.Transaction, error)

	GetEventsByFilter(filter domain.EventFilter) ([]domain.Event, error)
	GetEventByTxHashLogIndex(hash string, logIndex int) (*domain.Event, error)
	GetEventsByBatchTxsHash(txsHash []string) (map[string][]domain.Event, error)
}

type IndexerService interface {
	Create(block *domain.Block, txs []domain.Transaction, events []domain.Event) error
	Delete(blockId int) error
}

type BackfillingService interface {
	Run(from, to uint64) error
}


