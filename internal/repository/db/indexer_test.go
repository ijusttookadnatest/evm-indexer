//go:build integration

package repository

import (
	"context"
	"errors"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"testing"
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

		err := indexerRepo.Create(block, txs, events)
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
		err := indexerRepo.Create(block, nil, nil)
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

		if err := indexerRepo.BulkCreate(items); err != nil {
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
		if err != nil || len(event.Topics) != 2 {
			t.Fatalf("expected event with 2 topics, got err=%v topics=%v", err, event.Topics)
		}
	})

	t.Run("duplicate blocks are idempotent", func(t *testing.T) {
		truncateAll(t)
		seedFixtures(t)

		items := []domain.BlockTxsEvents{
			{Block: domain.Block{Hash: "0xblock100", Id: 100}, Txs: nil, Events: nil},
			{Block: domain.Block{Hash: "0xblock101", Id: 101}, Txs: nil, Events: nil},
		}

		if err := indexerRepo.BulkCreate(items); err != nil {
			t.Fatalf("duplicate blocks should be silently ignored, got: %v", err)
		}
	})

	t.Run("blocks with no txs or events", func(t *testing.T) {
		truncateAll(t)

		items := []domain.BlockTxsEvents{
			{Block: domain.Block{Hash: "0xEmpty1", Id: 10, ParentHash: "0x0", GasLimit: 1000, GasUsed: 0, Miner: "0xM", Timestamp: 300}},
			{Block: domain.Block{Hash: "0xEmpty2", Id: 11, ParentHash: "0xEmpty1", GasLimit: 1000, GasUsed: 0, Miner: "0xM", Timestamp: 400}},
		}

		if err := indexerRepo.BulkCreate(items); err != nil {
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
		cursor, err := indexerRepo.GetBackfillCursor()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cursor != 0 {
			t.Errorf("want 0, got %d", cursor)
		}
	})

	t.Run("update and read cursor", func(t *testing.T) {
		err := indexerRepo.UpdateBackfillCursor(12345)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cursor, err := indexerRepo.GetBackfillCursor()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cursor != 12345 {
			t.Errorf("want 12345, got %d", cursor)
		}
		// reset
		_ = indexerRepo.UpdateBackfillCursor(0)
	})
}

func TestResetBackfillCursor_Integration(t *testing.T) {
	indexerRepo := NewIndexerRepository(testDB)

	_ = indexerRepo.UpdateBackfillCursor(12345)

	err := indexerRepo.ResetBackfillCursor()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cursor, err := indexerRepo.GetBackfillCursor()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cursor != 0 {
		t.Errorf("want 0 after reset, got %d", cursor)
	}
}

func TestDelete_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	queryRepo := NewQueryRepository(testDB)
	indexerRepo := NewIndexerRepository(testDB)

	t.Run("delete existing block", func(t *testing.T) {
		err := indexerRepo.Delete(100)
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
		err := indexerRepo.Delete(999)
		if err != nil {
			t.Fatalf("shouldn't have error for non-existing block: %v", err)
		}
	})
}
