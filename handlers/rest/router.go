package rest

import (
	"github/ijusttookadnatest/indexer-evm/core/ports"
	"net/http"
)

func newRouter(service ports.QueryService) http.Handler {
	mux := http.NewServeMux()
	handler := NewHandler(service)

	mux.HandleFunc("GET api/blocks", handler.GetBlock)
	mux.HandleFunc("GET api/transactions", handler.GetTransaction)
	mux.HandleFunc("GET api/events", handler.GetEvent)
	mux.HandleFunc("GET api/events/log", handler.GetEventByTxLog)
	
	return mux
}
