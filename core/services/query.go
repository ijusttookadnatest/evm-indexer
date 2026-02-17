package service

import (
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
)

type QueryService struct {
	repo ports.QueryRepository
	rangeMaxId uint64
	rangeMaxTime uint64
}

func NewQueryService(repo ports.QueryRepository, rangeMaxId uint64, rangeMaxTime uint64) *QueryService {
	return &QueryService{ 
		repo: repo, 
		rangeMaxId: rangeMaxId,
		rangeMaxTime: rangeMaxTime,
	}
}

func (service *QueryService) GetByHash(hash string, tx bool) (*domain.BlockTxs,error) {
	if err := domain.ParseHash(hash) ; err != nil {
		return nil, err
	}
	blockData, err := service.repo.GetByHash(hash)
	if err != nil {
		return nil, err
	}

	var txs []domain.Transaction
	if tx {
		filter := domain.TransactionFilter{BlockId: &blockData.Id}
		txs, err = service.repo.GetByTransactionFilter(filter)
		if err != nil {
			return nil, err
		}
	}
	return &domain.BlockTxs{
		Block: *blockData,
		Txs: txs,
	}, nil
}

func (service *QueryService) GetById(id uint64, tx bool) (*domain.BlockTxs,error) {
	if id == 0 {
		return nil, domain.ErrInvalidId
	}
	blockData, err := service.repo.GetById(id)
	if err != nil {
		return nil, err
	}

	var txs []domain.Transaction
	if tx {
		filter := domain.TransactionFilter{BlockId: &blockData.Id}
		txs, err = service.repo.GetByTransactionFilter(filter)
		if err != nil {
			return nil, err
		}
	}
	return &domain.BlockTxs{
		Block: *blockData,
		Txs: txs,
	}, nil
}

func (service *QueryService) GetByRangeId(from, to uint64, tx bool) ([]domain.BlockTxs,error) {
	if to == 0 {
		return nil, domain.ErrInvalidId
	}
	if from >= to {
		return nil, domain.ErrInvalidId
	}
	if err := domain.ValidateBlockRange(from, to, service.rangeMaxId) ; err != nil {
		return nil, domain.ErrInvalidId
	}

	blocksData, err := service.repo.GetByRangeId(from, to)
	if err != nil {
		return nil, err
	}

	var blocks = make([]domain.BlockTxs, len(blocksData))

	for i, blockData := range blocksData {
		var txs []domain.Transaction
		if (tx) {
			filter := domain.TransactionFilter{BlockId: &blockData.Id}
			txs, err = service.repo.GetByTransactionFilter(filter)
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

func (service *QueryService) GetByRangeTime(from, to uint64, tx bool) ([]domain.BlockTxs,error) {
	if from == 0 || to == 0{
		return nil, domain.ErrInvalidId
	}
	if from >= to {
		return nil, domain.ErrInvalidId
	}
	if err := domain.ValidateBlockRange(from, to, service.rangeMaxTime) ; err != nil {
		return nil, domain.ErrInvalidId
	}

	blocksData, err := service.repo.GetByRangeTime(from, to)
	if err != nil {
		return nil, err
	}

	var blocks = make([]domain.BlockTxs, len(blocksData))
	
	for i, blockData := range blocksData {
		var txs []domain.Transaction
		if (tx) {
			filter := domain.TransactionFilter{BlockId: &blockData.Id}
			txs, err = service.repo.GetByTransactionFilter(filter)
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


func (service *QueryService) GetByEventFilter(filter domain.EventFilter) ([]domain.Event,error) {
	if filter.Emitter == nil && filter.TxHash == nil && filter.FromBlock == nil && len(filter.Topics) == 0 && filter.Limit == nil && filter.ToBlock == nil {
		return nil, domain.ErrEmptyFilter
	}
	if filter.FromBlock != nil || filter.ToBlock != nil {
		if (filter.FromBlock != nil && filter.ToBlock == nil) || (filter.ToBlock != nil && filter.FromBlock == nil) {
			return nil, domain.ErrInvalidBlockRange
		}
		if err := domain.ValidateBlockRange(*filter.FromBlock, *filter.ToBlock, service.rangeMaxId) ; err != nil {
			return nil, err
		}
	}
	if filter.Limit != nil && *filter.Limit <= 0 {
		return nil, domain.ErrInvalidLimit
	}
	if filter.TxHash != nil {
		if err := domain.ParseHash(*filter.TxHash) ; err != nil {
			return nil, err
		}
	}
	if filter.Emitter != nil {
		if err := domain.ParseAddress(*filter.Emitter) ; err != nil {
			return nil, err
		}
	}
	if len(filter.Topics) > 0 {
		if err := domain.ParseTopics(filter.Topics) ; err != nil {
			return nil, err
		}
	}
	events, err := service.repo.GetByEventFilter(filter)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (service *QueryService) GetByTxHashLogIndex(hash string, logIndex int) (*domain.Event,error) {
	if err := domain.ParseHash(hash) ; err != nil {
		return nil, err
	}
	if logIndex < 0 {
		return nil, domain.ErrInvalidId
	}
	event, err := service.repo.GetByTxHashLogIndex(hash, logIndex)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (service *QueryService) GetByTransactionFilter(filter domain.TransactionFilter) ([]domain.Transaction,error) {
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
		if err := domain.ValidateBlockRange(*filter.FromBlock, *filter.ToBlock, service.rangeMaxId) ; err != nil {
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
	tsx, err := service.repo.GetByTransactionFilter(filter)
	if err != nil {
		return nil, err
	}
	return tsx, nil
}
