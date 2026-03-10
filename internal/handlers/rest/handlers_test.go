package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
)

type serviceMock struct {
	block     *domain.BlockTxs
	blocksTxs []domain.BlockTxs
	events    []domain.Event
	event     *domain.Event
	txs       []domain.Transaction

	err error
}

func newServiceMock(block *domain.BlockTxs, blocksTxs []domain.BlockTxs, events []domain.Event, txs []domain.Transaction, err error) *serviceMock {
	return &serviceMock{
		block:     block,
		blocksTxs: blocksTxs,
		events:    events,
		txs:       txs,
		err:       err,
	}
}

func (service serviceMock) GetBlockByHash(_ context.Context, hash string, tx bool) (*domain.BlockTxs, error) {
	return service.block, service.err
}

func (service serviceMock) GetBlockById(_ context.Context, id uint64, tx bool) (*domain.BlockTxs, error) {
	return service.block, service.err
}

func (service serviceMock) GetBlocksWithOffset(_ context.Context, fromId, offset uint64, tx bool) ([]domain.BlockTxs, error) {
	return service.blocksTxs, service.err
}

func (service serviceMock) GetBlocksByRangeTime(_ context.Context, from, to uint64, tx bool) ([]domain.BlockTxs, error) {
	return service.blocksTxs, service.err
}

func (service serviceMock) GetTransactionsByFilter(_ context.Context, filter domain.TransactionFilter) ([]domain.Transaction, error) {
	return service.txs, service.err
}

func (service serviceMock) GetEventsByFilter(_ context.Context, filter domain.EventFilter) ([]domain.Event, error) {
	return service.events, service.err
}

func (service serviceMock) GetEventByTxHashLogIndex(_ context.Context, hash string, logIndex int) (*domain.Event, error) {
	return service.event, service.err
}

func (service serviceMock) GetTransactionsByBatchBlocksId(_ context.Context, blockIDs []uint64, tx bool) (map[uint64][]domain.Transaction, error) {
	return nil, service.err
}

