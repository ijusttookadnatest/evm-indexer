package ws

import (
	"context"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

type Handler struct {
	entities map[string]*Entity
	ctx context.Context
}

func NewHandler(ctx context.Context, entities map[string]*Entity) *Handler {
	return &Handler{ctx:ctx, entities: entities}
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
	go client.messageWriter(handler.ctx)

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

func NewRouter(ctx context.Context, indexerStreams domain.IndexerStreams) http.Handler {
	entities := map[string]*Entity{
		"blocks":       newEntity("block", indexerStreams.Block),
		"transactions": newEntity("transaction", indexerStreams.Txs),
		"events":       newEntity("event", indexerStreams.Events),
	}
	handler := NewHandler(ctx, entities)

	go entities["blocks"].broadcast(ctx)
	go entities["transactions"].broadcast(ctx)
	go entities["events"].broadcast(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.entitySubscription)

	return mux
}
