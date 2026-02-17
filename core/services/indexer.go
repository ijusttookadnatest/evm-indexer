package service

import (
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
)

type IndexerService struct {
	repo ports.IndexerRepository
}

func NewIndexerService(repo ports.IndexerRepository) *IndexerService {
	return &IndexerService{ repo: repo }
}

func (service *IndexerService) Create(block *domain.Block, txs []domain.Transaction, events []domain.Event) error {
	if err := domain.ParseBlock(*block) ; err != nil {
		return err
	}
	for _, tx := range txs {
		if err := domain.ParseTx(tx) ; err != nil {
			return err
		}
	}
	for _, event := range events {
		if err := domain.ParseEvent(event) ; err != nil {
			return err
		}
	}
	err := service.repo.Create(*block, txs, events)
	if err != nil {
		return err
	}
	return nil
}


func (service *IndexerService) Delete(blockId int) error {
	if blockId <= 0 {
		return domain.ErrInvalidId
	}
	if err := service.repo.Delete(blockId) ; err != nil {
		return err
	}
	return nil
}



