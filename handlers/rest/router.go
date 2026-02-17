package rest

import (
	"fmt"
	"github/ijusttookadnatest/indexer-evm/core/ports"
	"net/http"
)

func newRouter(service ports.QueryService) *http.ServeMux {
	mux := http.NewServeMux()

	handler := NewHandler(service)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "ok")
	})
	mux.HandleFunc("GET /v1/blocks", handler.GetBlock)
	mux.HandleFunc("GET /v1/transactions", handler.GetTransaction)
	mux.HandleFunc("GET /v1/events", handler.GetEvent)
	mux.HandleFunc("GET /v1/events/log", handler.GetEventByTxLog)

	return mux
}