//go:build integration

package repository

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
)

func TestCreate_Integration(t *testing.T) {
	queryRepo := NewQueryRepository(testDB)
	indexerRepo := NewIndexerRepository(testDB)

	t.Run("create block with txs and events", func(t *testing.T) {
		truncateAll(t)

		to := "0xBob"
		block := domain.Block{
			Hash: "0xnew", Id: 200, ParentHash: "0xprev",
			GasLimit: 30000000, GasUsed: 10000000, Miner: "0xminer", Timestamp: 1800000000,
		}
		txs := []domain.Transaction{
			{BlockId: 200, Hash: "0xnewtx1", From: "0xAlice", To: &to, GasUsed: 21000},
		}
		events := []domain.Event{
			{BlockId: 200, LogIndex: 0, TxHash: "0xnewtx1", Emitter: "0xContract", Datas: "0xdata", Topics: []string{"0xTopic"}},
		}

		err := indexerRepo.Create(context.Background(), block, txs, events)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}

		// Verify block was persisted
		got, err := queryRepo.GetBlockById(context.Background(), 200)
		if err != nil {
			t.Fatalf("block should exist: %v", err)
		}
		if got.Hash != "0xnew" || got.Miner != "0xminer" {
			t.Errorf("invalid block data: %v", got)
		}
	})

	t.Run("create duplicate block is idempotent", func(t *testing.T) {
		truncateAll(t)
		seedFixtures(t)

		block := domain.Block{
			Hash: "0xblock100", Id: 100,
		}
		err := indexerRepo.Create(context.Background(), block, nil, nil)
		if err != nil {
			t.Fatalf("duplicate block should be silently ignored, got: %v", err)
		}
	})
}

func TestBulkCreate_Integration(t *testing.T) {
	queryRepo := NewQueryRepository(testDB)
	indexerRepo := NewIndexerRepository(testDB)

	to := "0xBob"

	t.Run("inserts all blocks, txs and events", func(t *testing.T) {
		truncateAll(t)

		items := []domain.BlockTxsEvents{
			{
				Block: domain.Block{Hash: "0xA", Id: 1, ParentHash: "0x0", GasLimit: 1000, GasUsed: 500, Miner: "0xM1", Timestamp: 100},
				Txs: []domain.Transaction{
					{BlockId: 1, Hash: "0xtxA1", From: "0xAlice", To: &to, GasUsed: 21000},
					{BlockId: 1, Hash: "0xtxA2", From: "0xAlice", To: nil, GasUsed: 50000},
				},
				Events: []domain.Event{
					{BlockId: 1, LogIndex: 0, TxHash: "0xtxA1", Emitter: "0xC", Datas: "0xd", Topics: []string{"0xSig", "0xAlice"}},
				},
			},
			{
				Block: domain.Block{Hash: "0xB", Id: 2, ParentHash: "0xA", GasLimit: 1000, GasUsed: 200, Miner: "0xM2", Timestamp: 200},
				Txs: []domain.Transaction{
					{BlockId: 2, Hash: "0xtxB1", From: "0xBob", To: &to, GasUsed: 21000},
				},
				Events: []domain.Event{},
			},
		}

		if err := indexerRepo.BulkCreate(context.Background(), items); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		blocks, err := queryRepo.GetBlocksByRangeId(context.Background(), 0, 3)
		if err != nil || len(blocks) != 2 {
			t.Fatalf("expected 2 blocks, got %d (err: %v)", len(blocks), err)
		}

		txs, err := queryRepo.GetTransactionsByBatchBlocksId(context.Background(), []uint64{1, 2})
		if err != nil || len(txs) != 3 {
			t.Fatalf("expected 3 txs, got %d (err: %v)", len(txs), err)
		}

		event, err := queryRepo.GetEventByTxHashLogIndex(context.Background(), "0xtxA1", 0)
		if err != nil {
			t.Fatalf("expected event, got err=%v", err)
		}
		if len(event.Topics) != 2 {
			t.Fatalf("expected 2 topics, got %v", event.Topics)
		}
	})

	t.Run("duplicate blocks are idempotent", func(t *testing.T) {
		truncateAll(t)
		seedFixtures(t)

		items := []domain.BlockTxsEvents{
			{Block: domain.Block{Hash: "0xblock100", Id: 100}, Txs: nil, Events: nil},
			{Block: domain.Block{Hash: "0xblock101", Id: 101}, Txs: nil, Events: nil},
		}

		if err := indexerRepo.BulkCreate(context.Background(), items); err != nil {
			t.Fatalf("duplicate blocks should be silently ignored, got: %v", err)
		}
	})

	t.Run("blocks with no txs or events", func(t *testing.T) {
		truncateAll(t)

		items := []domain.BlockTxsEvents{
			{Block: domain.Block{Hash: "0xEmpty1", Id: 10, ParentHash: "0x0", GasLimit: 1000, GasUsed: 0, Miner: "0xM", Timestamp: 300}},
			{Block: domain.Block{Hash: "0xEmpty2", Id: 11, ParentHash: "0xEmpty1", GasLimit: 1000, GasUsed: 0, Miner: "0xM", Timestamp: 400}},
		}

		if err := indexerRepo.BulkCreate(context.Background(), items); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		blocks, err := queryRepo.GetBlocksByRangeId(context.Background(), 9, 12)
		if err != nil || len(blocks) != 2 {
			t.Fatalf("expected 2 blocks, got %d (err: %v)", len(blocks), err)
		}
	})
}

