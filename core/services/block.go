package service

import (
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
)

type BlockService struct {
	blockRepo ports.BlockRepository
	txRepo ports.TransactionRepository
}

func NewBlockService(blockRepo ports.BlockRepository, txRepo ports.TransactionRepository) *BlockService {
	return &BlockService{
		blockRepo: blockRepo,
		txRepo: txRepo,
	}
}

func (service *BlockService) Create(block *domain.Block, txs []domain.Transaction, events []domain.Event) error {
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
	err := service.blockRepo.Create(*block, txs, events)
	if err != nil {
		return err
	}
	return nil
}

func (service *BlockService) GetByHash(hash string, tx bool) (*domain.BlockTxs,error) {
	if err := domain.ParseHash(hash) ; err != nil {
		return nil, err
	}
	blockData, err := service.blockRepo.GetByHash(hash)
	if err != nil {
		return nil, err
	}

	var txs []domain.Transaction
	if tx {
		filter := domain.TransactionFilter{BlockId: &blockData.Id}
		txs, err = service.txRepo.GetByTransactionFilter(filter)
		if err != nil {
			return nil, err
		}
	}
	return &domain.BlockTxs{
		Block: *blockData,
		Txs: txs,
	}, nil
}

func (service *BlockService) GetById(id uint64, tx bool) (*domain.BlockTxs,error) {
	if id == 0 {
		return nil, domain.ErrInvalidId
	}
	blockData, err := service.blockRepo.GetById(id)
	if err != nil {
		return nil, err
	}

	var txs []domain.Transaction
	if tx {
		filter := domain.TransactionFilter{BlockId: &blockData.Id}
		txs, err = service.txRepo.GetByTransactionFilter(filter)
		if err != nil {
			return nil, err
		}
	}
	return &domain.BlockTxs{
		Block: *blockData,
		Txs: txs,
	}, nil
}

func (service *BlockService) GetByRangeId(from, to uint64, tx bool) ([]domain.BlockTxs,error) {
	if from == 0 || to == 0{
		return nil, domain.ErrInvalidId
	}
	if from >= to {
		return nil, domain.ErrInvalidId
	}

	blocksData, err := service.blockRepo.GetByRangeId(from, to)
	if err != nil {
		return nil, err
	}

	var blocks = make([]domain.BlockTxs, len(blocksData))

	for i, blockData := range blocksData {
		var txs []domain.Transaction
		if (tx) {
			filter := domain.TransactionFilter{BlockId: &blockData.Id}
			txs, err = service.txRepo.GetByTransactionFilter(filter)
			if err != nil {
				return nil, err
			}
		}
		blocks[i] = domain.BlockTxs{
			Block: blockData,
			Txs: txs,
		}
	}
	return blocks, nil
}

func (service *BlockService) GetByRangeTime(from, to uint64, tx bool) ([]domain.BlockTxs,error) {
	if from == 0 || to == 0{
		return nil, domain.ErrInvalidId
	}
	if from >= to {
		return nil, domain.ErrInvalidId
	}

	blocksData, err := service.blockRepo.GetByRangeTime(from, to)
	if err != nil {
		return nil, err
	}

	var blocks = make([]domain.BlockTxs, len(blocksData))
	
	for i, blockData := range blocksData {
		var txs []domain.Transaction
		if (tx) {
			filter := domain.TransactionFilter{BlockId: &blockData.Id}
			txs, err = service.txRepo.GetByTransactionFilter(filter)
			if err != nil {
				return nil, err
			}
		}
		blocks[i] = domain.BlockTxs{
			Block: blockData,
			Txs: txs,
		}
	}
	return blocks, nil
}

func (service *BlockService) Delete(blockId int) error {
	if blockId <= 0 {
		return domain.ErrInvalidId
	}
	if err := service.blockRepo.Delete(blockId) ; err != nil {
		return err
	}
	return nil
}
