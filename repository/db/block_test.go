//go:build integration

package repository

import (
	"errors"
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"testing"
)

func TestGetByNumber_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewBlockRepository(testDB)

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

	repo := NewBlockRepository(testDB)

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

	repo := NewBlockRepository(testDB)

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

	repo := NewBlockRepository(testDB)

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

func TestCreate_Integration(t *testing.T) {
	repo := NewBlockRepository(testDB)

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

		err := repo.Create(block, txs, events)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}

		// Verify block was persisted
		got, err := repo.GetById(200)
		if err != nil {
			t.Fatalf("block should exist: %v", err)
		}
		if got.Hash != "0xnew" || got.Miner != "0xminer" {
			t.Errorf("invalid block data: %v", got)
		}
	})

	t.Run("create duplicate block fails", func(t *testing.T) {
		truncateAll(t)
		seedFixtures(t)

		block := domain.Block{
			Hash: "0xdupe", Id: 100, ParentHash: "0xparent",
			GasLimit: 30000000, GasUsed: 10000000, Miner: "0xminer", Timestamp: 1800000000,
		}
		err := repo.Create(block, nil, nil)
		if err == nil {
			t.Fatal("should return an error for duplicate block_id")
		}
	})
}

func TestDelete_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewBlockRepository(testDB)

	t.Run("delete existing block", func(t *testing.T) {
		err := repo.Delete(100)
		if err != nil {
			t.Fatalf("shouldn't have error: %v", err)
		}

		// Verify block is gone
		_, err = repo.GetById(100)
		if err == nil {
			t.Fatal("block should be deleted")
		}
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("should have ErrNotFound, has %v", err)
		}
	})

	t.Run("delete non-existing block", func(t *testing.T) {
		err := repo.Delete(999)
		if err != nil {
			t.Fatalf("shouldn't have error for non-existing block: %v", err)
		}
	})
}
