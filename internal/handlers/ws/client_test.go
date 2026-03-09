package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func testEntities() map[string]*Entity {
	return map[string]*Entity{
		"blocks":       newEntity(make(chan any)),
		"transactions": newEntity(make(chan any)),
		"events":       newEntity(make(chan any)),
	}
}

func subMsg(msgType, topic, address, topic0 string) []byte {
	b, _ := json.Marshal(SubscribeMessage{
		Type:    msgType,
		Topic:   topic,
		Address: address,
		Topic0:  topic0,
	})
	return b
}

func newTestWSConn(t *testing.T) (*websocket.Conn, func()) {
	t.Helper()
	u := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := u.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("ws dial: %v", err)
	}
	return conn, srv.Close
}

// ── TestSubscribe ─────────────────────────────────────────────────────────────

func TestSubscribe(t *testing.T) {
	tests := []struct {
		name        string
		message     []byte
		wantErr     bool
		wantPosLen  int
		wantChanLen int
	}{
		{
			name:        "valid — blocks no filter",
			message:     subMsg("subscribe", "blocks", "", ""),
			wantErr:     false,
			wantPosLen:  1,
			wantChanLen: 1,
		},
		{
			name:        "valid — events with address",
			message:     subMsg("subscribe", "events", "0xabc", ""),
			wantErr:     false,
			wantPosLen:  1,
			wantChanLen: 1,
		},
		{
			name:    "invalid JSON",
			message: []byte("not json"),
			wantErr: true,
		},
		{
			name:    "wrong type",
			message: subMsg("unsubscribe", "blocks", "", ""),
			wantErr: true,
		},
		{
			name:    "unknown topic",
			message: subMsg("subscribe", "unknown", "", ""),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entities := testEntities()
			client := newClient(nil, entities) // conn unused in subscribe

			err := client.subscribe(tc.message)

			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
				return
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tc.wantErr {
				return
			}

			if len(client.pos) != tc.wantPosLen {
				t.Errorf("pos len: got %d, want %d", len(client.pos), tc.wantPosLen)
			}

			sub := new(SubscribeMessage)
			json.Unmarshal(tc.message, sub)
			filter := extractFilter(*sub)
			entity := entities[sub.Topic]

			entity.mu.RLock()
			chanLen := len(entity.clientsChan[filter])
			entity.mu.RUnlock()

			if chanLen != tc.wantChanLen {
				t.Errorf("clientsChan len: got %d, want %d", chanLen, tc.wantChanLen)
			}
		})
	}
}

// ── TestDelete ────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	t.Run("removes client from entity after subscribe", func(t *testing.T) {
		conn, cleanup := newTestWSConn(t)
		defer cleanup()

		entities := testEntities()
		client := newClient(conn, entities)

		if err := client.subscribe(subMsg("subscribe", "blocks", "", "")); err != nil {
			t.Fatalf("subscribe: %v", err)
		}

		filter := SubscriptionFilter{}
		entities["blocks"].mu.RLock()
		before := len(entities["blocks"].clientsChan[filter])
		entities["blocks"].mu.RUnlock()

		if before != 1 {
			t.Fatalf("expected 1 client before delete, got %d", before)
		}

		client.delete()

		entities["blocks"].mu.RLock()
		after := len(entities["blocks"].clientsChan[filter])
		entities["blocks"].mu.RUnlock()

		if after != 0 {
			t.Errorf("expected 0 clients after delete, got %d", after)
		}
	})

	t.Run("removes client from multiple subscriptions", func(t *testing.T) {
		conn, cleanup := newTestWSConn(t)
		defer cleanup()

		entities := testEntities()
		client := newClient(conn, entities)

		client.subscribe(subMsg("subscribe", "blocks", "", ""))
		client.subscribe(subMsg("subscribe", "events", "0xabc", ""))

		client.delete()

		filter := SubscriptionFilter{}
		entities["blocks"].mu.RLock()
		blockLen := len(entities["blocks"].clientsChan[filter])
		entities["blocks"].mu.RUnlock()

		eventFilter := SubscriptionFilter{Address: "0xabc"}
		entities["events"].mu.RLock()
		eventLen := len(entities["events"].clientsChan[eventFilter])
		entities["events"].mu.RUnlock()

		if blockLen != 0 || eventLen != 0 {
			t.Errorf("subscriptions not fully removed: blocks=%d events=%d", blockLen, eventLen)
		}
	})
}
