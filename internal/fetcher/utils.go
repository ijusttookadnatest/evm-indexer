package fetcher

import (
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func extractEvent(log types.Log) domain.Event {
	return domain.Event{
		BlockId:  log.BlockNumber,
		TxHash:   log.TxHash.Hex(),
		LogIndex: uint64(log.Index),
		Emitter:  log.Address.Hex(),
		Datas:    hexutil.Encode(log.Data),
		Topics:   extractTopics(log.Topics),
	}
}

func extractTransaction(tx RPCTransaction, receipt types.Receipt) domain.Transaction {
	var to *string
	if tx.To != nil {
		s := tx.To.Hex()
		to = &s
	}
	return domain.Transaction{
		BlockId: receipt.BlockNumber.Uint64(),
		Hash:    tx.Hash.Hex(),
		From:    tx.From.Hex(),
		To:      to,
		GasUsed: uint64(receipt.CumulativeGasUsed),
		Status:  receipt.Status,
	}
}

func extractBlock(body RPCBlock) domain.Block {
	return domain.Block{
		Id:         uint64(body.Number),
		Hash:       body.Hash.Hex(),
		ParentHash: body.ParentHash.Hex(),
		Miner:      body.Miner.Hex(),
		GasLimit:   uint64(body.GasLimit),
		GasUsed:    uint64(body.GasUsed),
		Timestamp:  uint64(body.Timestamp),
	}
}

func extractTopics(topics []common.Hash) []string {
	result := make([]string, len(topics))
	for i, topic := range topics {
		result[i] = topic.Hex()
	}
	return result
}
