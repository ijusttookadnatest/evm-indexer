package ws

import (
	"fmt"
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
      fmt.Println("Error upgrading:", err)
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
      client.subscribe(message)
   }
}

func newRouter() http.Handler {
   entities := map[string]*Entity{
       "block":       newEntity(),
       "transaction": newEntity(),
       "event":       newEntity(),
   }
   handler := NewHandler(entities)

   mux := http.NewServeMux()
   mux.HandleFunc("/ws", handler.entitySubscription)

   go entities["block"].broadcast()
   go entities["transaction"].broadcast()
   go entities["event"].broadcast()

   return mux
}