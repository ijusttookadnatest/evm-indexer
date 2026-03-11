package rest

import (
	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	"net/http"
)

func NewRouter(service ports.QueryService) http.Handler {
	mux := http.NewServeMux()
	handler := NewHandler(service)

	mux.HandleFunc("GET api/blocks", handler.GetBlock)
	mux.HandleFunc("GET api/transactions", handler.GetTransaction)
	mux.HandleFunc("GET api/events", handler.GetEvent)
	mux.HandleFunc("GET api/events/log", handler.GetEventByTxLog)

	return mux
}
