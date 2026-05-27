package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/time/rate"

	"github.com/cenkalti/backoff/v5"
)

type ethWrapper struct {
	*ethclient.Client
}

type priorityKey struct{}

func (w *ethWrapper) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	return w.Client.Client().BatchCallContext(ctx, b)
}

type EVMClient interface {
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
	BlockNumber(ctx context.Context) (uint64, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
}

type Fetcher struct {
	clientHTTP  EVMClient
	clientWS    EVMClient
	rateLimiter *rate.Limiter
}

func NewFetcher(urlHTTP string, urlWS string, rpcRateLimit float64) (*Fetcher, error) {
	clientHTTP, err := ethclient.Dial(urlHTTP)
	if err != nil {
		return nil, err
	}
	clientWS, err := ethclient.Dial(urlWS)
	if err != nil {
		return nil, err
	}
	var limiter *rate.Limiter
	if rpcRateLimit <= 0 {
		limiter = rate.NewLimiter(rate.Inf, 0)
	} else {
		limiter = rate.NewLimiter(rate.Limit(rpcRateLimit), 1)
	}
	return &Fetcher{
		clientHTTP:  &ethWrapper{clientHTTP},
		clientWS:    &ethWrapper{clientWS},
		rateLimiter: limiter,
	}, nil
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
		return nil
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

func (b *Fetcher) FetchBlock(ctx context.Context, id uint64) (domain.BlockTxsEvents, error) {
	idHex := fmt.Sprintf("0x%x", id)

	fetchAll := func() (domain.BlockTxsEvents, error) {
		if ctx.Value(priorityKey{}) == nil {
			if err := b.rateLimiter.Wait(ctx); err != nil {
				return domain.BlockTxsEvents{}, backoff.Permanent(err)
			}
		}

		body := new(RPCBlock)
		var receipts []*types.Receipt
		batch := []rpc.BatchElem{
			{Method: "eth_getBlockByNumber", Args: []interface{}{idHex, true}, Result: body},
			{Method: "eth_getBlockReceipts", Args: []interface{}{idHex}, Result: &receipts},
		}
		if err := b.clientHTTP.BatchCallContext(ctx, batch); err != nil {
			fmt.Println("here1")
			fmt.Println(err.Error())
			return domain.BlockTxsEvents{}, wrapRetryError(err)
		}
		fmt.Println("here2")
		if batch[0].Error != nil {
			fmt.Println("batch 0")
			fmt.Println(batch[0].Error.Error())
			return domain.BlockTxsEvents{}, wrapRetryError(batch[0].Error)
		}
		if batch[1].Error != nil {
			fmt.Println(batch[1].Error.Error())
			return domain.BlockTxsEvents{}, wrapRetryError(batch[1].Error)
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

	notify := func(err error, d time.Duration) {
		slog.Info("retrying", "err", err, "in", d)
	}

	result, err := backoff.Retry(ctx, fetchAll,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxElapsedTime(30*time.Second),
		backoff.WithNotify(notify),
	)
	if err != nil {
		slog.Error("failed to fetch block", "block", id, "err", err)
		return domain.BlockTxsEvents{}, err
	}
	return result, nil
}

func (b *Fetcher) FetchBlockPriority(ctx context.Context, id uint64) (domain.BlockTxsEvents, error) {
	priorityCtx := context.WithValue(ctx, priorityKey{}, true)
	return b.FetchBlock(priorityCtx, id)
}

func (b *Fetcher) GetLastBlockId() (uint64, error) {
	blockNumber := func() (uint64, error) {
		res, err := b.clientHTTP.BlockNumber(context.Background())
		return res, wrapRetryError(err)
	}
	res, err := backoff.Retry(context.Background(), blockNumber, backoff.WithBackOff(backoff.NewExponentialBackOff()), backoff.WithMaxElapsedTime(15*time.Second))
	if err != nil {
		slog.Error("failed to get last block number", "err", err)
	}
	return res, err
}

func (f *Fetcher) Subscribe(ctx context.Context, c chan<- uint64) error {
	defer close(c)

	op := func() (struct{}, error) {
		newHeader := make(chan *types.Header)
		sub, err := f.clientWS.SubscribeNewHead(ctx, newHeader)
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
	if err != nil {
		slog.Error("subscription failed", "err", err)
	}
	return err
}
