//go:build integration

package repository

import (
	"errors"
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"testing"
)

// --- Block queries ---

func TestGetByNumber_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewQueryRepository(testDB)

	t.Run("valid block", func(t *testing.T) {
		block, err := repo.GetById(100)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if block == nil {
			t.Fatal("should return a block")
		}
		if block.Id != 100 || block.Hash != "0xblock100" {
			t.Errorf("invalid block data: %v", block)
		}
	})

	t.Run("no block", func(t *testing.T) {
		_, err := repo.GetById(999)
		if err == nil {
			t.Fatal("should return an error")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("should have error of type %v, has %v", domain.ErrNotFound, err)
		}
	})
}

func TestGetByHash_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewQueryRepository(testDB)

	t.Run("valid block", func(t *testing.T) {
		block, err := repo.GetByHash("0xblock100")
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if block == nil {
			t.Fatal("should return a block")
		}
		if block.Id != 100 || block.Hash != "0xblock100" {
			t.Errorf("invalid block data: %v", block)
		}
	})

	t.Run("no block", func(t *testing.T) {
		_, err := repo.GetByHash("0xnonexistent")
		if err == nil {
			t.Fatal("should return an error")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("should have error of type %v, has %v", domain.ErrNotFound, err)
		}
	})
}

func TestGetByRangeNumber_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewQueryRepository(testDB)

	t.Run("valid range with results", func(t *testing.T) {
		blocks, err := repo.GetByRangeId(99, 103)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(blocks) != 3 {
			t.Fatalf("should return 3 blocks, has %v", len(blocks))
		}
		if blocks[0].Id != 100 || blocks[1].Id != 101 || blocks[2].Id != 102 {
			t.Errorf("invalid blocks data: %v", blocks)
		}
	})

	t.Run("range with one result", func(t *testing.T) {
		blocks, err := repo.GetByRangeId(99, 101)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(blocks) != 1 {
			t.Fatalf("should return 1 block, has %v", len(blocks))
		}
		if blocks[0].Id != 100 {
			t.Errorf("invalid block data: %v", blocks[0])
		}
	})

	t.Run("empty range", func(t *testing.T) {
		_, err := repo.GetByRangeId(500, 600)
		if err == nil {
			t.Fatal("should return an error")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("should have error of type %v, has %v", domain.ErrNotFound, err)
		}
	})
}

func TestGetByRangeTime_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewQueryRepository(testDB)

	t.Run("valid time range", func(t *testing.T) {
		blocks, err := repo.GetByRangeTime(1699999999, 1700000025)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(blocks) != 3 {
			t.Fatalf("should return 3 blocks, has %v", len(blocks))
		}
	})

	t.Run("partial time range", func(t *testing.T) {
		blocks, err := repo.GetByRangeTime(1700000010, 1700000020)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(blocks) != 1 {
			t.Fatalf("should return 1 block, has %v", len(blocks))
		}
		if blocks[0].Id != 101 {
			t.Errorf("invalid block data: %v", blocks[0])
		}
	})

	t.Run("empty time range", func(t *testing.T) {
		_, err := repo.GetByRangeTime(9000000000, 9000000001)
		if err == nil {
			t.Fatal("should return an error")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("should have error of type %v, has %v", domain.ErrNotFound, err)
		}
	})
}

// --- Transaction queries ---

