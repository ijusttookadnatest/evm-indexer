package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
)

// mockFFRepo is a dedicated repo mock for forwardfill tests.
// It notifies onCreate after each successful Create call.
type mockFFRepo struct {
	createErr   error
	createCalls int
	onCreate    chan struct{}
}

func (m *mockFFRepo) GetBackfillCursor() (uint64, error)  { return 0, nil }
func (m *mockFFRepo) UpdateBackfillCursor(_ uint64) error { return nil }
func (m *mockFFRepo) ResetBackfillCursor() error          { return nil }
func (m *mockFFRepo) Delete(_ int) error                  { return nil }
func (m *mockFFRepo) Create(_ domain.Block, _ []domain.Transaction, _ []domain.Event) error {
	m.createCalls++
	if m.createErr != nil {
		return m.createErr
	}
	if m.onCreate != nil {
		m.onCreate <- struct{}{}
	}
	return nil
}

// mockFFFetcher is a dedicated fetcher mock for forwardfill tests.
// Subscribe sends ids one by one, then an optional error, then blocks on ctx.Done().
type mockFFFetcher struct {
	ids      []uint64
	subErr   error
	fetchErr error
}

func (m *mockFFFetcher) GetLastBlockId() (uint64, error) { return 0, nil }
func (m *mockFFFetcher) FetchBlock(_ context.Context, id uint64) (domain.BlockTxsEvents, error) {
	if m.fetchErr != nil {
		return domain.BlockTxsEvents{}, m.fetchErr
	}
	return domain.BlockTxsEvents{Block: domain.Block{Id: id}}, nil
}
func (m *mockFFFetcher) Subscribe(ctx context.Context, c chan<- uint64) error {
	for _, id := range m.ids {
		select {
		case c <- id:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if m.subErr != nil {
		return m.subErr
	}
	<-ctx.Done()
	return ctx.Err()
}

func TestForwardfill(t *testing.T) {
	tests := []struct {
		name        string
		fetcherIds  []uint64
		subErr      error
		fetchErr    error
		createErr   error
		wantCreates int
		wantErr     bool
	}{
		{
			name:        "three blocks indexed then context cancelled",
			fetcherIds:  []uint64{1, 2, 3},
			wantCreates: 3,
		},
		{
			name:        "context cancelled before any block",
			wantCreates: 0,
		},
		{
			name:    "subscribe sends error",
			subErr:  errors.New("ws disconnected"),
			wantErr: true,
		},
		{
			name:       "FetchBlock error propagates",
			fetcherIds: []uint64{42},
			fetchErr:   errors.New("rpc timeout"),
			wantErr:    true,
		},
		{
			name:       "repo.Create error propagates",
			fetcherIds: []uint64{1},
			createErr:  errors.New("db write failed"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onCreate := make(chan struct{}, 10)
			repo := &mockFFRepo{createErr: tt.createErr, onCreate: onCreate}
			fetcher := &mockFFFetcher{ids: tt.fetcherIds, subErr: tt.subErr, fetchErr: tt.fetchErr}
			svc := NewIndexerService(repo, fetcher, domain.IndexerStreams{Block: make(chan any, 10), Txs: make(chan any, 10), Events: make(chan any, 10)})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			errCh := make(chan error, 1)
			go func() { errCh <- svc.forwardfill(ctx) }()

			if !tt.wantErr {
				for i := 0; i < tt.wantCreates; i++ {
					select {
					case <-onCreate:
					case <-time.After(5 * time.Second):
						t.Fatal("timeout waiting for block to be processed")
					}
				}
				cancel()
			}

			var err error
			select {
			case err = <-errCh:
			case <-time.After(5 * time.Second):
				t.Fatal("timeout waiting for forwardfill to finish")
			}

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if !errors.Is(err, context.Canceled) {
				t.Errorf("expected context.Canceled, got %v", err)
			}
			if repo.createCalls != tt.wantCreates {
				t.Errorf("expected %d Create calls, got %d", tt.wantCreates, repo.createCalls)
			}
		})
	}
}
