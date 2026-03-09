package ws

import (
	"fmt"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
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

func NewRouter(indexerStreams domain.IndexerStreams) http.Handler {
   entities := map[string]*Entity{
       "blocks":       newEntity(indexerStreams.Block),
       "transactions": newEntity(indexerStreams.Txs),
       "events":       newEntity(indexerStreams.Events),
   }
   handler := NewHandler(entities)

   mux := http.NewServeMux()
   mux.HandleFunc("/ws", handler.entitySubscription)

   go entities["blocks"].broadcast()
   go entities["transactions"].broadcast()
   go entities["events"].broadcast()

   return mux
}