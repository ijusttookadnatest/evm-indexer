package fetcher

import (
	"context"
	"errors"
	"testing"

	"github.com/cenkalti/backoff/v5"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/time/rate"
)

type mockRPCError struct {
	code int
	msg  string
}

func (e *mockRPCError) Error() string  { return e.msg }
func (e *mockRPCError) ErrorCode() int { return e.code }

func TestWrapRetryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantNil  bool
		wantPerm bool
	}{
		{name: "nil returns nil", err: nil, wantNil: true},
		{name: "non-RPC error is retriable", err: errors.New("network timeout"), wantPerm: false},
		{name: "invalid request (-32600) is permanent", err: &mockRPCError{code: -32600}, wantPerm: true},
		{name: "method not found (-32601) is permanent", err: &mockRPCError{code: -32601}, wantPerm: true},
		{name: "invalid params (-32602) is permanent", err: &mockRPCError{code: -32602}, wantPerm: true},
		{name: "just outside permanent range (-32603) is retriable", err: &mockRPCError{code: -32603}, wantPerm: false},
		{name: "server error (-32000) is retriable", err: &mockRPCError{code: -32000}, wantPerm: false},
		{name: "rate limit (-32005) is retriable", err: &mockRPCError{code: -32005}, wantPerm: false},
		{name: "parse error (-32700) is retriable", err: &mockRPCError{code: -32700}, wantPerm: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapRetryError(tt.err)
			if tt.wantNil {
				if got != nil {
					t.Errorf("want nil, got %v", got)
				}
				return
			}
			var permErr *backoff.PermanentError
			isPerm := errors.As(got, &permErr)
			if isPerm != tt.wantPerm {
				t.Errorf("permanent: want %v, got %v (err=%v)", tt.wantPerm, isPerm, got)
			}
		})
	}
}

// seqMockEVMClient retourne une séquence d'erreurs à chaque appel CallContext.
// Une fois la séquence épuisée, le dernier élément est répété.
type seqMockEVMClient struct {
	mockEVMClient
	callErrs  []error
	callCount int
}

func (m *seqMockEVMClient) CallContext(_ context.Context, result interface{}, _ string, _ ...interface{}) error {
	idx := m.callCount
	if idx >= len(m.callErrs) {
		idx = len(m.callErrs) - 1
	}
	m.callCount++
	if err := m.callErrs[idx]; err != nil {
		return err
	}
	*(result.(*RPCBlock)) = m.block
	return nil
}

func TestFetchBlock_Retry(t *testing.T) {
	tests := []struct {
		name          string
		client        *seqMockEVMClient
		wantErr       bool
		wantCallCount int
	}{
		{
			name: "permanent error stops immediately (1 call)",
			client: &seqMockEVMClient{
				callErrs: []error{&mockRPCError{code: -32601, msg: "method not found"}},
			},
			wantErr:       true,
			wantCallCount: 1,
		},
		{
			name: "transient error retries and succeeds (2 calls)",
			client: &seqMockEVMClient{
				mockEVMClient: mockEVMClient{
					block:    RPCBlock{Number: 1},
					receipts: []*types.Receipt{},
				},
				callErrs: []error{errors.New("transient"), nil},
			},
			wantErr:       false,
			wantCallCount: 2,
		},
		{
			name: "rate limit (-32000) retries and succeeds (2 calls)",
			client: &seqMockEVMClient{
				mockEVMClient: mockEVMClient{
					block:    RPCBlock{Number: 1},
					receipts: []*types.Receipt{},
				},
				callErrs: []error{&mockRPCError{code: -32000, msg: "too many requests"}, nil},
			},
			wantErr:       false,
			wantCallCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Fetcher{clientHTTP: tt.client, rateLimiter: rate.NewLimiter(rate.Inf, 1)}
			_, err := b.FetchBlock(context.Background(), 1)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tt.client.callCount != tt.wantCallCount {
				t.Errorf("callCount: want %d, got %d", tt.wantCallCount, tt.client.callCount)
			}
		})
	}
}

// seqMockReceiptsClient sequences errors for BlockReceipts; CallContext always succeeds.
type seqMockReceiptsClient struct {
	mockEVMClient
	receiptsErrs  []error
	receiptsCount int
}

