package graph

import (
	"context"
	"errors"
	"testing"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
	"github/ijusttookadnatest/indexer-evm/internal/handlers/graphql/graph/dto"
)

// ── serviceMock ───────────────────────────────────────────────────────────────

type serviceMock struct {
	block     *domain.BlockTxs
	blocks    []domain.BlockTxs
	events    []domain.Event
	event     *domain.Event
	txs       []domain.Transaction
	txsMap    map[uint64][]domain.Transaction
	eventsMap map[string][]domain.Event
	err       error
}

func (m serviceMock) GetBlockByHash(_ context.Context, _ string, _ bool) (*domain.BlockTxs, error) {
	return m.block, m.err
}
func (m serviceMock) GetBlockById(_ context.Context, _ uint64, _ bool) (*domain.BlockTxs, error) {
	return m.block, m.err
}
func (m serviceMock) GetBlocksWithOffset(_ context.Context, _, _ uint64, _ bool) ([]domain.BlockTxs, error) {
	return m.blocks, m.err
}
func (m serviceMock) GetBlocksByRangeTime(_ context.Context, _, _ uint64, _ bool) ([]domain.BlockTxs, error) {
	return m.blocks, m.err
}
func (m serviceMock) GetTransactionsByFilter(_ context.Context, _ domain.TransactionFilter) ([]domain.Transaction, error) {
	return m.txs, m.err
}
func (m serviceMock) GetTransactionsByBatchBlocksId(_ context.Context, _ []uint64, _ bool) (map[uint64][]domain.Transaction, error) {
	return m.txsMap, m.err
}
func (m serviceMock) GetEventsByFilter(_ context.Context, _ domain.EventFilter) ([]domain.Event, error) {
	return m.events, m.err
}
func (m serviceMock) GetEventByTxHashLogIndex(_ context.Context, _ string, _ int) (*domain.Event, error) {
	return m.event, m.err
}
func (m serviceMock) GetEventsByBatchTxsHash(_ context.Context, _ []string) (map[string][]domain.Event, error) {
	return m.eventsMap, m.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func ptr[T any](v T) *T { return &v }

func validHash() string {
	return "0x" + "a1b2c3d4e5f60718293a4b5c6d7e8f901a2b3c4d5e6f708192a3b4c5d6e7f801"
}

func validAddr() string {
	return "0x" + "abcdef1234567890abcdef1234567890abcdef12"
}

func newQueryResolver(svc serviceMock) *queryResolver {
	return &queryResolver{&Resolver{Service: svc}}
}

func ctxWithLoaders(svc serviceMock) context.Context {
	loaders := NewLoaders(svc)
	return context.WithValue(context.Background(), loadersKey, loaders)
}

// ── TestQueryBlocks ───────────────────────────────────────────────────────────

func TestQueryBlocks(t *testing.T) {
	sampleBlock := domain.BlockTxs{Block: domain.Block{Id: 42, Hash: validHash()}}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		service serviceMock
		filter  *dto.BlockFilter
		wantErr bool
		wantLen int
	}{
		{
			name:    "no params",
			service: serviceMock{},
			filter:  &dto.BlockFilter{},
			wantErr: true,
		},
		{
			name:    "conflicting — id and range time",
			service: serviceMock{},
			filter:  &dto.BlockFilter{ID: ptr(uint64(42)), FromTime: ptr(uint64(1000)), ToTime: ptr(uint64(2000))},
			wantErr: true,
		},
		{
			name:    "conflicting — id and range id",
			service: serviceMock{},
			filter:  &dto.BlockFilter{ID: ptr(uint64(42)), FromID: ptr(uint64(10)), Offset: ptr(uint64(5))},
			wantErr: true,
		},
		{
			name:    "service error",
			service: serviceMock{err: repoErr},
			filter:  &dto.BlockFilter{ID: ptr(uint64(42))},
			wantErr: true,
		},
		{
			name:    "by id",
			service: serviceMock{block: &sampleBlock},
			filter:  &dto.BlockFilter{ID: ptr(uint64(42))},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "by range id",
			service: serviceMock{blocks: []domain.BlockTxs{sampleBlock, sampleBlock}},
			filter:  &dto.BlockFilter{FromID: ptr(uint64(10)), Offset: ptr(uint64(5))},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:    "by range time",
			service: serviceMock{blocks: []domain.BlockTxs{sampleBlock}},
			filter:  &dto.BlockFilter{FromTime: ptr(uint64(1000)), ToTime: ptr(uint64(2000))},
			wantErr: false,
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newQueryResolver(tc.service)
			got, err := r.Blocks(context.Background(), tc.filter)

			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr && len(got) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

// ── TestQueryTransactions ─────────────────────────────────────────────────────

func TestQueryTransactions(t *testing.T) {
	sampleTx := domain.Transaction{Hash: validHash(), From: validAddr()}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		service serviceMock
		filter  *dto.TransactionFilter
		wantErr bool
		wantLen int
	}{
		{
			name:    "service error",
			service: serviceMock{err: repoErr},
			filter:  &dto.TransactionFilter{},
			wantErr: true,
		},
		{
			name:    "happy path",
			service: serviceMock{txs: []domain.Transaction{sampleTx}},
			filter:  &dto.TransactionFilter{From: ptr(validAddr())},
			wantErr: false,
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newQueryResolver(tc.service)
			got, err := r.Transactions(context.Background(), tc.filter)

			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr && len(got) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(got), tc.wantLen)
			}
			if !tc.wantErr && got[0].Hash != sampleTx.Hash {
				t.Errorf("hash: got %s, want %s", got[0].Hash, sampleTx.Hash)
			}
		})
	}
}

