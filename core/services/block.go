package service

import repository "github/ijusttookadnatest/indexer-evm/repository/db"

type BlockService struct {
	blockRepo *repository.BlockRepository
}

func NewBlockService(blockRepo *repository.BlockRepository) *BlockService {
	return &BlockService{blockRepo: blockRepo}
}

