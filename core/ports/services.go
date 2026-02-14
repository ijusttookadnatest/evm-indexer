package ports

import "github/ijusttookadnatest/indexer-evm/core/domain"

type BlockService interface {
	Create(block *domain.Block) error

	GetByHash(hash string, tx bool) (*domain.Block,error)
	GetByNumber(number uint64, tx bool) (*domain.Block,error)
	GetByRangeNumber(from, to uint64, tx bool) ([]domain.Block,error)
	GetByRangeTime(from, to uint64, tx bool) ([]domain.Block,error)

	Delete(blockId int) error
}

type TransactionService interface {
	GetByTransactionFilter(filter domain.TransactionFilter) ([]*domain.Transaction,error)
}

type EventService interface {
	GetByEventFilter(filter domain.EventFilter) ([]*domain.Event,error)
	GetByTxHashLogIndex(hash string, logIndex int) (*domain.Event,error)
}