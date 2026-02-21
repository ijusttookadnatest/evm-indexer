package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var entities = map[string]*Entity{
    "block":       newEntity(),
    "transaction": newEntity(),
    "event":       newEntity(),
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
       return true
    },
}

func handler(w http.ResponseWriter, r *http.Request) {
   conn, err := upgrader.Upgrade(w, r, nil)
   if err != nil {
      fmt.Println("Error upgrading:", err)
      return
   }
   defer conn.Close()

   client := newClient(conn)
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
   mux := http.NewServeMux()
   mux.HandleFunc("/ws", handler)

   go entities["block"].broadcaster()
   go entities["transaction"].broadcaster()
   go entities["event"].broadcaster()

   return mux
}