package rest

import (
	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	"net/http"
)

func NewRouter(service ports.QueryService) http.Handler {
	mux := http.NewServeMux()
	handler := NewHandler(service)

	mux.HandleFunc("GET /blocks", handler.GetBlock)
	mux.HandleFunc("GET /transactions", handler.GetTransaction)
	mux.HandleFunc("GET /events", handler.GetEvent)
	mux.HandleFunc("GET /events/log", handler.GetEventByTxLog)

	return mux
}
