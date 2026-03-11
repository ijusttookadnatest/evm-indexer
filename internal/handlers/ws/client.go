package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type SubscribeMessage struct {
	Type 	string `json:"type"`
	Topic 	string `json:"topic"`
	Address string `json:"address"`
	Topic0 	string `json:"topics0"`
}

type Position struct {
	topic 	string
	filter 	SubscriptionFilter
	index 	int
}

type Client struct {
   conn  		*websocket.Conn
   outgoing     chan []byte
   entities 	map[string]*Entity
   pos	    	[]Position
   once         sync.Once
}

func newClient(conn *websocket.Conn, entities map[string]*Entity) *Client {
   c := make(chan []byte)
   var pos []Position

   return &Client{
      conn: conn,
      outgoing: c,
	  entities: entities,
	  pos: pos,
   }
}

func (client *Client) subscribe(message []byte) error {
  	subscription := new(SubscribeMessage)
	if err := json.Unmarshal(message, subscription) ; err != nil {
		return err
	}
	if err := validateSubscription(*subscription) ; err != nil {
		return err
	}
	
	entity := client.entities[subscription.Topic]
	filter := extractFilter(*subscription)
	
	entity.mu.Lock()
	index := len(entity.clientsChan[filter])
	entity.clientsChan[filter] = append(entity.clientsChan[filter], client.outgoing)
	entity.mu.Unlock()

	client.pos = append(client.pos, Position{
		topic: subscription.Topic,
		filter: filter,
		index:index,
	})

	return nil
}

func (client *Client) delete() {
	client.once.Do(func() {
		for _, pos := range client.pos {
			entity := client.entities[pos.topic]
			
			entity.mu.Lock()
			c := entity.clientsChan[pos.filter]
			entity.clientsChan[pos.filter] = append(c[:pos.index], c[pos.index+1:]...) 
			entity.mu.Unlock()
		}
		client.conn.Close()
		close(client.outgoing)
	})
}

func (client *Client) messageWriter() {
	for {
		message, ok := <- client.outgoing
		if !ok {
			return
		}
		err := client.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.delete()
			return
		}
	}
}