package ws

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
	"github.com/gorilla/websocket"
)

// ── TestEntitySubscriptionReturnsErrorOnInvalidMessage ────────────────────────
//
// The server must respond with a {"type":"error","payload":{"message":"..."}}
// frame when the client sends an invalid subscription message.
func TestEntitySubscriptionReturnsErrorOnInvalidMessage(t *testing.T) {
	streams := domain.IndexerStreams{
		Block:  make(chan any),
		Txs:    make(chan any),
		Events: make(chan any),
	}
	srv := httptest.NewServer(NewRouter(streams))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("not json")); err != nil {
		t.Fatalf("write: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected error frame from server, got read error: %v", err)
	}

	var response WSMessage
	if err := json.Unmarshal(msg, &response); err != nil {
		t.Fatalf("response is not valid JSON: %s", msg)
	}
	if response.Type != "error" {
		t.Errorf("expected type %q, got %q", "error", response.Type)
	}
}
