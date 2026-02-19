package server

import (
	"fmt"
	"net/http"
	"time"
)


type Server struct {
	Server *http.Server
}

func NewHTTPServer(routers []http.Handler, port string) *Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})

	for i := range routers {
		mux.Handle("/", routers[i])
	}

	return &Server{
		Server: &http.Server{
			ReadTimeout: 10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Addr: ":" + port,
			Handler: mux,
		},
	}
}

func (s *Server) Run() error {
	if err := s.Server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
