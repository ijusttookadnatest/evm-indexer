package fetcher

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type mockEVMClient struct {
	block       RPCBlock
	receipts    []*types.Receipt
	lastBlockId uint64
	callErr     error
	receiptsErr error
	blockNumErr error
}

func (m *mockEVMClient) CallContext(_ context.Context, result interface{}, _ string, _ ...interface{}) error {
	if m.callErr != nil {
		return m.callErr
	}
	*(result.(*RPCBlock)) = m.block
	return nil
}

func (m *mockEVMClient) BlockReceipts(_ context.Context, _ rpc.BlockNumberOrHash) ([]*types.Receipt, error) {
	if m.receiptsErr != nil {
		return nil, m.receiptsErr
	}
	return m.receipts, nil
}

func (m *mockEVMClient) BlockNumber(_ context.Context) (uint64, error) {
	return m.lastBlockId, m.blockNumErr
}

// --- FetchBlock tests ---

func TestFetchBlock(t *testing.T) {
	toAddr := common.HexToAddress("0xBob")
	contract := common.HexToAddress("0xContract")

	tests := []struct {
		name           string
		client         *mockEVMClient
		blockId        uint64
		wantErr        bool
		wantBlockId    uint64
		wantTxCount    int
		wantEventCount int
	}{
		{
			name: "empty block (no txs no events)",
			client: &mockEVMClient{
				block:    RPCBlock{Hash: common.HexToHash("0xabc"), Number: hexutil.Uint64(100)},
				receipts: []*types.Receipt{},
			},
			blockId:     100,
			wantBlockId: 100,
		},
		{
			name: "block with tx and event",
			client: &mockEVMClient{
				block: RPCBlock{
					Hash:   common.HexToHash("0xdef"),
					Number: hexutil.Uint64(200),
					Transactions: []RPCTransaction{
						{Hash: common.HexToHash("0xtx1"), From: common.HexToAddress("0xAlice"), To: &toAddr},
					},
				},
				receipts: []*types.Receipt{
					{
						BlockNumber: big.NewInt(200), CumulativeGasUsed: 21000, Status: 1,
						Logs: []*types.Log{
							{BlockNumber: 200, Address: contract, Topics: []common.Hash{common.HexToHash("0xSig")}},
						},
					},
				},
			},
			blockId:        200,
			wantBlockId:    200,
			wantTxCount:    1,
			wantEventCount: 1,
		},
		{
			name:    "contract creation tx (To=nil)",
			blockId: 300,
			client: &mockEVMClient{
				block: RPCBlock{
					Number: hexutil.Uint64(300),
					Transactions: []RPCTransaction{
						{Hash: common.HexToHash("0xdeploy"), From: common.HexToAddress("0xDeployer"), To: nil},
					},
				},
				receipts: []*types.Receipt{
					{BlockNumber: big.NewInt(300), CumulativeGasUsed: 500000, Status: 1},
				},
			},
			wantBlockId: 300,
			wantTxCount: 1,
		},
		{
			name:    "CallContext error propagates",
			client:  &mockEVMClient{callErr: errors.New("rpc timeout")},
			blockId: 100,
			wantErr: true,
		},
		{
			name: "BlockReceipts error propagates",
			client: &mockEVMClient{
				block:       RPCBlock{Number: hexutil.Uint64(100)},
				receiptsErr: errors.New("receipts unavailable"),
			},
			blockId: 100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Backfiller{client: tt.client}
			got, err := b.FetchBlock(tt.blockId)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got.Block.Id != tt.wantBlockId {
				t.Errorf("BlockId: want %d, got %d", tt.wantBlockId, got.Block.Id)
			}
			if len(got.Txs) != tt.wantTxCount {
				t.Errorf("Txs: want %d, got %d", tt.wantTxCount, len(got.Txs))
			}
			if len(got.Events) != tt.wantEventCount {
				t.Errorf("Events: want %d, got %d", tt.wantEventCount, len(got.Events))
			}
		})
	}
}

// --- Extract function unit tests ---

