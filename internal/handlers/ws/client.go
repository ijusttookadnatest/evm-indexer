package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Client struct {
   conn  *websocket.Conn
   outgoing     chan []byte
}

func newClient(conn *websocket.Conn) *Client {
   c := make(chan []byte)

   return &Client{
      conn: conn,
      outgoing: c,
   }
}


func (client Client) subscribe(message []byte) error {
  	subscription := new(subscribeMessage)
	if err := json.Unmarshal(message, subscription) ; err != nil {
		return err
	}
	if err := validateSubscription(*subscription) ; err != nil {
		return err
	}
	
	entity := entities[subscription.Topic]
	filter := extractFilter(*subscription)

	entity.mu.Lock()
	entity.clients[filter] = append(entity.clients[filter], client)
	entity.mu.Unlock()

	return nil
}

func (client Client) delete() {
   client.conn.Close()
   for _, entity := range entities {
      entity.mu.Lock()
      for i := range entity.clients {
         for j := range entity.clients[i] {
            if entity.clients[i][j] == client {
               entity.clients[i] = append(entity.clients[i][:j], entity.clients[i][j+1:]...)
               break
            }
         }
      }
      entity.mu.Unlock()
   }
}

func (client Client) messageWriter() {
	for {
		message := <- client.outgoing
		err := client.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.delete()
			return
		}
	}
}