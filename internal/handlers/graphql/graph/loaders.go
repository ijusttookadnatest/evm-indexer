package graph

import (
	"context"
	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/graphql/graph/dto"
	"net/http"
	"time"

	"github.com/graph-gophers/dataloader/v7"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

type Loaders struct {
	TransactionsByBlockLoader *dataloader.Loader[uint64, []*dto.Transaction]
	EventsByTransactionLoader *dataloader.Loader[string, []*dto.Event]
}

type transactionsReader struct {
	service ports.QueryService
}

func (r *transactionsReader) getBatchTransactions(ctx context.Context, blockIDs []uint64) []*dataloader.Result[[]*dto.Transaction] {
	mTxs, err := r.service.GetTransactionsByBatchBlocksId(ctx, blockIDs, true)
	if err != nil {
		return handleError[[]*dto.Transaction](len(blockIDs), err)
	}

	result := make([]*dataloader.Result[[]*dto.Transaction], len(blockIDs))
	for i, id := range blockIDs {
		txs := mTxs[id]
		txsPerBlockID := make([]*dto.Transaction, len(txs))
		for j, tx := range txs {
			txsPerBlockID[j] = toTransactionDTO(tx)
		}
		result[i] = &dataloader.Result[[]*dto.Transaction]{Data: txsPerBlockID}
	}
	return result
}

type eventsReader struct {
	service ports.QueryService
}

func (r *eventsReader) getBatchEvents(ctx context.Context, txHashs []string) []*dataloader.Result[[]*dto.Event] {
	mEvents, err := r.service.GetEventsByBatchTxsHash(ctx, txHashs)
	if err != nil {
		return handleError[[]*dto.Event](len(txHashs), err)
	}

	result := make([]*dataloader.Result[[]*dto.Event], len(txHashs))
	for i, tx := range txHashs {
		events := mEvents[tx]
		eventsByTxHash := make([]*dto.Event, len(events))
		for j, e := range events {
			eventsByTxHash[j] = toEventDTO(e)
		}
		result[i] = &dataloader.Result[[]*dto.Event]{Data: eventsByTxHash}
	}
	return result
}

// handleError creates array of result with the same error repeated for as many items requested
func handleError[T any](itemsLength int, err error) []*dataloader.Result[T] {
	result := make([]*dataloader.Result[T], itemsLength)
	for i := 0; i < itemsLength; i++ {
		result[i] = &dataloader.Result[T]{Error: err}
	}
	return result
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(service ports.QueryService) *Loaders {
	// define the data loader
	tr := &transactionsReader{service: service}
	er := &eventsReader{service: service}
	return &Loaders{
		TransactionsByBlockLoader: dataloader.NewBatchedLoader(tr.getBatchTransactions, dataloader.WithWait[uint64, []*dto.Transaction](time.Millisecond)),
		EventsByTransactionLoader: dataloader.NewBatchedLoader(er.getBatchEvents, dataloader.WithWait[string, []*dto.Event](time.Millisecond)),
	}
}

// Middleware injects data loaders into the context
func Middleware(service ports.QueryService, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()

		loaders := NewLoaders(service)
		r = r.WithContext(context.WithValue(ctx, loadersKey, loaders))
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

func GetTransaction(ctx context.Context, blockID uint64) ([]*dto.Transaction, error) {
	loaders := For(ctx)
	return loaders.TransactionsByBlockLoader.Load(ctx, blockID)()
}

func GetTransactions(ctx context.Context, blockIDs []uint64) ([][]*dto.Transaction, []error) {
	loaders := For(ctx)
	return loaders.TransactionsByBlockLoader.LoadMany(ctx, blockIDs)()
}

func GetEvent(ctx context.Context, txHash string) ([]*dto.Event, error) {
	loaders := For(ctx)
	return loaders.EventsByTransactionLoader.Load(ctx, txHash)()
}

func GetEvents(ctx context.Context, txHashs []string) ([][]*dto.Event, []error) {
	loaders := For(ctx)
	return loaders.EventsByTransactionLoader.LoadMany(ctx, txHashs)()
}
