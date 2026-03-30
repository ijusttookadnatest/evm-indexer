package ws

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
)

// ── TestBroadcastMarshalError ──────────────────────────────────────────────────
//
// BUG (unhandled): when json.Marshal fails, execution continues with nil bytes.
// A subscriber with an empty filter will receive nil — this test documents that.
// Correct behavior: skip the iteration and log/return the error.
func TestBroadcastMarshalErrorSilentlyIgnored(t *testing.T) {
	incoming := make(chan []byte, 1)
	entity := newEntity("block", incoming, newTestApiMetrics())

	clientChan := make(chan []byte, 1)
	entity.mu.Lock()
	entity.clientsChan[SubscriptionFilter{}] = []chan []byte{clientChan}
	entity.mu.Unlock()

	go entity.broadcast(context.Background())

	incoming <- []byte("not json") // json.Unmarshal fails on invalid JSON

	select {
	case got := <-clientChan:
		t.Errorf("BUG: received %v bytes despite marshal error — should have received nothing", got)
	case <-time.After(100 * time.Millisecond):
		// correct: nothing broadcast
	}
}

// ── TestBroadcastSlowClientDropsMessage ───────────────────────────────────────
//
// HANDLED: broadcast uses select/default so a slow/blocked client does not
// block the fan-out loop. The message is simply dropped for that client.
func TestBroadcastSlowClientDropsMessage(t *testing.T) {
	incoming := make(chan []byte, 1)
	entity := newEntity("block", incoming, newTestApiMetrics())

	// Simulate a slow client: buffered channel that is already full.
	// broadcast's select/default will drop the new message instead of blocking.
	slowChan := make(chan []byte, 1)
	slowChan <- []byte("already full")
	entity.mu.Lock()
	entity.clientsChan[SubscriptionFilter{}] = []chan []byte{slowChan}
	entity.mu.Unlock()

	go entity.broadcast(context.Background())

	b, _ := json.Marshal(domain.Block{Id: 1, Hash: "0xabc"})
	incoming <- b
	time.Sleep(100 * time.Millisecond)

	// Drain: should only contain the original "already full" sentinel, not the broadcast.
	if len(slowChan) != 1 {
		t.Errorf("expected 1 message in channel (sentinel), got %d — broadcast was not dropped", len(slowChan))
	}
	if msg := <-slowChan; string(msg) != "already full" {
		t.Errorf("expected sentinel message, got: %s", msg)
	}
}

func TestBroadcast(t *testing.T) {
	tests := []struct {
		name     string
		filter   SubscriptionFilter
		payload  any
		wantRecv bool
	}{
		{
			name:     "block no filter — received",
			filter:   SubscriptionFilter{},
			payload:  domain.Block{Id: 1, Hash: "0xabc"},
			wantRecv: true,
		},
		{
			name:     "transaction matches address filter on From",
			filter:   SubscriptionFilter{Address: "0xsender"},
			payload:  domain.Transaction{From: "0xsender", Hash: "0xtx1"},
			wantRecv: true,
		},
		{
			name:     "transaction no match on address",
			filter:   SubscriptionFilter{Address: "0xother"},
			payload:  domain.Transaction{From: "0xsender", Hash: "0xtx1"},
			wantRecv: false,
		},
		{
			name:     "event matches emitter filter",
			filter:   SubscriptionFilter{Address: "0xcontract"},
			payload:  domain.Event{Emitter: "0xcontract", Topics: []string{"0xtopic0"}},
			wantRecv: true,
		},
		{
			name:     "event matches topic0 filter",
			filter:   SubscriptionFilter{Topic0: "0xtopic0"},
			payload:  domain.Event{Emitter: "0xcontract", Topics: []string{"0xtopic0"}},
			wantRecv: true,
		},
		{
			name:     "event no match on topic0",
			filter:   SubscriptionFilter{Topic0: "0xother"},
			payload:  domain.Event{Emitter: "0xcontract", Topics: []string{"0xtopic0"}},
			wantRecv: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			incoming := make(chan []byte, 1)
			entity := newEntity("block", incoming, newTestApiMetrics())

			clientChan := make(chan []byte, 1)
			entity.mu.Lock()
			entity.clientsChan[tc.filter] = []chan []byte{clientChan}
			entity.mu.Unlock()

			go entity.broadcast(context.Background())

			b, _ := json.Marshal(tc.payload)
			incoming <- b

			select {
			case got := <-clientChan:
				if !tc.wantRecv {
					t.Errorf("did not expect to receive, got: %s", got)
					return
				}
				expected, _ := marshalWSMessage("block", json.RawMessage(b))
				if string(got) != string(expected) {
					t.Errorf("payload mismatch:\n got  %s\n want %s", got, expected)
				}
			case <-time.After(100 * time.Millisecond):
				if tc.wantRecv {
					t.Error("expected to receive, got nothing")
				}
			}
		})
	}
}
