package ws

import (
	"context"
	"log/slog"
	"net/http"

	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	custprometheus "github/ijusttookadnatest/evm-indexer/internal/prometheus"

	"github.com/gorilla/websocket"
)

type Handler struct {
	entities map[string]*Entity
	ctx      context.Context
	metrics  *custprometheus.ApiMetrics
}

func NewHandler(ctx context.Context, entities map[string]*Entity, metrics *custprometheus.ApiMetrics) *Handler {
	return &Handler{ctx: ctx, entities: entities, metrics: metrics}
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
	handler.metrics.WsActiveConnection.Inc()
	defer handler.metrics.WsActiveConnection.Dec()

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

func NewRouter(ctx context.Context, pubsub ports.RedisPubSub, metrics *custprometheus.ApiMetrics) (http.Handler, error) {
	blockIncoming, err := pubsub.Subscribe(ctx, "block")
	if err != nil {
		return nil, err
	}
	txIncoming, err := pubsub.Subscribe(ctx, "transaction")
	if err != nil {
		return nil, err
	}
	eventIncoming, err := pubsub.Subscribe(ctx, "event")
	if err != nil {
		return nil, err
	}
	reorgIncoming, err := pubsub.Subscribe(ctx, "reorg")
	if err != nil {
		return nil, err
	}

	entities := map[string]*Entity{
		"blocks":       newEntity("block", blockIncoming, metrics),
		"transactions": newEntity("transaction", txIncoming, metrics),
		"events":       newEntity("event", eventIncoming, metrics),
		"reorgs":       newEntity("reorg", reorgIncoming, metrics),
	}
	handler := NewHandler(ctx, entities, metrics)

	go entities["blocks"].broadcast(ctx)
	go entities["transactions"].broadcast(ctx)
	go entities["events"].broadcast(ctx)
	go entities["reorgs"].broadcast(ctx)

	return http.HandlerFunc(handler.entitySubscription), nil
}
