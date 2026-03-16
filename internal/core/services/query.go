package service

import (
	"context"

	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
)

var offsetDefault uint64 = 1000

type QueryService struct {
	repo         ports.QueryRepository
	rangeMaxTime uint64
	offsetMax    uint64
}

func NewQueryService(repo ports.QueryRepository, offsetMax uint64, rangeMaxTime uint64) *QueryService {
	return &QueryService{
		repo:         repo,
		rangeMaxTime: rangeMaxTime,
		offsetMax:    offsetMax,
	}
}

func (service *QueryService) GetBlockByHash(ctx context.Context, hash string, tx bool) (*domain.BlockTxs, error) {
	if err := domain.ParseHash(hash); err != nil {
		return nil, err
	}
	blockData, err := service.repo.GetBlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	var txs []domain.Transaction
	if tx {
		filter := domain.TransactionFilter{BlockId: &blockData.Id}
		txs, err = service.repo.GetTransactionsByFilter(ctx, filter)
		if err != nil {
			return nil, err
		}
	}
	return &domain.BlockTxs{
		Block: *blockData,
		Txs:   txs,
	}, nil
}

func (service *QueryService) GetBlockById(ctx context.Context, id uint64, tx bool) (*domain.BlockTxs, error) {
	if id == 0 {
		return nil, domain.ErrInvalidId
	}
	blockData, err := service.repo.GetBlockById(ctx, id)
	if err != nil {
		return nil, err
	}

	var txs []domain.Transaction
	if tx {
		filter := domain.TransactionFilter{BlockId: &blockData.Id}
		txs, err = service.repo.GetTransactionsByFilter(ctx, filter)
		if err != nil {
			return nil, err
		}
	}
	return &domain.BlockTxs{
		Block: *blockData,
		Txs:   txs,
	}, nil
}

func aggregateBlocksId(blocksData []domain.Block) []uint64 {
	blocksId := make([]uint64, len(blocksData))
	for i := range blocksData {
		blocksId[i] = blocksData[i].Id
	}
	return blocksId
}

func (service *QueryService) GetBlocksWithOffset(ctx context.Context, from, offset uint64, tx bool) ([]domain.BlockTxs, error) {
	if offset > service.offsetMax {
		return nil, domain.ErrInvalidOffset
	}
	if offset == 0 {
		offset = offsetDefault
	}
	blocksData, err := service.repo.GetBlocksByRangeId(ctx, from, from+offset)
	if err != nil {
		return nil, err
	}

	blockIDs := aggregateBlocksId(blocksData)
	mTxs, err := service.GetTransactionsByBatchBlocksId(ctx, blockIDs, tx)
	if err != nil {
		return nil, err
	}

	var blocks = make([]domain.BlockTxs, len(blocksData))
	for i, blockData := range blocksData {
		id := blockData.Id
		blocks[i] = domain.BlockTxs{
			Block: blockData,
			Txs:   mTxs[id],
		}
	}

	return blocks, nil
}

func (service *QueryService) GetTransactionsByBatchBlocksId(ctx context.Context, blockIDs []uint64, tx bool) (map[uint64][]domain.Transaction, error) {
	var txs []domain.Transaction
	var err error

	if tx {
		txs, err = service.repo.GetTransactionsByBatchBlocksId(ctx, blockIDs)
		if err != nil {
			return nil, err
		}
	}

	mTxs := make(map[uint64][]domain.Transaction, len(blockIDs))
	for _, tx := range txs {
		id := tx.BlockId
		mTxs[id] = append(mTxs[id], tx)
	}

	return mTxs, nil
}

func (service *QueryService) GetBlocksByRangeTime(ctx context.Context, from, to uint64, tx bool) ([]domain.BlockTxs, error) {
	if from == 0 || to == 0 {
		return nil, domain.ErrInvalidId
	}
	if from >= to {
		return nil, domain.ErrInvalidId
	}
	if err := domain.ValidateBlockRange(from, to, service.rangeMaxTime); err != nil {
		return nil, domain.ErrInvalidId
	}

	blocksData, err := service.repo.GetBlocksByRangeTime(ctx, from, to)
	if err != nil {
		return nil, err
	}

	blockIDs := aggregateBlocksId(blocksData)
	mTxs, err := service.GetTransactionsByBatchBlocksId(ctx, blockIDs, tx)
	if err != nil {
		return nil, err
	}

	var blocks = make([]domain.BlockTxs, len(blocksData))
	for i, blockData := range blocksData {
		id := blockData.Id
		blocks[i] = domain.BlockTxs{
			Block: blockData,
			Txs:   mTxs[id],
		}
	}
	return blocks, nil
}

