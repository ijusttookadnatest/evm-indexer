package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
)

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
			incoming := make(chan any, 1)
			entity := newEntity(incoming)

			clientChan := make(chan []byte, 1)
			entity.mu.Lock()
			entity.clientsChan[tc.filter] = []chan[]byte{clientChan}
			entity.mu.Unlock()

			go entity.broadcast()

			incoming <- tc.payload

			select {
			case got := <-clientChan:
				if !tc.wantRecv {
					t.Errorf("did not expect to receive, got: %s", got)
					return
				}
				expected, _ := json.Marshal(tc.payload)
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