func (m *seqMockReceiptsClient) BlockReceipts(_ context.Context, _ rpc.BlockNumberOrHash) ([]*types.Receipt, error) {
	idx := m.receiptsCount
	if idx >= len(m.receiptsErrs) {
		idx = len(m.receiptsErrs) - 1
	}
	m.receiptsCount++
	if err := m.receiptsErrs[idx]; err != nil {
		return nil, err
	}
	return m.mockEVMClient.receipts, nil
}

func TestFetchBlockReceipts_Retry(t *testing.T) {
	tests := []struct {
		name              string
		client            *seqMockReceiptsClient
		wantErr           bool
		wantReceiptsCount int
	}{
		{
			name: "permanent error on BlockReceipts stops immediately (1 call)",
			client: &seqMockReceiptsClient{
				receiptsErrs: []error{&mockRPCError{code: -32601, msg: "method not found"}},
			},
			wantErr:           true,
			wantReceiptsCount: 1,
		},
		{
			name: "transient error on BlockReceipts retries and succeeds (2 calls)",
			client: &seqMockReceiptsClient{
				mockEVMClient: mockEVMClient{
					block:    RPCBlock{Number: 1},
					receipts: []*types.Receipt{},
				},
				receiptsErrs: []error{errors.New("network glitch"), nil},
			},
			wantErr:           false,
			wantReceiptsCount: 2,
		},
		{
			name: "rate limit (-32000) on BlockReceipts retries and succeeds (2 calls)",
			client: &seqMockReceiptsClient{
				mockEVMClient: mockEVMClient{
					block:    RPCBlock{Number: 1},
					receipts: []*types.Receipt{},
				},
				receiptsErrs: []error{&mockRPCError{code: -32000, msg: "too many requests"}, nil},
			},
			wantErr:           false,
			wantReceiptsCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Fetcher{clientHTTP: tt.client, rateLimiter: rate.NewLimiter(rate.Inf, 1)}
			_, err := b.FetchBlock(context.Background(), 1)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tt.client.receiptsCount != tt.wantReceiptsCount {
				t.Errorf("receiptsCount: want %d, got %d", tt.wantReceiptsCount, tt.client.receiptsCount)
			}
		})
	}
}

// seqMockBlockNumClient sequences errors for BlockNumber; other methods always succeed.
type seqMockBlockNumClient struct {
	mockEVMClient
	numErrs  []error
	numCount int
}

func (m *seqMockBlockNumClient) BlockNumber(_ context.Context) (uint64, error) {
	idx := m.numCount
	if idx >= len(m.numErrs) {
		idx = len(m.numErrs) - 1
	}
	m.numCount++
	if err := m.numErrs[idx]; err != nil {
		return 0, err
	}
	return m.mockEVMClient.lastBlockId, nil
}

func TestGetLastBlockId_Retry(t *testing.T) {
	tests := []struct {
		name          string
		client        *seqMockBlockNumClient
		wantErr       bool
		wantCallCount int
		wantBlockId   uint64
	}{
		{
			name:          "permanent error stops immediately (1 call)",
			client:        &seqMockBlockNumClient{numErrs: []error{&mockRPCError{code: -32601, msg: "method not found"}}},
			wantErr:       true,
			wantCallCount: 1,
		},
		{
			name: "transient error retries and succeeds (2 calls)",
			client: &seqMockBlockNumClient{
				mockEVMClient: mockEVMClient{lastBlockId: 42},
				numErrs:       []error{errors.New("network glitch"), nil},
			},
			wantErr:       false,
			wantCallCount: 2,
			wantBlockId:   42,
		},
		{
			name: "rate limit (-32000) retries and succeeds (2 calls)",
			client: &seqMockBlockNumClient{
				mockEVMClient: mockEVMClient{lastBlockId: 99},
				numErrs:       []error{&mockRPCError{code: -32000, msg: "too many requests"}, nil},
			},
			wantErr:       false,
			wantCallCount: 2,
			wantBlockId:   99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Fetcher{clientHTTP: tt.client, rateLimiter: rate.NewLimiter(rate.Inf, 1)}
			got, err := b.GetLastBlockId()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if !tt.wantErr && got != tt.wantBlockId {
				t.Errorf("blockId: want %d, got %d", tt.wantBlockId, got)
			}
			if tt.client.numCount != tt.wantCallCount {
				t.Errorf("callCount: want %d, got %d", tt.wantCallCount, tt.client.numCount)
			}
		})
	}
}
