package service

import (
	"errors"
	"testing"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
)

// mockRepo satisfies ports.QueryRepository without a real DB.
type mockRepo struct {
	blockResult  *domain.Block
	blocksResult []domain.Block
	eventsResult []domain.Event
	txsResult    []domain.Transaction

	blockErr error
	eventErr error
	txErr    error
}

func (m *mockRepo) GetBlockByHash(_ string) (*domain.Block, error) {
	return m.blockResult, m.blockErr
}

func (m *mockRepo) GetBlockById(_ uint64) (*domain.Block, error) {
	return m.blockResult, m.blockErr
}

func (m *mockRepo) GetBlocksByRangeId(_, _ uint64) ([]domain.Block, error) {
	return m.blocksResult, m.blockErr
}

func (m *mockRepo) GetBlocksByRangeTime(_, _ uint64) ([]domain.Block, error) {
	return m.blocksResult, m.blockErr
}

func (m *mockRepo) GetTransactionByFilter(_ domain.TransactionFilter) ([]domain.Transaction, error) {
	return m.txsResult, m.txErr
}

func (m *mockRepo) GetTransactionsByBatchBlocksId(_ []uint64) ([]domain.Transaction, error) {
	return m.txsResult, m.txErr
}

func (m *mockRepo) GetEventByFilter(_ domain.EventFilter) ([]domain.Event, error) {
	return m.eventsResult, m.eventErr
}

func (m *mockRepo) GetEventByTxHashLogIndex(_ string, _ int) (*domain.Event, error) {
	if len(m.eventsResult) == 0 {
		return nil, m.eventErr
	}
	return &m.eventsResult[0], m.eventErr
}

func (m *mockRepo) GetEventsByBatchTxsHash(_ []string) ([]domain.Event, error) {
	return m.eventsResult, m.eventErr
}

// ── helpers ──────────────────────────────────────────────────────────────────

// validHash returns a well-formed EVM hash (0x + 64 hex chars).
func validHash() string {
	return "0x" + "a1b2c3d4e5f60718293a4b5c6d7e8f901a2b3c4d5e6f708192a3b4c5d6e7f801"
}

// validAddr returns a well-formed EVM address (0x + 40 hex chars).
func validAddr() string {
	return "0x" + "abcdef1234567890abcdef1234567890abcdef12"
}

// ptr wraps any value in a pointer — avoids noisy &local declarations in test cases.
func ptr[T any](v T) *T { return &v }

func newService(repo *mockRepo) *QueryService {
	return NewQueryService(repo, 100, 3600)
}

// ── GetBlockById ─────────────────────────────────────────────────────────────

func TestGetBlockById(t *testing.T) {
	sampleBlock := &domain.Block{Id: 42, Hash: validHash()}
	sampleTxs := []domain.Transaction{
		{BlockId: 42, Hash: validHash(), From: "0x" + "abcdef1234567890abcdef1234567890abcdef12"},
	}
	repoErr := errors.New("db error")

	tests := []struct {
		name      string
		id        uint64
		withTxs   bool
		repo      *mockRepo
		wantErr   error
		wantBlock *domain.Block
		wantTxLen int
	}{
		{
			name:      "zero id",
			id:        0,
			withTxs:   false,
			repo:      &mockRepo{},
			wantErr:   domain.ErrInvalidId,
			wantBlock: nil,
			wantTxLen: 0,
		},
		{
			name:      "repo error",
			id:        1,
			withTxs:   false,
			repo:      &mockRepo{blockErr: repoErr},
			wantErr:   repoErr,
			wantBlock: nil,
			wantTxLen: 0,
		},
		{
			name:      "block found without txs",
			id:        42,
			withTxs:   false,
			repo:      &mockRepo{blockResult: sampleBlock},
			wantErr:   nil,
			wantBlock: sampleBlock,
			wantTxLen: 0,
		},
		{
			name:    "block found with txs",
			id:      42,
			withTxs: true,
			repo: &mockRepo{
				blockResult: sampleBlock,
				txsResult:   sampleTxs,
			},
			wantErr:   nil,
			wantBlock: sampleBlock,
			wantTxLen: len(sampleTxs),
		},
		{
			name:    "tx repo error",
			id:      42,
			withTxs: true,
			repo: &mockRepo{
				blockResult: sampleBlock,
				txErr:       repoErr,
			},
			wantErr:   repoErr,
			wantBlock: nil,
			wantTxLen: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetBlockById(tc.id, tc.withTxs)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if got.Block != *tc.wantBlock {
				t.Errorf("block mismatch: got %+v, want %+v", got.Block, *tc.wantBlock)
			}
			if len(got.Txs) != tc.wantTxLen {
				t.Errorf("txs length: got %d, want %d", len(got.Txs), tc.wantTxLen)
			}
		})
	}
}

