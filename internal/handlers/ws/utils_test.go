package ws

import "testing"

// ── validateSubscription ──────────────────────────────────────────────────────

func TestValidateSubscription(t *testing.T) {
	tests := []struct {
		name    string
		sub     SubscribeMessage
		wantErr bool
	}{
		{
			name:    "valid — blocks",
			sub:     SubscribeMessage{Type: "subscribe", Topic: "blocks"},
			wantErr: false,
		},
		{
			name:    "valid — transactions",
			sub:     SubscribeMessage{Type: "subscribe", Topic: "transactions"},
			wantErr: false,
		},
		{
			name:    "valid — events",
			sub:     SubscribeMessage{Type: "subscribe", Topic: "events"},
			wantErr: false,
		},
		{
			name:    "wrong type",
			sub:     SubscribeMessage{Type: "unsubscribe", Topic: "blocks"},
			wantErr: true,
		},
		{
			name:    "empty type",
			sub:     SubscribeMessage{Type: "", Topic: "blocks"},
			wantErr: true,
		},
		{
			name:    "unknown topic",
			sub:     SubscribeMessage{Type: "subscribe", Topic: "unknown"},
			wantErr: true,
		},
		{
			name:    "empty topic",
			sub:     SubscribeMessage{Type: "subscribe", Topic: ""},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSubscription(tc.sub)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ── extractFilter ─────────────────────────────────────────────────────────────

func TestExtractFilter(t *testing.T) {
	tests := []struct {
		name       string
		sub        SubscribeMessage
		wantFilter SubscriptionFilter
	}{
		{
			name:       "no address no topic0",
			sub:        SubscribeMessage{},
			wantFilter: SubscriptionFilter{Address: "", Topic0: ""},
		},
		{
			name:       "address only",
			sub:        SubscribeMessage{Address: "0xabc"},
			wantFilter: SubscriptionFilter{Address: "0xabc", Topic0: ""},
		},
		{
			name:       "topic0 only",
			sub:        SubscribeMessage{Topic0: "0xdef"},
			wantFilter: SubscriptionFilter{Address: "", Topic0: "0xdef"},
		},
		{
			name:       "address and topic0",
			sub:        SubscribeMessage{Address: "0xabc", Topic0: "0xdef"},
			wantFilter: SubscriptionFilter{Address: "0xabc", Topic0: "0xdef"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractFilter(tc.sub)
			if got != tc.wantFilter {
				t.Errorf("filter: got %+v, want %+v", got, tc.wantFilter)
			}
		})
	}
}

// ── matchesFilter ─────────────────────────────────────────────────────────────

func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name         string
		subscription SubscriptionFilter
		payload      PayloadFilter
		wantMatch    bool
	}{
		{
			name:         "empty filter matches everything",
			subscription: SubscriptionFilter{},
			payload:      PayloadFilter{From: "0xaaa", To: "0xbbb"},
			wantMatch:    true,
		},
		{
			name:         "address matches From",
			subscription: SubscriptionFilter{Address: "0xaaa"},
			payload:      PayloadFilter{From: "0xaaa"},
			wantMatch:    true,
		},
		{
			name:         "address matches To",
			subscription: SubscriptionFilter{Address: "0xbbb"},
			payload:      PayloadFilter{To: "0xbbb"},
			wantMatch:    true,
		},
		{
			name:         "address matches Emitter",
			subscription: SubscriptionFilter{Address: "0xccc"},
			payload:      PayloadFilter{Emitter: "0xccc"},
			wantMatch:    true,
		},
		{
			name:         "address does not match",
			subscription: SubscriptionFilter{Address: "0xaaa"},
			payload:      PayloadFilter{From: "0xbbb", To: "0xccc", Emitter: "0xddd"},
			wantMatch:    false,
		},
		{
			name:         "topic0 matches",
			subscription: SubscriptionFilter{Topic0: "0xtopic"},
			payload:      PayloadFilter{Topic: []string{"0xtopic", "0xother"}},
			wantMatch:    true,
		},
		{
			name:         "topic0 does not match",
			subscription: SubscriptionFilter{Topic0: "0xtopic"},
			payload:      PayloadFilter{Topic: []string{"0xwrong"}},
			wantMatch:    false,
		},
		{
			name:         "address and topic0 both match",
			subscription: SubscriptionFilter{Address: "0xaaa", Topic0: "0xtopic"},
			payload:      PayloadFilter{From: "0xaaa", Topic: []string{"0xtopic"}},
			wantMatch:    true,
		},
		{
			name:         "address matches but topic0 does not",
			subscription: SubscriptionFilter{Address: "0xaaa", Topic0: "0xtopic"},
			payload:      PayloadFilter{From: "0xaaa", Topic: []string{"0xwrong"}},
			wantMatch:    false,
		},
		{
			name:         "topic0 filter with empty Topics — should not panic",
			subscription: SubscriptionFilter{Topic0: "0xtopic"},
			payload:      PayloadFilter{Topic: []string{}},
			wantMatch:    false,
		},
		{
			name:         "topic0 filter with nil Topics — should not panic",
			subscription: SubscriptionFilter{Topic0: "0xtopic"},
			payload:      PayloadFilter{},
			wantMatch:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesFilter(tc.subscription, tc.payload)
			if got != tc.wantMatch {
				t.Errorf("matchesFilter: got %v, want %v", got, tc.wantMatch)
			}
		})
	}
}
