package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github/ijusttookadnatest/indexer-evm/config"
	"github/ijusttookadnatest/indexer-evm/redis"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func serialize(data blockData) ([]byte,error) {
	json, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json, nil
}

type txData struct {
	Tx *types.Transaction 	`json:"tx"`
	Logs []*types.Log		`json:"logs"`
}

type blockData struct {
	Block *types.Block `json:"block"`
	Txs []txData       `json:"txs"`
}

func Backfill() error {
	ctx := context.Background()
	client, err := ethclient.Dial(config.Get().Rpc)
	if err != nil {
		return err
	}

	lastBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	firstBlock := lastBlock - uint64(config.Get().LastBlocks)

	for currId := firstBlock ; currId < lastBlock ; {
		data := blockData{}
		time.Sleep(time.Millisecond * 500)
		fmt.Printf("Block: %v\n", currId)

		data.Block, err = client.BlockByNumber(ctx, big.NewInt(int64(currId)))
		if err != nil {
			if strings.Contains(err.Error(), "429 Too Many Requests") {
				fmt.Println(".")
				continue
			}
			return err
		}
		txs := data.Block.Transactions()

		for _, tx := range txs {
			receipt, err := client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				if strings.Contains(err.Error(), "429 Too Many Requests") {
					fmt.Println(".")
					continue
				}
				return err
			}
			data.Txs = append(data.Txs, txData{
				Tx: tx,
				Logs: receipt.Logs,
			})
		}

		currId++
		dataJSON, err := serialize(data)
		if err != nil {
			return err
		}
		if err = redis.Producer(ctx, dataJSON) ; err != nil {
			return err
		}
	}
	return nil
}

