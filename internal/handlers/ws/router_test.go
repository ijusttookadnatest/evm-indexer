package ws

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type mockRedisPubSub struct{}

func (m *mockRedisPubSub) Publish(_ context.Context, _ string, _ []byte) error { return nil }
func (m *mockRedisPubSub) Subscribe(_ context.Context, _ string) (<-chan []byte, error) {
	return make(chan []byte), nil
}

// ── TestEntitySubscriptionReturnsErrorOnInvalidMessage ────────────────────────
//
// The server must respond with a {"type":"error","payload":{"message":"..."}}
// frame when the client sends an invalid subscription message.
func TestEntitySubscriptionReturnsErrorOnInvalidMessage(t *testing.T) {
	handler, err := NewRouter(context.Background(), &mockRedisPubSub{}, newTestApiMetrics())
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
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

// ── TestEntitySubscriptionClosedOnContextCancel ───────────────────────────────
//
// When the server context is cancelled, the messageWriter calls client.delete()
// which closes the connection, causing ReadMessage to return an error and
// unblocking the entitySubscription loop. The client should observe a close frame.
func TestEntitySubscriptionClosedOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	handler, err := NewRouter(ctx, &mockRedisPubSub{}, newTestApiMetrics())
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	cancel()

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Error("expected connection to be closed after context cancel, but ReadMessage succeeded")
	}
}