func TestExtractBlock(t *testing.T) {
	tests := []struct {
		name        string
		body        RPCBlock
		wantId      uint64
		wantHash    string
		wantGasUsed uint64
	}{
		{
			name: "maps all fields correctly",
			body: RPCBlock{
				Hash:    common.HexToHash("0xabc"),
				Number:  hexutil.Uint64(100),
				GasUsed: hexutil.Uint64(15000000),
			},
			wantId:      100,
			wantHash:    common.HexToHash("0xabc").Hex(),
			wantGasUsed: 15000000,
		},
		{
			name:        "genesis block",
			body:        RPCBlock{Number: hexutil.Uint64(0)},
			wantId:      0,
			wantGasUsed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBlock(tt.body)
			if got.Id != tt.wantId {
				t.Errorf("Id: want %d, got %d", tt.wantId, got.Id)
			}
			if tt.wantHash != "" && got.Hash != tt.wantHash {
				t.Errorf("Hash: want %s, got %s", tt.wantHash, got.Hash)
			}
			if got.GasUsed != tt.wantGasUsed {
				t.Errorf("GasUsed: want %d, got %d", tt.wantGasUsed, got.GasUsed)
			}
		})
	}
}

func TestExtractTopics(t *testing.T) {
	tests := []struct {
		name  string
		input []common.Hash
		want  []string
	}{
		{name: "nil input", input: nil, want: nil},
		{
			name:  "single topic",
			input: []common.Hash{common.HexToHash("0xdeadbeef")},
			want:  []string{common.HexToHash("0xdeadbeef").Hex()},
		},
		{
			name:  "multiple topics",
			input: []common.Hash{common.HexToHash("0x111"), common.HexToHash("0x222")},
			want:  []string{common.HexToHash("0x111").Hex(), common.HexToHash("0x222").Hex()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTopics(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("len: want %d, got %d", len(tt.want), len(got))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d]: want %s, got %s", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestExtractTransaction(t *testing.T) {
	toAddr := common.HexToAddress("0xBob")

	tests := []struct {
		name        string
		tx          RPCTransaction
		receipt     types.Receipt
		wantTo      *string
		wantStatus  uint64
		wantBlockId uint64
	}{
		{
			name:        "normal tx",
			tx:          RPCTransaction{Hash: common.HexToHash("0xtx1"), From: common.HexToAddress("0xAlice"), To: &toAddr},
			receipt:     types.Receipt{BlockNumber: big.NewInt(100), CumulativeGasUsed: 21000, Status: 1},
			wantTo:      func() *string { s := toAddr.Hex(); return &s }(),
			wantStatus:  1,
			wantBlockId: 100,
		},
		{
			name:        "failed tx",
			tx:          RPCTransaction{Hash: common.HexToHash("0xfailed"), From: common.HexToAddress("0xAlice"), To: &toAddr},
			receipt:     types.Receipt{BlockNumber: big.NewInt(300), Status: 0},
			wantTo:      func() *string { s := toAddr.Hex(); return &s }(),
			wantStatus:  0,
			wantBlockId: 300,
		},
		{
			name:        "contract creation (To=nil)",
			tx:          RPCTransaction{Hash: common.HexToHash("0xdeploy"), From: common.HexToAddress("0xDeployer"), To: nil},
			receipt:     types.Receipt{BlockNumber: big.NewInt(200), Status: 1},
			wantTo:      nil,
			wantStatus:  1,
			wantBlockId: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTransaction(tt.tx, tt.receipt)
			if tt.wantTo == nil && got.To != nil {
				t.Errorf("To: want nil, got %s", *got.To)
			}
			if tt.wantTo != nil && (got.To == nil || *got.To != *tt.wantTo) {
				t.Errorf("To: want %v, got %v", tt.wantTo, got.To)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status: want %d, got %d", tt.wantStatus, got.Status)
			}
			if got.BlockId != tt.wantBlockId {
				t.Errorf("BlockId: want %d, got %d", tt.wantBlockId, got.BlockId)
			}
		})
	}
}
