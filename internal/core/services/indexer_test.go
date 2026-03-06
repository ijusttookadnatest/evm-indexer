package service

import (
	"errors"
	"testing"

	"github/ijusttookadnatest/indexer-evm/internal/core/ports"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		repo    ports.IndexerRepository
		fetcher ports.Fetcher
		from    uint64
		wantErr bool
	}{
		{
			name:    "GetLastBlockId error returns immediately",
			repo:    &mockIndexerRepo{},
			fetcher: &mockBackfiller{lastBlockIdErr: errors.New("rpc unreachable")},
			wantErr: true,
		},
		{
			name:    "GetBackfillCursor error propagates",
			repo:    &mockIndexerRepo{cursorErr: errors.New("db connection lost")},
			fetcher: &mockBackfiller{lastBlockId: 5},
			wantErr: true,
		},
		{
			name:    "FetchBlock error propagates",
			repo:    &mockIndexerRepo{},
			fetcher: &mockBackfiller{lastBlockId: 3, fetchErr: errors.New("rpc timeout")},
			wantErr: true,
		},
		{
			name:    "repo.Create error propagates",
			repo:    &mockIndexerRepo{createErr: errors.New("db write failed")},
			fetcher: &mockBackfiller{lastBlockId: 2},
			wantErr: true,
		},
		{
			name:    "Subscribe error propagates",
			repo:    &mockIndexerRepo{},
			fetcher: &mockFFFetcher{subErr: errors.New("ws disconnected")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewIndexerService(tt.repo, tt.fetcher)
			err := svc.Run(tt.from, 1)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
		})
	}
}