func TestGetByTransactionFilter_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewQueryRepository(testDB)

	t.Run("all nil filters returns all", func(t *testing.T) {
		filter := domain.TransactionFilter{}
		txs, err := repo.GetByTransactionFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(txs) != 3 {
			t.Fatalf("should return 3 txs, has %v", len(txs))
		}
	})

	t.Run("filter by hash", func(t *testing.T) {
		hash := "0xtx1"
		filter := domain.TransactionFilter{Hash: &hash}
		txs, err := repo.GetByTransactionFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(txs) != 1 {
			t.Fatalf("should return 1 tx, has %v", len(txs))
		}
		if txs[0].Hash != "0xtx1" {
			t.Errorf("invalid tx hash: %v", txs[0].Hash)
		}
	})

	t.Run("filter by from", func(t *testing.T) {
		from := "0xAlice"
		filter := domain.TransactionFilter{From: &from}
		txs, err := repo.GetByTransactionFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(txs) != 2 {
			t.Fatalf("should return 2 txs from Alice, has %v", len(txs))
		}
	})

	t.Run("filter by to", func(t *testing.T) {
		to := "0xBob"
		filter := domain.TransactionFilter{To: &to}
		txs, err := repo.GetByTransactionFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(txs) != 1 {
			t.Fatalf("should return 1 tx to Bob, has %v", len(txs))
		}
		if txs[0].Hash != "0xtx1" {
			t.Errorf("invalid tx: %v", txs[0])
		}
	})

	t.Run("filter by from and to", func(t *testing.T) {
		from := "0xAlice"
		to := "0xContract"
		filter := domain.TransactionFilter{From: &from, To: &to}
		txs, err := repo.GetByTransactionFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(txs) != 1 {
			t.Fatalf("should return 1 tx, has %v", len(txs))
		}
		if txs[0].Hash != "0xtx2" {
			t.Errorf("invalid tx: %v", txs[0])
		}
	})

	t.Run("no match", func(t *testing.T) {
		from := "0xNobody"
		filter := domain.TransactionFilter{From: &from}
		txs, err := repo.GetByTransactionFilter(filter)
		if len(txs) != 0 {
			t.Errorf("should return empty slice, has %v", len(txs))
		}
		if err != nil {
			if !errors.Is(domain.ErrNotFound, err) {
				t.Errorf("should have no row error, has: %v", err)
			}
		}
	})
}

// --- Event queries ---

func TestGetByTxHashLogIndex_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewQueryRepository(testDB)

	t.Run("valid event", func(t *testing.T) {
		events, err := repo.GetByTxHashLogIndex("0xtx2", 0)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("should return 1 event, has %v", len(events))
		}
		if events[0].LogIndex != 0 || events[0].TxHash != "0xtx2" {
			t.Errorf("invalid events data: %v", events)
		}
	})

	t.Run("multiple events same tx", func(t *testing.T) {
		events, err := repo.GetByTxHashLogIndex("0xtx2", 1)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("should return 1 event, has %v", len(events))
		}
		if events[0].Emitter != "0xContract" {
			t.Errorf("invalid emitter: %v", events[0].Emitter)
		}
	})

	t.Run("no event", func(t *testing.T) {
		_, err := repo.GetByTxHashLogIndex("0xnonexistent", 0)
		if err == nil {
			t.Fatal("should return an error")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("should have error of type %v, has %v", domain.ErrNotFound, err)
		}
	})
}

func TestGetByEventFilter_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewQueryRepository(testDB)

	t.Run("filter by tx_hash", func(t *testing.T) {
		txHash := "0xtx2"
		filter := domain.EventFilter{TxHash: &txHash}
		events, err := repo.GetByEventFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("should return 2 events, has %v", len(events))
		}
	})

	t.Run("filter by emitter", func(t *testing.T) {
		emitter := "0xContract"
		filter := domain.EventFilter{Emitter: &emitter}
		events, err := repo.GetByEventFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("should return 2 events, has %v", len(events))
		}
	})

	t.Run("filter by topics", func(t *testing.T) {
		topics := []string{"0xTransferSig"}
		filter := domain.EventFilter{Topics: topics}
		events, err := repo.GetByEventFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("should return 1 event, has %v", len(events))
		}
		if events[0].TxHash != "0xtx2" || events[0].LogIndex != 0 {
			t.Errorf("invalid event data: %v", events[0])
		}
	})

	t.Run("filter by emitter and topics", func(t *testing.T) {
		emitter := "0xContract"
		topics := []string{"0xApprovalSig"}
		filter := domain.EventFilter{Emitter: &emitter, Topics: topics}
		events, err := repo.GetByEventFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("should return 1 event, has %v", len(events))
		}
		if events[0].LogIndex != 1 {
			t.Errorf("should be log_index 1, got %v", events[0].LogIndex)
		}
	})

	t.Run("no match", func(t *testing.T) {
		emitter := "0xNobody"
		filter := domain.EventFilter{Emitter: &emitter}
		events, err := repo.GetByEventFilter(filter)
		if len(events) != 0 {
			t.Errorf("should return empty slice, has %v", len(events))
		}
		if err != nil {
			if !errors.Is(domain.ErrNotFound, err) {
				t.Errorf("should have no row error, has: %v", err)
			}
		}
	})

	t.Run("all nil filters", func(t *testing.T) {
		filter := domain.EventFilter{}
		events, err := repo.GetByEventFilter(filter)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("should return all 2 events, has %v", len(events))
		}
	})
}
