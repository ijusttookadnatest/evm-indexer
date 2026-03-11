package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
)

type PayloadFilter struct {
	From string `json:"From"`
	To string `json:"To"`
	Emitter string `json:"Emitter"`
	Topic []string `json:"Topics"`
}

type SubscriptionFilter struct {
	Address string
	Topic0 string
}

type Entity struct {
   clientsChan map[SubscriptionFilter][]chan[]byte
   mu          *sync.RWMutex
   incoming    chan any
   name        string
}

func newEntity(name string, c chan any) *Entity {
	return &Entity{
		clientsChan: make(map[SubscriptionFilter][]chan[]byte),
		mu:          &sync.RWMutex{},
		incoming:    c,
		name:        name,
	}
}

func (entity Entity) broadcast() {
   for {
		data := <- entity.incoming
		payload, err := json.Marshal(data)
		if err != nil {
			slog.Error("broadcast: marshal failed", "err", err)
      		continue
		}
		var filteredPayload = new(PayloadFilter)
		
		err = json.Unmarshal(payload, &filteredPayload)
		if err != nil {
			slog.Error("broadcast: unmarshal failed", "err", err)
      		continue
		}
		bytes, _ := marshalWSMessage(entity.name, data)
		entity.mu.RLock()
		for filter, clientsChan := range entity.clientsChan {
			if matchesFilter(filter, *filteredPayload) {
				for _, clientChan := range clientsChan {
					select {
					case clientChan <- bytes:
					default:
					}
				}
			}
		}
		entity.mu.RUnlock()
   }
}