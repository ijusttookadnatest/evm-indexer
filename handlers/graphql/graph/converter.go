package graph

import (
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/handlers/graphql/graph/dto"
)

func toBlockDTO(b domain.BlockTxs) *dto.Block {
	return &dto.Block {
		ID : 		b.Block.Id,
		GasLimit : 	b.Block.GasLimit,
		GasUsed : 	b.Block.GasUsed,
		Hash : 		b.Block.Hash,
		Miner : 		b.Block.Miner,
		ParentHash : 	b.Block.ParentHash,
		Timestamp : 	b.Block.Timestamp,
	}
}

func toTransactionDTO(t domain.Transaction) *dto.Transaction {
	return &dto.Transaction{
		Hash:    t.Hash,
		From:    t.From,
		To:      t.To,
		GasUsed: t.GasUsed,
	}
}

func toEventDTO(e domain.Event) *dto.Event {
	return &dto.Event{
		LogIndex: e.LogIndex,
		Emitter:  e.Emitter,
		Data:     &e.Datas,
		Topics:   e.Topics,
	}
}