// ── TestQueryEvents ───────────────────────────────────────────────────────────

func TestQueryEvents(t *testing.T) {
	sampleEvent := domain.Event{TxHash: validHash(), Emitter: validAddr()}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		service serviceMock
		filter  *dto.EventFilter
		wantErr bool
		wantLen int
	}{
		{
			name:    "service error",
			service: serviceMock{err: repoErr},
			filter:  &dto.EventFilter{Emitter: ptr(validAddr())},
			wantErr: true,
		},
		{
			name:    "happy path by emitter",
			service: serviceMock{events: []domain.Event{sampleEvent}},
			filter:  &dto.EventFilter{Emitter: ptr(validAddr())},
			wantErr: false,
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newQueryResolver(tc.service)
			got, err := r.Events(context.Background(), tc.filter)

			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr && len(got) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(got), tc.wantLen)
			}
			if !tc.wantErr && got[0].Emitter != sampleEvent.Emitter {
				t.Errorf("emitter: got %s, want %s", got[0].Emitter, sampleEvent.Emitter)
			}
		})
	}
}

// ── TestBlockResolverTransactions ─────────────────────────────────────────────

func TestBlockResolverTransactions(t *testing.T) {
	blockID := uint64(42)
	sampleTx := domain.Transaction{BlockId: blockID, Hash: validHash(), From: validAddr()}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		service serviceMock
		wantErr bool
		wantLen int
	}{
		{
			name:    "loader error",
			service: serviceMock{err: repoErr},
			wantErr: true,
		},
		{
			name:    "happy path",
			service: serviceMock{txsMap: map[uint64][]domain.Transaction{blockID: {sampleTx}}},
			wantErr: false,
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ctxWithLoaders(tc.service)
			r := &blockResolver{&Resolver{Service: tc.service}}
			got, err := r.Transactions(ctx, &dto.Block{ID: blockID})

			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr && len(got) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

// ── TestTransactionResolverEvents ─────────────────────────────────────────────

func TestTransactionResolverEvents(t *testing.T) {
	txHash := validHash()
	sampleEvent := domain.Event{TxHash: txHash, Emitter: validAddr()}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		service serviceMock
		wantErr bool
		wantLen int
	}{
		{
			name:    "loader error",
			service: serviceMock{err: repoErr},
			wantErr: true,
		},
		{
			name:    "happy path",
			service: serviceMock{eventsMap: map[string][]domain.Event{txHash: {sampleEvent}}},
			wantErr: false,
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ctxWithLoaders(tc.service)
			r := &transactionResolver{&Resolver{Service: tc.service}}
			got, err := r.Events(ctx, &dto.Transaction{Hash: txHash})

			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.wantErr && len(got) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}
