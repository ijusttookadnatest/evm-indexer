package ports

import "github/ijusttookadnatest/indexer-evm/core/domain"

type BlockService interface {
	Create(block *domain.Block, txs []domain.Transaction, events []domain.Event) error

	GetByHash(hash string, tx bool) (*domain.BlockTxs,error)
	GetById(id uint64, tx bool) (*domain.BlockTxs,error)
	GetByRangeId(from, to uint64, tx bool) ([]domain.BlockTxs,error)
	GetByRangeTime(from, to uint64, tx bool) ([]domain.BlockTxs,error)

	Delete(blockId int) error
}

type TransactionService interface {
	GetByTransactionFilter(filter domain.TransactionFilter) ([]domain.Transaction,error)
}

type EventService interface {
	GetByEventFilter(filter domain.EventFilter) ([]domain.Event,error)
	GetByTxHashLogIndex(hash string, logIndex int) (*domain.Event,error)
}