func (service *QueryService) GetEventsByFilter(ctx context.Context, filter domain.EventFilter) ([]domain.Event, error) {
	if filter.Emitter == nil && filter.TxHash == nil && filter.FromBlock == nil && len(filter.Topics) == 0 && filter.Limit == nil && filter.ToBlock == nil {
		return nil, domain.ErrEmptyFilter
	}
	if filter.FromBlock != nil || filter.ToBlock != nil {
		if (filter.FromBlock != nil && filter.ToBlock == nil) || (filter.ToBlock != nil && filter.FromBlock == nil) {
			return nil, domain.ErrInvalidBlockRange
		}
		if err := domain.ValidateBlockRange(*filter.FromBlock, *filter.ToBlock, service.offsetMax); err != nil {
			return nil, err
		}
	}
	if filter.Limit != nil && (*filter.Limit <= 0 || *filter.Limit > int(service.offsetMax)) {
		return nil, domain.ErrInvalidLimit
	}
	if filter.TxHash != nil {
		if err := domain.ParseHash(*filter.TxHash); err != nil {
			return nil, err
		}
	}
	if filter.Emitter != nil {
		if err := domain.ParseAddress(*filter.Emitter); err != nil {
			return nil, err
		}
	}
	if len(filter.Topics) > 0 {
		if err := domain.ParseTopics(filter.Topics); err != nil {
			return nil, err
		}
	}

	if filter.Limit == nil {
		defaultLimit := int(offsetDefault)
		filter.Limit = &defaultLimit
	}
	events, err := service.repo.GetEventsByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (service *QueryService) GetEventByTxHashLogIndex(ctx context.Context, hash string, logIndex int) (*domain.Event, error) {
	if err := domain.ParseHash(hash); err != nil {
		return nil, err
	}
	if logIndex < 0 {
		return nil, domain.ErrInvalidId
	}
	event, err := service.repo.GetEventByTxHashLogIndex(ctx, hash, logIndex)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (service *QueryService) GetEventsByBatchTxsHash(ctx context.Context, txsHash []string) (map[string][]domain.Event, error) {
	var events []domain.Event
	var err error

	events, err = service.repo.GetEventsByBatchTxsHash(ctx, txsHash)
	if err != nil {
		return nil, err
	}

	mEvents := make(map[string][]domain.Event, len(txsHash))
	for _, event := range events {
		hash := event.TxHash
		mEvents[hash] = append(mEvents[hash], event)
	}

	return mEvents, nil
}

func (service *QueryService) GetTransactionsByFilter(ctx context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	if filter.Hash == nil && filter.BlockId == nil && filter.FromBlock == nil && filter.From == nil && filter.To == nil && filter.ToBlock == nil {
		return nil, domain.ErrEmptyFilter
	}

	if filter.BlockId != nil && *filter.BlockId == 0 {
		return nil, domain.ErrInvalidId
	}
	if filter.FromBlock != nil || filter.ToBlock != nil {
		if (filter.FromBlock != nil && filter.ToBlock == nil) || (filter.ToBlock != nil && filter.FromBlock == nil) {
			return nil, domain.ErrInvalidBlockRange
		}
		if err := domain.ValidateBlockRange(*filter.FromBlock, *filter.ToBlock, service.offsetMax); err != nil {
			return nil, err
		}
	}
	if filter.Limit != nil && (*filter.Limit <= 0 || *filter.Limit > int(service.offsetMax)) {
		return nil, domain.ErrInvalidLimit
	}
	if filter.Hash != nil {
		if err := domain.ParseHash(*filter.Hash); err != nil {
			return nil, err
		}
	}
	if filter.From != nil {
		if err := domain.ParseAddress(*filter.From); err != nil {
			return nil, err
		}
	}
	if filter.To != nil {
		if err := domain.ParseAddress(*filter.To); err != nil {
			return nil, err
		}
	}

	if filter.Limit == nil {
		defaultLimit := int(offsetDefault)
		filter.Limit = &defaultLimit
	}

	tsx, err := service.repo.GetTransactionsByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	return tsx, nil
}
