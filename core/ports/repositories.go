package ports

import "github/ijusttookadnatest/indexer-evm/core/domain"

type BlockRepository interface {
	Create(block *domain.Block) error

	GetByHash(hash string) (*domain.Block,error)
	GetByNumber(number uint64) (*domain.Block,error)
	GetByRangeNumber(from, to uint64) ([]domain.Block,error)
	GetByRangeTime(from, to uint64) ([]domain.Block,error)

	Delete(blockId int) error
}

type TransactionRepository interface {
	GetByTransactionFilter(filter domain.TransactionFilter) ([]*domain.Transaction,error)
}

type EventRepository interface {
	GetByEventFilter(filter domain.EventFilter) ([]*domain.Event,error)
	GetByTxHashLogIndex(hash string, logIndex int) (*domain.Event,error)
}
