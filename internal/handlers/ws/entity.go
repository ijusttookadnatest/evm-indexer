package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	custmetrics "github/ijusttookadnatest/evm-indexer/internal/metrics"
)

type PayloadFilter struct {
	From    string   `json:"From"`
	To      string   `json:"To"`
	Emitter string   `json:"Emitter"`
	Topic   []string `json:"Topics"`
}

type SubscriptionFilter struct {
	Address string
	Topic0  string
}

type Entity struct {
	clientsChan map[SubscriptionFilter][]chan []byte
	mu          *sync.RWMutex
	incoming    <-chan []byte
	name        string
	metrics     *custmetrics.ApiMetrics
}

func newEntity(name string, c <-chan []byte, metrics *custmetrics.ApiMetrics) *Entity {
	return &Entity{
		clientsChan: make(map[SubscriptionFilter][]chan []byte),
		mu:          &sync.RWMutex{},
		incoming:    c,
		name:        name,
		metrics:     metrics,
	}
}

func (entity Entity) broadcast(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-entity.incoming:
			var filteredPayload = new(PayloadFilter)

			err := json.Unmarshal(payload, &filteredPayload)
			if err != nil {
				slog.Error("broadcast: unmarshal failed", "err", err)
				continue
			}
			bytes, _ := marshalWSMessage(entity.name, json.RawMessage(payload))
			entity.mu.RLock()
			for filter, clientsChan := range entity.clientsChan {
				if matchesFilter(filter, *filteredPayload) {
					for _, clientChan := range clientsChan {
						select {
						case clientChan <- bytes:
							entity.metrics.WsTotalMessageSent.Inc()
						default:
						}
					}
				}
			}
			entity.mu.RUnlock()
		}
	}
}
