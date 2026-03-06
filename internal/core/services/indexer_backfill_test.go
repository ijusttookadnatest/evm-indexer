package service

import (
	"context"
	"errors"
	"testing"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
)

type mockIndexerRepo struct {
	lastIndexedId uint64
	lastIdErr     error
	createErr     error
	createCalls   int
	deleteErr     error
	deleteCalls   int
}

func (m *mockIndexerRepo) GetLastIndexedId() (uint64, error) {
	return m.lastIndexedId, m.lastIdErr
}

func (m *mockIndexerRepo) Create(_ domain.Block, _ []domain.Transaction, _ []domain.Event) error {
	m.createCalls++
	return m.createErr
}

func (m *mockIndexerRepo) Delete(_ int) error {
	m.deleteCalls++
	return m.deleteErr
}

type mockBackfiller struct {
	lastBlockId    uint64
	lastBlockIdErr error
	fetchErr       error
}

func (m *mockBackfiller) GetLastBlockId() (uint64, error) {
	return m.lastBlockId, m.lastBlockIdErr
}

func (m *mockBackfiller) FetchBlock(id uint64) (domain.BlockTxsEvents, error) {
	if m.fetchErr != nil {
		return domain.BlockTxsEvents{}, m.fetchErr
	}
	return domain.BlockTxsEvents{Block: domain.Block{Id: id}}, nil
}

func (m *mockBackfiller) Subscribe(_ context.Context, _ chan<- uint64, _ chan<- error) {}

func TestBackfill(t *testing.T) {
	// slog.SetLogLoggerLevel(slog.LevelDebug)
	tests := []struct {
		name       string
		repo       *mockIndexerRepo
		backfiller *mockBackfiller
		from       uint64
		wantCalls  int
		wantErr    bool
	}{
		{
			name:       "WithoutDB_FiveBlocks (0→5)",
			repo:       &mockIndexerRepo{lastIdErr: domain.ErrNotFound},
			backfiller: &mockBackfiller{lastBlockId: 5},
			from:       0,
			wantCalls:  6,
		},
		{
			name:       "WithoutDB_OneBlock (10→10)",
			repo:       &mockIndexerRepo{lastIdErr: domain.ErrNotFound},
			backfiller: &mockBackfiller{lastBlockId: 10},
			from:       10,
			wantCalls:  1,
		},
		{
			name:       "WithoutDB_OneThousandBlocks (0→1000)",
			repo:       &mockIndexerRepo{lastIdErr: domain.ErrNotFound},
			backfiller: &mockBackfiller{lastBlockId: 1000},
			from:       0,
			wantCalls:  1001,
		},
		{
			name:       "WithDB_FiveBlocks (lastIndexed=6, 7→10)",
			repo:       &mockIndexerRepo{lastIndexedId: 6},
			backfiller: &mockBackfiller{lastBlockId: 10},
			from:       0,
			wantCalls:  4,
		},
		{
			name:       "already up to date (lastIndexed=10, target=10)",
			repo:       &mockIndexerRepo{lastIndexedId: 10},
			backfiller: &mockBackfiller{lastBlockId: 10},
			from:       0,
			wantCalls:  0,
		},
		{
			name:       "GetLastIndexedId non-ErrNotFound error",
			repo:       &mockIndexerRepo{lastIdErr: errors.New("db connection lost")},
			backfiller: &mockBackfiller{lastBlockId: 10},
			from:       0,
			wantErr:    true,
		},
		{
			name:       "GetLastBlockId error",
			repo:       &mockIndexerRepo{lastIdErr: domain.ErrNotFound},
			backfiller: &mockBackfiller{lastBlockIdErr: errors.New("rpc unreachable")},
			from:       0,
			wantErr:    true,
		},
		{
			name:       "repo.Create error propagates",
			repo:       &mockIndexerRepo{lastIdErr: domain.ErrNotFound, createErr: errors.New("db write failed")},
			backfiller: &mockBackfiller{lastBlockId: 3},
			from:       0,
			wantErr:    true,
		},
		{
			name:       "FetchBlock error propagates",
			repo:       &mockIndexerRepo{lastIdErr: domain.ErrNotFound},
			backfiller: &mockBackfiller{lastBlockId: 100, fetchErr: errors.New("rpc timeout")},
			from:       0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewIndexerService(tt.repo, tt.backfiller)
			err := svc.Backfill(tt.from, 2)
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
		})
	}
}
