//go:build integration

package repository

import (
	"errors"
	"github/ijusttookadnatest/indexer-evm/domain"
	"testing"
)

func TestGetByTransactionFilter_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewTransactionRepository(testDB)

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
