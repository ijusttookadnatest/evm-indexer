package ws

import (
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

type Handler struct {
	entities map[string]*Entity
}

func NewHandler(entities map[string]*Entity) *Handler {
	return &Handler{entities: entities}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (handler *Handler) entitySubscription(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error upgrading:", "reason", err)
		return
	}
	defer conn.Close()

	client := newClient(conn, handler.entities)
	go client.messageWriter()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			client.delete()
			break
		}
		if err = client.subscribe(message); err != nil {
			b, _ := marshalWSMessage("error", map[string]string{"message": err.Error()})
			if err = client.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				slog.Error("ws: failed to write error response", "err", err)
				client.delete()
				break
			}
		}
	}
}

func NewRouter(indexerStreams domain.IndexerStreams) http.Handler {
	entities := map[string]*Entity{
		"blocks":       newEntity("block", indexerStreams.Block),
		"transactions": newEntity("transaction", indexerStreams.Txs),
		"events":       newEntity("event", indexerStreams.Events),
	}
	handler := NewHandler(entities)

	go entities["blocks"].broadcast()
	go entities["transactions"].broadcast()
	go entities["events"].broadcast()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler.entitySubscription)

	return mux
}
