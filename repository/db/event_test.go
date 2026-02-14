//go:build integration

package repository

import (
	"errors"
	"github/ijusttookadnatest/indexer-evm/domain"
	"testing"
)

func TestGetByTxHashLogIndex_Integration(t *testing.T) {
	truncateAll(t)
	seedFixtures(t)

	repo := NewEventRepository(testDB)

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

	repo := NewEventRepository(testDB)

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
		filter := domain.EventFilter{Topics: &topics}
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
		filter := domain.EventFilter{Emitter: &emitter, Topics: &topics}
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
