package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/cenkalti/backoff/v5"
)

type ethWrapper struct {
	*ethclient.Client
}

func (w *ethWrapper) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	return w.Client.Client().CallContext(ctx, result, method, args...)
}

type EVMClient interface {
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
	BlockReceipts(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) ([]*types.Receipt, error)
	BlockNumber(ctx context.Context) (uint64, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
}

type Fetcher struct {
	client EVMClient
}

func NewFetcher(url string) (*Fetcher, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Fetcher{client: &ethWrapper{client}}, nil
}

type RPCTransaction struct {
	Hash    common.Hash     `json:"hash"`
	From    common.Address  `json:"from"`
	To      *common.Address `json:"to"`
	GasUsed hexutil.Uint64  `json:"gasUsed"`
}

type RPCBlock struct {
	Hash         common.Hash      `json:"hash"`
	Number       hexutil.Uint64   `json:"number"`
	ParentHash   common.Hash      `json:"parentHash"`
	Timestamp    hexutil.Uint64   `json:"timestamp"`
	GasLimit     hexutil.Uint64   `json:"gasLimit"`
	GasUsed      hexutil.Uint64   `json:"gasUsed"`
	Miner        common.Address   `json:"miner"`
	Transactions []RPCTransaction `json:"transactions"`
}

func wrapRetryError(err error) error {
	if err == nil {
		return err
	}
	var rpcErr rpc.Error
	if errors.As(err, &rpcErr) {
		code := rpcErr.ErrorCode()
		if code >= -32602 && code <= -32600 {
			return backoff.Permanent(err)
		}
	}
	return err
}

func (b *Fetcher) FetchBlock(id uint64) (domain.BlockTxsEvents, error) {
	ctx := context.Background()
	idHex := fmt.Sprintf("0x%x", id)
	body := new(RPCBlock)

	callContext := func() (struct{}, error) {
		err := b.client.CallContext(ctx, body, "eth_getBlockByNumber", idHex, true)
		return struct{}{}, wrapRetryError(err)
	}

	_, err := backoff.Retry(ctx, callContext, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxElapsedTime(15*time.Second))
	if err != nil {
		return domain.BlockTxsEvents{}, err
	}

	blockReceipts := func() ([]*types.Receipt, error) {
		receipts, err := b.client.BlockReceipts(ctx, rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(id)))
		return receipts, wrapRetryError(err)
	}
	receipts, err := backoff.Retry(ctx, blockReceipts, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxElapsedTime(15*time.Second))
	if err != nil {
		return domain.BlockTxsEvents{}, err
	}

	block := extractBlock(*body)

	txs := make([]domain.Transaction, len(body.Transactions))
	for i, tx := range body.Transactions {
		txs[i] = extractTransaction(tx, *receipts[i])
	}

	var events []domain.Event
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			events = append(events, extractEvent(*log))
		}
	}

	return domain.BlockTxsEvents{Block: block, Txs: txs, Events: events}, nil
}

func (b *Fetcher) GetLastBlockId() (uint64, error) {
	blockNumber := func() (uint64, error) {
		res, err := b.client.BlockNumber(context.Background())
		return res, wrapRetryError(err)
	}
	return backoff.Retry(context.Background(), blockNumber, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxElapsedTime(15*time.Second))
}

func (f *Fetcher) Subscribe(ctx context.Context, c chan<- uint64) error {
	defer close(c)

	op := func() (struct{}, error) {
		newHeader := make(chan *types.Header)
		sub, err := f.client.SubscribeNewHead(ctx, newHeader)
		if err != nil {
			return struct{}{}, backoff.Permanent(err)
		}

		for {
			select {
			case result := <-newHeader:
				{
					c <- result.Number.Uint64()
				}
			case err := <-sub.Err():
				{
					sub.Unsubscribe()
					return struct{}{}, err
				}
			case <-ctx.Done():
				{
					sub.Unsubscribe()
					return struct{}{}, backoff.Permanent(ctx.Err())
				}
			}
		}
	}
	_, err := backoff.Retry(ctx, op, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxElapsedTime(15*time.Second))
	return err
}