func TestBackfillCursor_Integration(t *testing.T) {
	indexerRepo := NewIndexerRepository(testDB)

	t.Run("initial cursor is 0", func(t *testing.T) {
		cursor, err := indexerRepo.GetBackfillCursor(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cursor != 0 {
			t.Errorf("want 0, got %d", cursor)
		}
	})

	t.Run("update and read cursor", func(t *testing.T) {
		err := indexerRepo.UpdateBackfillCursor(context.Background(), 12345)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cursor, err := indexerRepo.GetBackfillCursor(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cursor != 12345 {
			t.Errorf("want 12345, got %d", cursor)
		}
		// reset
		_ = indexerRepo.UpdateBackfillCursor(context.Background(), 0)
	})
}

func TestResetBackfillCursor_Integration(t *testing.T) {
	indexerRepo := NewIndexerRepository(testDB)

	_ = indexerRepo.UpdateBackfillCursor(context.Background(), 12345)

	err := indexerRepo.ResetBackfillCursor(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cursor, err := indexerRepo.GetBackfillCursor(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cursor != 0 {
		t.Errorf("want 0 after reset, got %d", cursor)
	}
}

func TestGetLogsByTopic_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewIndexerRepository(testDB)
	ctx := context.Background()

	t.Run("match single topic in block range returns correct log", func(t *testing.T) {
		logs, err := repo.GetLogsByTopic(ctx, domain.LogFilter{
			Topics:    []string{"0xTransferSig"},
			FromBlock: 100,
			ToBlock:   100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("expected 1 log, got %d", len(logs))
		}
		if logs[0].Emitter != "0xContract" {
			t.Errorf("unexpected emitter: %v", logs[0].Emitter)
		}
		if logs[0].Id == 0 {
			t.Error("id should be set (non-zero)")
		}
		if logs[0].BlockId != 100 {
			t.Errorf("expected block_id 100, got %d", logs[0].BlockId)
		}
	})

	t.Run("match multiple topics in block range returns all", func(t *testing.T) {
		logs, err := repo.GetLogsByTopic(ctx, domain.LogFilter{
			Topics:    []string{"0xTransferSig", "0xApprovalSig"},
			FromBlock: 100,
			ToBlock:   100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 2 {
			t.Fatalf("expected 2 logs, got %d", len(logs))
		}
	})

	t.Run("block range excludes events outside range", func(t *testing.T) {
		logs, err := repo.GetLogsByTopic(ctx, domain.LogFilter{
			Topics:    []string{"0xTransferSig", "0xApprovalSig"},
			FromBlock: 101,
			ToBlock:   102,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 0 {
			t.Fatalf("expected 0 logs for blocks 101-102, got %d", len(logs))
		}
	})

	t.Run("no matching topic returns empty", func(t *testing.T) {
		logs, err := repo.GetLogsByTopic(ctx, domain.LogFilter{
			Topics:    []string{"0xUnknownSig"},
			FromBlock: 100,
			ToBlock:   102,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(logs) != 0 {
			t.Errorf("expected empty slice, got %d", len(logs))
		}
	})
}

func TestBatchUpsertBalance_Integration(t *testing.T) {
	ctx := context.Background()
	repo := NewIndexerRepository(testDB)

	readBalance := func(t *testing.T, wallet, token, tokenId string) string {
		t.Helper()
		var amount string
		err := testDB.QueryRowContext(ctx,
			`SELECT amount FROM wallet_balance WHERE wallet_address=$1 AND token_address=$2 AND token_id=$3`,
			wallet, token, tokenId,
		).Scan(&amount)
		if err != nil {
			t.Fatalf("failed to read balance: %v", err)
		}
		return amount
	}

	t.Run("inserts new entries", func(t *testing.T) {
		testDB.Exec("TRUNCATE wallet_balance")

		entries := []domain.BalanceEntry{
			{WalletAddress: "0xAlice", TokenAddress: "0xToken", TokenId: "", Amount: big.NewInt(100)},
			{WalletAddress: "0xBob", TokenAddress: "0xToken", TokenId: "", Amount: big.NewInt(50)},
		}
		if err := repo.BatchUpsertBalance(ctx, entries); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := readBalance(t, "0xAlice", "0xToken", ""); got != "100" {
			t.Errorf("Alice: want 100, got %s", got)
		}
		if got := readBalance(t, "0xBob", "0xToken", ""); got != "50" {
			t.Errorf("Bob: want 50, got %s", got)
		}
	})

	t.Run("upsert accumulates amounts on conflict", func(t *testing.T) {
		testDB.Exec("TRUNCATE wallet_balance")

		first := []domain.BalanceEntry{
			{WalletAddress: "0xAlice", TokenAddress: "0xToken", TokenId: "", Amount: big.NewInt(100)},
		}
		second := []domain.BalanceEntry{
			{WalletAddress: "0xAlice", TokenAddress: "0xToken", TokenId: "", Amount: big.NewInt(40)},
		}
		if err := repo.BatchUpsertBalance(ctx, first); err != nil {
			t.Fatalf("unexpected error on first upsert: %v", err)
		}
		if err := repo.BatchUpsertBalance(ctx, second); err != nil {
			t.Fatalf("unexpected error on second upsert: %v", err)
		}

		if got := readBalance(t, "0xAlice", "0xToken", ""); got != "140" {
			t.Errorf("want 140, got %s", got)
		}
	})

	t.Run("erc721 entry uses token_id", func(t *testing.T) {
		testDB.Exec("TRUNCATE wallet_balance")

		entries := []domain.BalanceEntry{
			{WalletAddress: "0xAlice", TokenAddress: "0xNFT", TokenId: "0x01", Amount: big.NewInt(1)},
		}
		if err := repo.BatchUpsertBalance(ctx, entries); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := readBalance(t, "0xAlice", "0xNFT", "0x01"); got != "1" {
			t.Errorf("want 1, got %s", got)
		}
	})

	t.Run("erc20 and erc721 same wallet are independent rows", func(t *testing.T) {
		testDB.Exec("TRUNCATE wallet_balance")

		entries := []domain.BalanceEntry{
			{WalletAddress: "0xAlice", TokenAddress: "0xToken", TokenId: "", Amount: big.NewInt(200)},
			{WalletAddress: "0xAlice", TokenAddress: "0xToken", TokenId: "0x01", Amount: big.NewInt(1)},
		}
		if err := repo.BatchUpsertBalance(ctx, entries); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := readBalance(t, "0xAlice", "0xToken", ""); got != "200" {
			t.Errorf("erc20 balance: want 200, got %s", got)
		}
		if got := readBalance(t, "0xAlice", "0xToken", "0x01"); got != "1" {
			t.Errorf("erc721 balance: want 1, got %s", got)
		}
	})
}

func TestDelete_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	queryRepo := NewQueryRepository(testDB)
	indexerRepo := NewIndexerRepository(testDB)

	t.Run("delete existing block", func(t *testing.T) {
		err := indexerRepo.Delete(context.Background(), 100)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}

		// Verify block is gone
		_, err = queryRepo.GetBlockById(context.Background(), 100)
		if err == nil {
			t.Fatal("block should be deleted")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("should have ErrNotFound, has %v", err)
		}
	})

	t.Run("delete non-existing block", func(t *testing.T) {
		err := indexerRepo.Delete(context.Background(), 999)
		if err != nil {
			t.Fatalf("shouldn't have error for non-existing block: %v", err)
		}
	})
}
