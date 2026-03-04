package fetcher

import (
	"context"
	"fmt"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Backfiller struct {
	client *ethclient.Client
}

func NewBackfiller(url string) (*Backfiller,error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Backfiller{client:client}, nil
}

type RPCTransaction struct {
	Hash 	common.Hash      `json:"hash"`
	From 	common.Address   `json:"from"`
	To 		*common.Address   `json:"to"`
	GasUsed hexutil.Uint64   `json:"gasUsed"`
}

type RPCBlock struct {
	Hash        common.Hash      `json:"hash"`
	Number      hexutil.Uint64   `json:"number"`
	ParentHash  common.Hash      `json:"parentHash"`
	Timestamp   hexutil.Uint64   `json:"timestamp"`
	GasLimit    hexutil.Uint64   `json:"gasLimit"`
	GasUsed     hexutil.Uint64   `json:"gasUsed"`
	Miner       common.Address   `json:"miner"`
	Transactions []RPCTransaction `json:"transactions"`
}

func (b *Backfiller) FetchBlock(id uint64) (*domain.Block,[]domain.Transaction,[]domain.Event,error) {
	ctx := context.Background()
	idHex := fmt.Sprintf("0x%x", id)
	body := new(RPCBlock)

	err := b.client.Client().CallContext(context.Background(), &body, "eth_getBlockByNumber", idHex, true)
	if err != nil {
		return nil,nil,nil,err
	}
	receipts, err := b.client.BlockReceipts(ctx, rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(id)))
	if err != nil {
		return nil, nil,nil, err
	}
	
	block := extractBlock(*body)

	txs := make([]domain.Transaction, len(body.Transactions))
	for i, tx := range body.Transactions {
		txs[i] = extractTransaction(tx, *receipts[i])
	}
	
	var events []domain.Event
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			event := extractEvent(*log)
			events = append(events, event)
		}
	}

	return &block, txs, events, nil
}

func (b *Backfiller) GetLastBlockId() (uint64,error) {
	ctx := context.Background()
	id, err := b.client.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func extractEvent(log types.Log) domain.Event {
	topics := extractTopics(log.Topics)
	return domain.Event{
		BlockId: log.BlockNumber,
		TxHash: log.TxHash.Hex(),
		LogIndex: uint64(log.Index),
		Emitter: log.Address.Hex(),
		Datas: string(log.Data),
		Topics: topics,
	}
}

func extractTransaction(tx RPCTransaction, receipt types.Receipt) domain.Transaction {
	to := tx.To.Hex()
	return domain.Transaction{
		BlockId: receipt.BlockNumber.Uint64(),
		Hash: tx.Hash.Hex(),
		From: tx.From.Hex(),
		To: &to,
		GasUsed: uint64(receipt.CumulativeGasUsed),
		Status: receipt.Status,
	}
}

func extractBlock(body RPCBlock) domain.Block {
	return domain.Block{
		Id:			  uint64(body.Number),
		Hash:         body.Hash.Hex(),
		ParentHash:   body.ParentHash.Hex(),
		Miner:        body.Miner.Hex(),
		GasLimit:     uint64(body.GasLimit),
		GasUsed:      uint64(body.GasUsed),
		Timestamp:    uint64(body.Timestamp),
	}
}

func extractTopics(topics []common.Hash) []string {
	result := make([]string, len(topics))
	for i, topic := range topics {
		result[i] = topic.Hex()
	}
	return result
}