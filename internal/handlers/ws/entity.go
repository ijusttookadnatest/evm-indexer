package ws

import (
	"encoding/json"
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
   mu *sync.RWMutex
   incoming chan []byte
}

func newEntity() *Entity {
	clientsChan := make(map[SubscriptionFilter][]chan[]byte)
	c := make(chan []byte)
	return &Entity{
		clientsChan:clientsChan,
		mu: &sync.RWMutex{},
		incoming: c,
	}
}

func (entity Entity) broadcast() {
   for {
		data := <- entity.incoming
		var payload = new(PayloadFilter)
		
		json.Unmarshal(data, &payload)
		entity.mu.RLock()
		for filter, clientsChan := range entity.clientsChan {
			if matchesFilter(filter, *payload) {
				for _, clientChan := range clientsChan {
					select {
					case clientChan <- data:
					default:
					}
				}
			}
		}
		entity.mu.RUnlock()
   }
}