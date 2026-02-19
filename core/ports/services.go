package ports

import "github/ijusttookadnatest/indexer-evm/core/domain"

type QueryService interface {
	GetBlockByHash(hash string, tx bool) (*domain.BlockTxs,error)
	GetBlockById(id uint64, tx bool) (*domain.BlockTxs,error)
	GetBlocksByRangeId(from, to uint64, tx bool) ([]domain.BlockTxs,error)
	GetBlocksByRangeTime(from, to uint64, tx bool) ([]domain.BlockTxs,error)
	
	GetTransactionByFilter(filter domain.TransactionFilter) ([]domain.Transaction,error)
	GetTransactionsByBatchBlocksId(blockIDs []uint64, tx bool) (map[uint64][]domain.Transaction,error)

	GetEventByFilter(filter domain.EventFilter) ([]domain.Event,error)
	GetEventByTxHashLogIndex(hash string, logIndex int) (*domain.Event,error)
	GetEventsByBatchTxsHash(txsHash []string) (map[string][]domain.Event,error)
}

type IndexerService interface {
	Create(block *domain.Block, txs []domain.Transaction, events []domain.Event) error
	Delete(blockId int) error
}
