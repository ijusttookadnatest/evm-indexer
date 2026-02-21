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
   clients map[SubscriptionFilter][]Client
   mu *sync.RWMutex
   incoming chan []byte
}

func newEntity() *Entity {
	clients := make(map[SubscriptionFilter][]Client)
	c := make(chan []byte)
	return &Entity{
		clients:clients,
		mu: &sync.RWMutex{},
		incoming: c,
	}
}

func (entity Entity) broadcaster() {
   for {
		data := <- entity.incoming
		var payload = new(PayloadFilter)
		
		json.Unmarshal(data, &payload)
		entity.mu.RLock()
		for filter, clients := range entity.clients {
			if matchesFilter(filter, *payload) {
				for _, client := range clients {
					select {
					case client.outgoing <- data:
					default:
					}
				}
			}
		}
		entity.mu.RUnlock()
   }
}