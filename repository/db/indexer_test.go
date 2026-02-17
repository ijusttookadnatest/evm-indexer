//go:build integration

package repository

import (
	"errors"
	"github/ijusttookadnatest/indexer-evm/core/domain"
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
		got, err := queryRepo.GetById(200)
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
		err := indexerRepo.Create(block, nil, nil)
		if err == nil {
			t.Fatal("should return an error for duplicate block_id")
		}
	})
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
		_, err = queryRepo.GetById(100)
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
