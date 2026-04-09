package service

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
)

const (
	sigTransfer       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	sigTransferSingle = "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"
	sigTransferBatch  = "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb"
)

func mustPackBatch(ids, values []*big.Int) string {
	uint256Arr, _ := abi.NewType("uint256[]", "", nil)
	args := abi.Arguments{{Type: uint256Arr}, {Type: uint256Arr}}
	packed, err := args.Pack(ids, values)
	if err != nil {
		panic(err)
	}
	return "0x" + hex.EncodeToString(packed)
}

func bigHex(n int64) string {
	b := make([]byte, 32)
	new(big.Int).SetInt64(n).FillBytes(b)
	return "0x" + hex.EncodeToString(b)
}

// ── findStandard ─────────────────────────────────────────────────────────────

func TestFindStandard(t *testing.T) {
	tests := []struct {
		name   string
		topics []string
		want   int
	}{
		{
			name:   "ERC20: 3 topics, Transfer sig",
			topics: []string{sigTransfer, "0xfrom", "0xto"},
			want:   ERC20,
		},
		{
			name:   "ERC721: 4 topics, Transfer sig",
			topics: []string{sigTransfer, "0xfrom", "0xto", "0xtokenId"},
			want:   ERC721,
		},
		{
			name:   "ERC1155 Single",
			topics: []string{sigTransferSingle, "0xop", "0xfrom", "0xto"},
			want:   ERC1155_SINGLE,
		},
		{
			name:   "ERC1155 Batch",
			topics: []string{sigTransferBatch, "0xop", "0xfrom", "0xto"},
			want:   ERC1155_BATCH,
		},
		{
			name:   "empty topics returns -1",
			topics: []string{},
			want:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findStandard(domain.Log{Topics: tt.topics})
			if got != tt.want {
				t.Errorf("findStandard() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ── extractBalanceEntriesFromLog ─────────────────────────────────────────────

func TestExtractBalanceEntriesFromLog(t *testing.T) {
	contract := "0xcontract"
	from := "0xfrom"
	to := "0xto"
	operator := "0xoperator"
	amount := big.NewInt(500)

	t.Run("ERC20: two entries with correct delta", func(t *testing.T) {
		entries := extractBalanceEntriesFromLog(domain.Log{
			Emitter: contract,
			Topics:  []string{sigTransfer, from, to},
			Datas:   bigHex(500),
		})
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Amount.Cmp(new(big.Int).Neg(amount)) != 0 {
			t.Errorf("from amount: want %s, got %s", new(big.Int).Neg(amount), entries[0].Amount)
		}
		if entries[1].Amount.Cmp(amount) != 0 {
			t.Errorf("to amount: want %s, got %s", amount, entries[1].Amount)
		}
		if entries[0].TokenId != "" || entries[1].TokenId != "" {
			t.Error("ERC20 entries should have empty TokenId")
		}
	})

	t.Run("ERC721: two entries with tokenId and amount ±1", func(t *testing.T) {
		tokenId := "0xtokenId42"
		entries := extractBalanceEntriesFromLog(domain.Log{
			Emitter: contract,
			Topics:  []string{sigTransfer, from, to, tokenId},
			Datas:   "",
		})
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].TokenId != tokenId || entries[1].TokenId != tokenId {
			t.Errorf("TokenId mismatch: got %s / %s", entries[0].TokenId, entries[1].TokenId)
		}
		if entries[0].Amount.Cmp(big.NewInt(-1)) != 0 {
			t.Errorf("from amount: want -1, got %s", entries[0].Amount)
		}
		if entries[1].Amount.Cmp(big.NewInt(1)) != 0 {
			t.Errorf("to amount: want 1, got %s", entries[1].Amount)
		}
	})

	t.Run("ERC1155 Single: tokenId and value from data", func(t *testing.T) {
		tokenIdBytes := make([]byte, 32)
		valueBytes := make([]byte, 32)
		big.NewInt(42).FillBytes(tokenIdBytes)
		big.NewInt(10).FillBytes(valueBytes)
		data := "0x" + hex.EncodeToString(append(tokenIdBytes, valueBytes...))

		entries := extractBalanceEntriesFromLog(domain.Log{
			Emitter: contract,
			Topics:  []string{sigTransferSingle, operator, from, to},
			Datas:   data,
		})
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].WalletAddress != from {
			t.Errorf("from wallet: want %s, got %s", from, entries[0].WalletAddress)
		}
		if entries[0].Amount.Cmp(big.NewInt(-10)) != 0 {
			t.Errorf("from amount: want -10, got %s", entries[0].Amount)
		}
		if entries[1].Amount.Cmp(big.NewInt(10)) != 0 {
			t.Errorf("to amount: want 10, got %s", entries[1].Amount)
		}
	})

	t.Run("ERC1155 Batch: one entry per (tokenId, value) pair", func(t *testing.T) {
		ids := []*big.Int{big.NewInt(1), big.NewInt(2)}
		values := []*big.Int{big.NewInt(100), big.NewInt(200)}
		data := mustPackBatch(ids, values)

		entries := extractBalanceEntriesFromLog(domain.Log{
			Emitter: contract,
			Topics:  []string{sigTransferBatch, operator, from, to},
			Datas:   data,
		})
		if len(entries) != 4 {
			t.Fatalf("expected 4 entries (2 pairs), got %d", len(entries))
		}
		if entries[0].Amount.Cmp(big.NewInt(-100)) != 0 {
			t.Errorf("pair[0] from amount: want -100, got %s", entries[0].Amount)
		}
		if entries[1].Amount.Cmp(big.NewInt(100)) != 0 {
			t.Errorf("pair[0] to amount: want 100, got %s", entries[1].Amount)
		}
	})

	t.Run("malformed data returns nil", func(t *testing.T) {
		entries := extractBalanceEntriesFromLog(domain.Log{
			Emitter: contract,
			Topics:  []string{sigTransferSingle, operator, from, to},
			Datas:   "0x",
		})
		if entries != nil {
			t.Errorf("expected nil for short data, got %v", entries)
		}
	})
}
