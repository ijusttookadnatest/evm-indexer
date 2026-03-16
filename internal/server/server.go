package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)


type Server struct {
	Server *http.Server
}

func NewHTTPServer(restHandler, wsHandler, graphqlHandler http.Handler, port string) *Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})

	mux.Handle("/api/", http.StripPrefix("/api", restHandler))
	mux.Handle("/ws", wsHandler)
	mux.Handle("/graphql", graphqlHandler)

	return &Server{
		Server: &http.Server{
			ReadTimeout: 10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Addr: ":" + port,
			Handler: mux,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Server.ListenAndServe()
	}()

	select {
        case err := <-errChan:
            slog.Info("Server error: ", "err", err)
        case <-ctx.Done():
            slog.Info("Received shutdown signal: ","reason", ctx.Err())
    }

    slog.Info("Server is shutting down...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := s.Server.Shutdown(ctx); err != nil {
        slog.Info("Server shutdown error: ", "err", err)
        return err
    }

    slog.Info("Server exited properly")
	return nil
}
