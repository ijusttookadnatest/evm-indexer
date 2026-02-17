package ports

import "github/ijusttookadnatest/indexer-evm/core/domain"

type QueryRepository interface {
	
	GetByHash(hash string) (*domain.Block,error)
	GetById(id uint64) (*domain.Block,error)
	GetByRangeId(from, to uint64) ([]domain.Block,error)
	GetByRangeTime(from, to uint64) ([]domain.Block,error)
	
	GetByTransactionFilter(filter domain.TransactionFilter) ([]domain.Transaction,error)

	GetByEventFilter(filter domain.EventFilter) ([]domain.Event,error)
	GetByTxHashLogIndex(hash string, logIndex int) (*domain.Event,error)
}

type IndexerRepository interface {
	Create(block domain.Block, txs []domain.Transaction, events []domain.Event) error
	Delete(blockId int) error
}

