package service

import (
	"context"
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	custmetrics "github/ijusttookadnatest/evm-indexer/internal/metrics"
)

func newTestMetrics() *custmetrics.IndexerMetrics {
	return custmetrics.NewIndexerMetrics(prometheus.NewRegistry())
}

type mockPubSub struct{}

func (m *mockPubSub) Publish(_ context.Context, _ string, _ []byte) error { return nil }
func (m *mockPubSub) Subscribe(_ context.Context, _ string) (<-chan []byte, error) {
	return make(chan []byte), nil
}

type mockIndexerRepo struct {
	cursor        uint64
	cursorErr     error
	updatedCursor uint64
	createErr     error
	createCalls   int
	deleteErr     error
	deleteCalls   int
}

func (m *mockIndexerRepo) GetBackfillCursor(_ context.Context) (uint64, error) {
	return m.cursor, m.cursorErr
}

func (m *mockIndexerRepo) UpdateBackfillCursor(_ context.Context, id uint64) error {
	m.updatedCursor = id
	return nil
}

func (m *mockIndexerRepo) Create(_ context.Context, _ domain.Block, _ []domain.Transaction, _ []domain.Event) error {
	m.createCalls++
	return m.createErr
}

func (m *mockIndexerRepo) Delete(_ context.Context, _ uint64) error {
	m.deleteCalls++
	return m.deleteErr
}

func (m *mockIndexerRepo) GetBlockById(_ context.Context, _ uint64) (*domain.Block, error) {
	return nil, domain.ErrNotFound
}

func (m *mockIndexerRepo) ResetBackfillCursor(_ context.Context) error               { return nil }
func (m *mockIndexerRepo) GetBalancefillCursor(_ context.Context) (uint64, error)    { return 0, nil }
func (m *mockIndexerRepo) UpdateBalancefillCursor(_ context.Context, _ uint64) error { return nil }
func (m *mockIndexerRepo) ResetBalancefillCursor(_ context.Context) error            { return nil }
func (m *mockIndexerRepo) GetMaxIndexedBlock(_ context.Context) (uint64, error)      { return 0, nil }
func (m *mockIndexerRepo) BatchUpsertBalance(_ context.Context, _ []domain.BalanceEntry) error {
	return nil
}
func (m *mockIndexerRepo) GetLogsByTopic(_ context.Context, _ domain.LogFilter) ([]domain.Log, error) {
	return nil, nil
}

func (m *mockIndexerRepo) BulkCreate(_ context.Context, items []domain.BlockTxsEvents) error {
	m.createCalls += len(items)
	return m.createErr
}

type mockBackfiller struct {
	lastBlockId    uint64
	lastBlockIdErr error
	fetchErr       error
}

func (m *mockBackfiller) GetLastBlockId() (uint64, error) {
	return m.lastBlockId, m.lastBlockIdErr
}

func (m *mockBackfiller) FetchBlock(_ context.Context, id uint64) (domain.BlockTxsEvents, error) {
	if m.fetchErr != nil {
		return domain.BlockTxsEvents{}, m.fetchErr
	}
	return domain.BlockTxsEvents{Block: domain.Block{Id: id}}, nil
}

func (m *mockBackfiller) FetchBlockPriority(ctx context.Context, id uint64) (domain.BlockTxsEvents, error) {
	return m.FetchBlock(ctx, id)
}

func (m *mockBackfiller) Subscribe(_ context.Context, _ chan<- uint64) error { return nil }

func TestBackfill(t *testing.T) {
	// slog.SetLogLoggerLevel(slog.LevelDebug)
	tests := []struct {
		name       string
		repo       *mockIndexerRepo
		backfiller *mockBackfiller
		from       uint64
		targetId   uint64
		wantCalls  int
		wantErr    bool
		wantCursor uint64
	}{
		{
			name:       "FreshDB_FiveBlocks (0→5)",
			repo:       &mockIndexerRepo{cursor: 0},
			backfiller: &mockBackfiller{},
			from:       0,
			targetId:   5,
			wantCalls:  6,
			wantCursor: 5,
		},
		{
			name:       "FreshDB_OneBlock (10→10)",
			repo:       &mockIndexerRepo{cursor: 0},
			backfiller: &mockBackfiller{},
			from:       10,
			targetId:   10,
			wantCalls:  1,
			wantCursor: 10,
		},
		{
			name:       "FreshDB_OneThousandBlocks (0→1000)",
			repo:       &mockIndexerRepo{cursor: 0},
			backfiller: &mockBackfiller{},
			from:       0,
			targetId:   1000,
			wantCalls:  1001,
			wantCursor: 1000,
		},
		{
			name:       "Resume_FiveBlocks (cursor=6, 7→10)",
			repo:       &mockIndexerRepo{cursor: 6},
			backfiller: &mockBackfiller{},
			from:       0,
			targetId:   10,
			wantCalls:  4,
			wantCursor: 10,
		},
		{
			name:       "already up to date (cursor=10, target=10)",
			repo:       &mockIndexerRepo{cursor: 10},
			backfiller: &mockBackfiller{},
			from:       0,
			targetId:   10,
			wantCalls:  0,
			wantCursor: 0, // UpdateBackfillCursor not called
		},
		{
			name:       "GetBackfillCursor error",
			repo:       &mockIndexerRepo{cursorErr: errors.New("db connection lost")},
			backfiller: &mockBackfiller{},
			from:       0,
			targetId:   10,
			wantErr:    true,
		},
		{
			name:       "repo.Create error propagates",
			repo:       &mockIndexerRepo{cursor: 0, createErr: errors.New("db write failed")},
			backfiller: &mockBackfiller{},
			from:       0,
			targetId:   3,
			wantErr:    true,
		},
		{
			name:       "FetchBlock error propagates",
			repo:       &mockIndexerRepo{cursor: 0},
			backfiller: &mockBackfiller{fetchErr: errors.New("rpc timeout")},
			from:       0,
			targetId:   100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewIndexerService(tt.repo, tt.backfiller, &mockPubSub{}, newTestMetrics())
			done := make(chan struct{}, 1)
			err := svc.backfill(context.Background(), tt.from, tt.targetId, 2, done)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tt.repo.createCalls != tt.wantCalls {
				t.Errorf("expected %d Create calls, got %d", tt.wantCalls, tt.repo.createCalls)
			}
			if tt.repo.updatedCursor != tt.wantCursor {
				t.Errorf("cursor: want %d, got %d", tt.wantCursor, tt.repo.updatedCursor)
			}
		})
	}
}