// ── GetBlocksWithOffset ──────────────────────────────────────────────────────

func TestGetBlocksWithOffset(t *testing.T) {
	// Two blocks with distinct IDs so we can verify the tx-map routing.
	sampleBlocks := []domain.Block{
		{Id: 10, Hash: validHash()},
		{Id: 11, Hash: validHash()},
	}
	sampleTxs := []domain.Transaction{
		{BlockId: 10, Hash: validHash(), From: "0x" + "abcdef1234567890abcdef1234567890abcdef12"},
		{BlockId: 11, Hash: validHash(), From: "0x" + "abcdef1234567890abcdef1234567890abcdef12"},
	}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		from    uint64
		offset  uint64
		withTxs bool
		repo    *mockRepo
		wantErr error
		wantLen int // number of BlockTxs returned
	}{
		{
			// offsetMax=100, so 101 must be rejected.
			name:    "offset exceeds max",
			from:    0,
			offset:  101,
			withTxs: false,
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidOffset,
			wantLen: 0,
		},
		{
			// offset=0 is not an error: the service defaults it to 100.
			name:    "offset zero uses default",
			from:    0,
			offset:  0,
			withTxs: false,
			repo:    &mockRepo{blocksResult: sampleBlocks},
			wantErr: nil,
			wantLen: len(sampleBlocks),
		},
		{
			name:    "repo error",
			from:    0,
			offset:  10,
			withTxs: false,
			repo:    &mockRepo{blockErr: repoErr},
			wantErr: repoErr,
			wantLen: 0,
		},
		{
			name:    "blocks found without txs",
			from:    10,
			offset:  5,
			withTxs: false,
			repo:    &mockRepo{blocksResult: sampleBlocks},
			wantErr: nil,
			wantLen: len(sampleBlocks),
		},
		{
			// txsResult contains one tx per block — each BlockTxs.Txs should be populated.
			name:    "blocks found with txs",
			from:    10,
			offset:  5,
			withTxs: true,
			repo: &mockRepo{
				blocksResult: sampleBlocks,
				txsResult:    sampleTxs,
			},
			wantErr: nil,
			wantLen: len(sampleBlocks),
		},
		{
			name:    "tx repo error",
			from:    10,
			offset:  5,
			withTxs: true,
			repo: &mockRepo{
				blocksResult: sampleBlocks,
				txErr:        repoErr,
			},
			wantErr: repoErr,
			wantLen: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetBlocksWithOffest(tc.from, tc.offset, tc.withTxs)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if len(got) != tc.wantLen {
				t.Errorf("blocks length: got %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

// ── GetBlocksByRangeTime ─────────────────────────────────────────────────────

func TestGetBlocksByRangeTime(t *testing.T) {
	sampleBlocks := []domain.Block{
		{Id: 10, Hash: validHash()},
		{Id: 11, Hash: validHash()},
	}
	sampleTxs := []domain.Transaction{
		{BlockId: 10, Hash: validHash(), From: "0x" + "abcdef1234567890abcdef1234567890abcdef12"},
		{BlockId: 11, Hash: validHash(), From: "0x" + "abcdef1234567890abcdef1234567890abcdef12"},
	}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		from    uint64
		to      uint64
		withTxs bool
		repo    *mockRepo
		wantErr error
		wantLen int
	}{
		{
			name:    "from zero",
			from:    0,
			to:      100,
			withTxs: false,
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidId,
			wantLen: 0,
		},
		{
			name:    "to zero",
			from:    100,
			to:      0,
			withTxs: false,
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidId,
			wantLen: 0,
		},
		{
			// from == to is also invalid (range must be strictly increasing).
			name:    "from equal to to",
			from:    100,
			to:      100,
			withTxs: false,
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidId,
			wantLen: 0,
		},
		{
			// rangeMaxTime=3600, so a span of 3601 must be rejected.
			name:    "range exceeds max",
			from:    1000,
			to:      4602,
			withTxs: false,
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidId,
			wantLen: 0,
		},
		{
			name:    "repo error",
			from:    1000,
			to:      1100,
			withTxs: false,
			repo:    &mockRepo{blockErr: repoErr},
			wantErr: repoErr,
			wantLen: 0,
		},
		{
			name:    "blocks found without txs",
			from:    1000,
			to:      1100,
			withTxs: false,
			repo:    &mockRepo{blocksResult: sampleBlocks},
			wantErr: nil,
			wantLen: len(sampleBlocks),
		},
		{
			name:    "blocks found with txs",
			from:    1000,
			to:      1100,
			withTxs: true,
			repo: &mockRepo{
				blocksResult: sampleBlocks,
				txsResult:    sampleTxs,
			},
			wantErr: nil,
			wantLen: len(sampleBlocks),
		},
	}

	_ = sampleBlocks
	_ = sampleTxs
	_ = repoErr

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetBlocksByRangeTime(tc.from, tc.to, tc.withTxs)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if len(got) != tc.wantLen {
				t.Errorf("blocks length: got %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

// ── GetBlockByHash ────────────────────────────────────────────────────────────

func TestGetBlockByHash(t *testing.T) {
	sampleBlock := &domain.Block{
		Id:   42,
		Hash: validHash(),
	}
	sampleTxs := []domain.Transaction{
		{BlockId: 42, Hash: validHash(), From: "0x" + "abcdef1234567890abcdef1234567890abcdef12"},
	}
	repoErr := errors.New("db error")

	tests := []struct {
		name      string
		hash      string
		withTxs   bool
		repo      *mockRepo
		wantErr   error
		wantBlock *domain.Block
		wantTxLen int
	}{
		{
			name: "invalid hash",
			hash : "abcdef1234567890abcdef1234567890abcdef12",
			withTxs: false,
			repo: &mockRepo{},
			wantErr: domain.ErrInvalidHash,
			wantBlock: nil,
			wantTxLen: 0,
		},

		{
			// 66 chars but no "0x" prefix — length passes, prefix fails.
			name:      "hash wrong prefix",
			hash:      "1x" + "a1b2c3d4e5f60718293a4b5c6d7e8f901a2b3c4d5e6f708192a3b4c5d6e7f801",
			withTxs:   false,
			repo:      &mockRepo{},
			wantErr:   domain.ErrInvalidHash,
			wantBlock: nil,
			wantTxLen: 0,
		},
		{
			// Hash is valid but the repository layer returns an error (e.g. DB down).
			name:      "repo error",
			hash:      validHash(),
			withTxs:   false,
			repo:      &mockRepo{blockErr: repoErr},
			wantErr:   repoErr,
			wantBlock: nil,
			wantTxLen: 0,
		},
		{
			// Happy path: caller does not want transactions hydrated.
			name:      "block found without txs",
			hash:      validHash(),
			withTxs:   false,
			repo:      &mockRepo{blockResult: sampleBlock},
			wantErr:   nil,
			wantBlock: sampleBlock,
			wantTxLen: 0,
		},
		{
			// Happy path: caller wants transactions — repo returns one tx.
			name:    "block found with txs",
			hash:    validHash(),
			withTxs: true,
			repo: &mockRepo{
				blockResult: sampleBlock,
				txsResult:   sampleTxs,
			},
			wantErr:   nil,
			wantBlock: sampleBlock,
			wantTxLen: len(sampleTxs),
		},
		{
			// Block fetch succeeds but the tx fetch fails — error must bubble up.
			name:    "tx repo error",
			hash:    validHash(),
			withTxs: true,
			repo: &mockRepo{
				blockResult: sampleBlock,
				txErr:       repoErr,
			},
			wantErr:   repoErr,
			wantBlock: nil,
			wantTxLen: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetBlockByHash(tc.hash, tc.withTxs)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return // error path: no further checks needed
			}
			if got.Block != *tc.wantBlock {
				t.Errorf("block mismatch: got %+v, want %+v", got.Block, *tc.wantBlock)
			}
			if len(got.Txs) != tc.wantTxLen {
				t.Errorf("txs length: got %d, want %d", len(got.Txs), tc.wantTxLen)
			}
		})
	}
}

// ── GetEventsByFilter ────────────────────────────────────────────────────────

func TestGetEventsByFilter(t *testing.T) {
	sampleEvents := []domain.Event{
		{BlockId: 1, TxHash: validHash(), LogIndex: 0, Emitter: validAddr()},
	}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		filter  domain.EventFilter
		repo    *mockRepo
		wantErr error
		wantLen int
	}{
		{
			name:    "empty filter",
			filter:  domain.EventFilter{},
			repo:    &mockRepo{},
			wantErr: domain.ErrEmptyFilter,
		},
		{
			// FromBlock set but ToBlock missing.
			name:    "from block without to",
			filter:  domain.EventFilter{FromBlock: ptr(uint64(1))},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidBlockRange,
		},
		{
			// ToBlock set but FromBlock missing.
			name:    "to block without from",
			filter:  domain.EventFilter{ToBlock: ptr(uint64(100))},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidBlockRange,
		},
		{
			// offsetMax=100: span of 101 exceeds the limit.
			name: "block range too wide",
			filter: domain.EventFilter{
				FromBlock: ptr(uint64(1)),
				ToBlock:   ptr(uint64(102)),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidBlockRange,
		},
		{
			name: "limit zero",
			filter: domain.EventFilter{
				TxHash: ptr(validHash()),
				Limit:  ptr(0),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidLimit,
		},
		{
			// offsetMax=100: any limit above 100 is rejected.
			name: "limit exceeds max",
			filter: domain.EventFilter{
				TxHash: ptr(validHash()),
				Limit:  ptr(101),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidLimit,
		},
		{
			name:    "invalid tx hash",
			filter:  domain.EventFilter{TxHash: ptr("not-a-hash")},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidHash,
		},
		{
			name:    "invalid emitter address",
			filter:  domain.EventFilter{Emitter: ptr("0xinvalid")},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidAddress,
		},
		{
			name:    "invalid topic",
			filter:  domain.EventFilter{Topics: []string{"short"}},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidTopics,
		},
		{
			name: "repo error",
			filter: domain.EventFilter{
				TxHash: ptr(validHash()),
				Limit:  ptr(10),
			},
			repo:    &mockRepo{eventErr: repoErr},
			wantErr: repoErr,
		},
		{
			// Limit is nil — triggers the default path inside GetEventsByFilter.
			name:    "events found",
			filter:  domain.EventFilter{TxHash: ptr(validHash())},
			repo:    &mockRepo{eventsResult: sampleEvents},
			wantErr: nil,
			wantLen: len(sampleEvents),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetEventsByFilter(tc.filter)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if len(got) != tc.wantLen {
				t.Errorf("events length: got %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

// ── GetTransactionsByFilter ──────────────────────────────────────────────────

func TestGetTransactionsByFilter(t *testing.T) {
	sampleTxs := []domain.Transaction{
		{BlockId: 1, Hash: validHash(), From: validAddr()},
	}
	repoErr := errors.New("db error")

	tests := []struct {
		name    string
		filter  domain.TransactionFilter
		repo    *mockRepo
		wantErr error
		wantLen int
	}{
		{
			name:    "empty filter",
			filter:  domain.TransactionFilter{},
			repo:    &mockRepo{},
			wantErr: domain.ErrEmptyFilter,
		},
		{
			// BlockId pointer is non-nil but the value is 0.
			name:    "block id zero",
			filter:  domain.TransactionFilter{BlockId: ptr(uint64(0))},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidId,
		},
		{
			// Hash keeps the filter non-empty; FromBlock without ToBlock triggers range error.
			name: "from block without to",
			filter: domain.TransactionFilter{
				Hash:      ptr(validHash()),
				FromBlock: ptr(uint64(1)),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidBlockRange,
		},
		{
			name: "to block without from",
			filter: domain.TransactionFilter{
				Hash:    ptr(validHash()),
				ToBlock: ptr(uint64(100)),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidBlockRange,
		},
		{
			// offsetMax=100: span of 101 exceeds the limit.
			name: "block range too wide",
			filter: domain.TransactionFilter{
				Hash:      ptr(validHash()),
				FromBlock: ptr(uint64(1)),
				ToBlock:   ptr(uint64(102)),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidBlockRange,
		},
		{
			name: "limit zero",
			filter: domain.TransactionFilter{
				Hash:  ptr(validHash()),
				Limit: ptr(0),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidLimit,
		},
		{
			name: "limit exceeds max",
			filter: domain.TransactionFilter{
				Hash:  ptr(validHash()),
				Limit: ptr(101),
			},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidLimit,
		},
		{
			name:    "invalid hash",
			filter:  domain.TransactionFilter{Hash: ptr("short")},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidHash,
		},
		{
			name:    "invalid from address",
			filter:  domain.TransactionFilter{From: ptr("0xinvalid")},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidAddress,
		},
		{
			name:    "invalid to address",
			filter:  domain.TransactionFilter{To: ptr("0xinvalid")},
			repo:    &mockRepo{},
			wantErr: domain.ErrInvalidAddress,
		},
		{
			name: "repo error",
			filter: domain.TransactionFilter{
				Hash:  ptr(validHash()),
				Limit: ptr(10),
			},
			repo:    &mockRepo{txErr: repoErr},
			wantErr: repoErr,
		},
		{
			// Limit is nil — triggers the default path inside GetTransactionsByFilter.
			name:    "txs found",
			filter:  domain.TransactionFilter{Hash: ptr(validHash())},
			repo:    &mockRepo{txsResult: sampleTxs},
			wantErr: nil,
			wantLen: len(sampleTxs),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetTransactionsByFilter(tc.filter)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if len(got) != tc.wantLen {
				t.Errorf("txs length: got %d, want %d", len(got), tc.wantLen)
			}
		})
	}
}

// ── GetEventByTxHashLogIndex ─────────────────────────────────────────────────

func TestGetEventByTxHashLogIndex(t *testing.T) {
	sampleEvent := domain.Event{BlockId: 1, TxHash: validHash(), LogIndex: 2, Emitter: validAddr()}
	repoErr := errors.New("db error")

	tests := []struct {
		name      string
		hash      string
		logIndex  int
		repo      *mockRepo
		wantErr   error
		wantEvent *domain.Event
	}{
		{
			name:     "invalid hash",
			hash:     "short",
			logIndex: 0,
			repo:     &mockRepo{},
			wantErr:  domain.ErrInvalidHash,
		},
		{
			name:     "negative log index",
			hash:     validHash(),
			logIndex: -1,
			repo:     &mockRepo{},
			wantErr:  domain.ErrInvalidId,
		},
		{
			name:     "repo error",
			hash:     validHash(),
			logIndex: 0,
			repo:     &mockRepo{eventErr: repoErr},
			wantErr:  repoErr,
		},
		{
			name:      "event found",
			hash:      validHash(),
			logIndex:  2,
			repo:      &mockRepo{eventsResult: []domain.Event{sampleEvent}},
			wantErr:   nil,
			wantEvent: &sampleEvent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetEventByTxHashLogIndex(tc.hash, tc.logIndex)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if got.TxHash != tc.wantEvent.TxHash || got.LogIndex != tc.wantEvent.LogIndex {
				t.Errorf("event mismatch: got %+v, want %+v", *got, *tc.wantEvent)
			}
		})
	}
}

// ── GetTransactionsByBatchBlocksId ───────────────────────────────────────────

func TestGetTransactionsByBatchBlocksId(t *testing.T) {
	repoErr := errors.New("db error")

	// Two txs for block 10, one tx for block 11.
	txBlock10a := domain.Transaction{BlockId: 10, Hash: validHash(), From: validAddr()}
	txBlock10b := domain.Transaction{BlockId: 10, Hash: validHash(), From: validAddr()}
	txBlock11 := domain.Transaction{BlockId: 11, Hash: validHash(), From: validAddr()}

	tests := []struct {
		name     string
		blockIDs []uint64
		withTxs  bool
		repo     *mockRepo
		wantErr  error
		wantMap  map[uint64]int // blockID → expected tx count
	}{
		{
			// withTxs=false: repo is never called, result is an empty map.
			name:     "txs not requested",
			blockIDs: []uint64{10, 11},
			withTxs:  false,
			repo:     &mockRepo{},
			wantErr:  nil,
			wantMap:  map[uint64]int{},
		},
		{
			name:     "repo error",
			blockIDs: []uint64{10, 11},
			withTxs:  true,
			repo:     &mockRepo{txErr: repoErr},
			wantErr:  repoErr,
		},
		{
			// Repo returns txs flat; the function must group them by BlockId.
			name:     "routing by block id",
			blockIDs: []uint64{10, 11},
			withTxs:  true,
			repo: &mockRepo{
				txsResult: []domain.Transaction{txBlock10a, txBlock10b, txBlock11},
			},
			wantErr: nil,
			wantMap: map[uint64]int{10: 2, 11: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetTransactionsByBatchBlocksId(tc.blockIDs, tc.withTxs)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			for blockID, wantCount := range tc.wantMap {
				if gotCount := len(got[blockID]); gotCount != wantCount {
					t.Errorf("block %d: got %d txs, want %d", blockID, gotCount, wantCount)
				}
			}
		})
	}

	_ = txBlock10a
	_ = txBlock10b
	_ = txBlock11
}

// ── GetEventsByBatchTxsHash ──────────────────────────────────────────────────

func TestGetEventsByBatchTxsHash(t *testing.T) {
	repoErr := errors.New("db error")

	hash1 := "0x" + "1111111111111111111111111111111111111111111111111111111111111111"
	hash2 := "0x" + "2222222222222222222222222222222222222222222222222222222222222222"

	// Two events for hash1, one for hash2.
	ev1a := domain.Event{TxHash: hash1, LogIndex: 0}
	ev1b := domain.Event{TxHash: hash1, LogIndex: 1}
	ev2 := domain.Event{TxHash: hash2, LogIndex: 0}

	tests := []struct {
		name    string
		hashes  []string
		repo    *mockRepo
		wantErr error
		wantMap map[string]int // txHash → expected event count
	}{
		{
			name:    "repo error",
			hashes:  []string{hash1, hash2},
			repo:    &mockRepo{eventErr: repoErr},
			wantErr: repoErr,
		},
		{
			name:   "empty result",
			hashes: []string{hash1},
			repo:   &mockRepo{},
			wantMap: map[string]int{},
		},
		{
			// Repo returns events flat; the function must group them by TxHash.
			name:   "routing by tx hash",
			hashes: []string{hash1, hash2},
			repo: &mockRepo{
				eventsResult: []domain.Event{ev1a, ev1b, ev2},
			},
			wantErr: nil,
			wantMap: map[string]int{hash1: 2, hash2: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService(tc.repo)
			got, err := svc.GetEventsByBatchTxsHash(tc.hashes)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			for hash, wantCount := range tc.wantMap {
				if gotCount := len(got[hash]); gotCount != wantCount {
					t.Errorf("hash %s: got %d events, want %d", hash, gotCount, wantCount)
				}
			}
		})
	}

}
