package service

import (
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
)

type TransactionService struct {
	txRepo ports.TransactionRepository
	rangeMax uint
}

func NewTransactionService(txRepo ports.TransactionRepository, rangeMax uint) *TransactionService {
	return &TransactionService{
		txRepo: txRepo,
		rangeMax: rangeMax,
	}
}

func (service *TransactionService) GetByTransactionFilter(filter domain.TransactionFilter) ([]domain.Transaction,error) {
	if filter.Hash == nil && filter.BlockId == nil && filter.FromBlock == nil && filter.From == nil && filter.To == nil {
		return nil, domain.ErrEmptyFilter
	}

	if filter.BlockId != nil && *filter.BlockId == 0 {
		return nil, domain.ErrInvalidId
	}
	if filter.FromBlock != nil || filter.ToBlock != nil {
		if (filter.FromBlock != nil && filter.ToBlock == nil) || (filter.ToBlock != nil && filter.FromBlock == nil) {
			return nil, domain.ErrInvalidBlockRange
		}
		if err := domain.ValidateBlockRange(*filter.FromBlock, *filter.ToBlock, service.rangeMax) ; err != nil {
			return nil, err
		}
	}
	if filter.Limit != nil && *filter.Limit <= 0 {
		return nil, domain.ErrInvalidLimit
	}
	if filter.Hash != nil {
		if err := domain.ParseHash(*filter.Hash) ; err != nil {
			return nil, err
		}
	}
	if filter.From != nil {
		if err := domain.ParseAddress(*filter.From) ; err != nil {
			return nil, err
		}
	}
	if filter.To != nil {
		if err := domain.ParseAddress(*filter.To) ; err != nil {
			return nil, err
		}
	}
	tsx, err := service.txRepo.GetByTransactionFilter(filter)
	if err != nil {
		return nil, err
	}
	return tsx, nil
}