func (service serviceMock) GetEventsByBatchTxsHash(_ context.Context, txsHash []string) (map[string][]domain.Event, error) {
	return nil, service.err
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

func newHandler(service *serviceMock) *Handler {
	return NewHandler(service)
}

func TestGetBlock(t *testing.T) {
	blockTxs := &domain.BlockTxs{
		Block: domain.Block{Id: 42, Hash: validHash()},
		Txs: []domain.Transaction{
			{BlockId: 42, Hash: validHash(), From: "0x" + "abcdef1234567890abcdef1234567890abcdef12"},
		},
	}

	tests := []struct {
		name       string
		service    serviceMock
		args       string
		w          *httptest.ResponseRecorder
		wantStatus int
		wantEmpty  bool
	}{
		{
			name:       "happy path",
			service:    serviceMock{block: blockTxs},
			args:       "id=42&tx=yes",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusOK,
			wantEmpty:  false,
		},
		{
			name:       "invalid params — no query group",
			service:    serviceMock{},
			args:       "tx=yes",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "invalid params — conflicting groups",
			service:    serviceMock{},
			args:       "id=42&hash=" + validHash(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "not found",
			service:    serviceMock{err: domain.ErrNotFound},
			args:       "id=42",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusNotFound,
			wantEmpty:  true,
		},
		{
			name:       "service error",
			service:    serviceMock{err: errors.New("db error")},
			args:       "id=42",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusInternalServerError,
			wantEmpty:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rStr := "/api/blocks?" + test.args
			r := httptest.NewRequest("GET", rStr, nil)

			var b []domain.BlockTxs
			handler := NewHandler(test.service)

			handler.GetBlock(test.w, r)

			res := test.w.Result()
			defer res.Body.Close()

			if test.wantStatus != res.StatusCode {
				t.Errorf("invalid status code")
			}
			if test.wantEmpty == false {
				json.NewDecoder(res.Body).Decode(&b)
				if b[0].Block != blockTxs.Block {
					t.Errorf("invalid block data")
				}
				for i, tx := range b[0].Txs {
					if tx != blockTxs.Txs[i] {
						t.Errorf("invalid txs data")
					}
				}
			}
		})
	}
}

// ── TestGetTransaction ────────────────────────────────────────────────────────

func TestGetTransaction(t *testing.T) {
	sampleTx := domain.Transaction{BlockId: 1, Hash: validHash(), From: validAddr()}

	tests := []struct {
		name       string
		service    serviceMock
		args       string
		w          *httptest.ResponseRecorder
		wantStatus int
		wantEmpty  bool
	}{
		{
			name:       "happy path — by hash",
			service:    serviceMock{txs: []domain.Transaction{sampleTx}},
			args:       "hash=" + validHash(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusOK,
			wantEmpty:  false,
		},
		{
			name:       "invalid params — no params",
			service:    serviceMock{},
			args:       "",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "invalid params — fromBlock without toBlock",
			service:    serviceMock{},
			args:       "fromBlock=1",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "invalid params — address filter without range",
			service:    serviceMock{},
			args:       "from=" + validAddr(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "not found",
			service:    serviceMock{err: domain.ErrNotFound},
			args:       "hash=" + validHash(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusNotFound,
			wantEmpty:  true,
		},
		{
			name:       "service error",
			service:    serviceMock{err: errors.New("db error")},
			args:       "hash=" + validHash(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusInternalServerError,
			wantEmpty:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/transactions?"+test.args, nil)
			handler := NewHandler(test.service)

			handler.GetTransaction(test.w, r)

			res := test.w.Result()
			defer res.Body.Close()

			if test.wantStatus != res.StatusCode {
				t.Errorf("status: got %d, want %d", res.StatusCode, test.wantStatus)
			}
			if !test.wantEmpty {
				var got []domain.Transaction
				json.NewDecoder(res.Body).Decode(&got)
				if got[0].Hash != sampleTx.Hash {
					t.Errorf("tx hash: got %s, want %s", got[0].Hash, sampleTx.Hash)
				}
			}
		})
	}
}

// ── TestGetEvent ──────────────────────────────────────────────────────────────

func TestGetEvent(t *testing.T) {
	sampleEvent := domain.Event{BlockId: 1, TxHash: validHash(), LogIndex: 0, Emitter: validAddr()}

	tests := []struct {
		name       string
		service    serviceMock
		args       string
		w          *httptest.ResponseRecorder
		wantStatus int
		wantEmpty  bool
	}{
		{
			name:       "happy path — by address",
			service:    serviceMock{events: []domain.Event{sampleEvent}},
			args:       "address=" + validAddr(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusOK,
			wantEmpty:  false,
		},
		{
			name:       "invalid params — neither address nor topics",
			service:    serviceMock{},
			args:       "fromBlock=1&toBlock=10",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "invalid params — fromBlock without toBlock",
			service:    serviceMock{},
			args:       "address=" + validAddr() + "&fromBlock=1",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "service error",
			service:    serviceMock{err: errors.New("db error")},
			args:       "address=" + validAddr(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusInternalServerError,
			wantEmpty:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/events?"+test.args, nil)
			handler := NewHandler(test.service)

			handler.GetEvent(test.w, r)

			res := test.w.Result()
			defer res.Body.Close()

			if test.wantStatus != res.StatusCode {
				t.Errorf("status: got %d, want %d", res.StatusCode, test.wantStatus)
			}
			if !test.wantEmpty {
				var got []domain.Event
				json.NewDecoder(res.Body).Decode(&got)
				if got[0].TxHash != sampleEvent.TxHash {
					t.Errorf("event txHash: got %s, want %s", got[0].TxHash, sampleEvent.TxHash)
				}
			}
		})
	}
}

// ── TestGetEventByTxLog ───────────────────────────────────────────────────────

func TestGetEventByTxLog(t *testing.T) {
	sampleEvent := &domain.Event{BlockId: 1, TxHash: validHash(), LogIndex: 2, Emitter: validAddr()}

	tests := []struct {
		name       string
		service    serviceMock
		args       string
		w          *httptest.ResponseRecorder
		wantStatus int
		wantEmpty  bool
	}{
		{
			name:       "happy path",
			service:    serviceMock{event: sampleEvent},
			args:       "txHash=" + validHash() + "&logIndex=2",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusOK,
			wantEmpty:  false,
		},
		{
			name:       "invalid params — missing txHash",
			service:    serviceMock{},
			args:       "logIndex=2",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "invalid params — missing logIndex",
			service:    serviceMock{},
			args:       "txHash=" + validHash(),
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "invalid params — non-numeric logIndex",
			service:    serviceMock{},
			args:       "txHash=" + validHash() + "&logIndex=abc",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusBadRequest,
			wantEmpty:  true,
		},
		{
			name:       "not found",
			service:    serviceMock{err: domain.ErrNotFound},
			args:       "txHash=" + validHash() + "&logIndex=2",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusNotFound,
			wantEmpty:  true,
		},
		{
			name:       "service error",
			service:    serviceMock{err: errors.New("db error")},
			args:       "txHash=" + validHash() + "&logIndex=2",
			w:          httptest.NewRecorder(),
			wantStatus: http.StatusInternalServerError,
			wantEmpty:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/events/log?"+test.args, nil)
			handler := NewHandler(test.service)

			handler.GetEventByTxLog(test.w, r)

			res := test.w.Result()
			defer res.Body.Close()

			if test.wantStatus != res.StatusCode {
				t.Errorf("status: got %d, want %d", res.StatusCode, test.wantStatus)
			}
			if !test.wantEmpty {
				var got domain.Event
				json.NewDecoder(res.Body).Decode(&got)
				if got.TxHash != sampleEvent.TxHash || got.LogIndex != sampleEvent.LogIndex {
					t.Errorf("event mismatch: got %+v, want %+v", got, *sampleEvent)
				}
			}
		})
	}
}
