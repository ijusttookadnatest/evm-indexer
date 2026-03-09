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
   incoming chan any
}

func newEntity(c chan any) *Entity {
	clientsChan := make(map[SubscriptionFilter][]chan[]byte)
	return &Entity{
		clientsChan:clientsChan,
		mu: &sync.RWMutex{},
		incoming: c,
	}
}

func (entity Entity) broadcast() {
   for {
		data := <- entity.incoming
		bytes, err := json.Marshal(data)
		if err != nil {
			// error handling
		}
		var payload = new(PayloadFilter)
		
		err = json.Unmarshal(bytes, &payload)
		if err != nil {
			// error handling
		}
		entity.mu.RLock()
		for filter, clientsChan := range entity.clientsChan {
			if matchesFilter(filter, *payload) {
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