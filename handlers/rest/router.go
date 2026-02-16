package rest

import (
	"github/ijusttookadnatest/indexer-evm/core/ports"
	"net/http"
)

func newRouter(blockSvc ports.BlockService, txSvc ports.TransactionService, eventSvc ports.EventService) *http.ServeMux {
	mux := http.NewServeMux()

	blockHandler := NewBlockHandler(blockSvc)
	txHandler := NewTransactionHandler(txSvc)
	eventHandler := NewEventHandler(eventSvc)

	mux.HandleFunc("GET /v1/blocks", blockHandler.GetBlock)
	mux.HandleFunc("GET /v1/transactions", txHandler.GetTransaction)
	mux.HandleFunc("GET /v1/events", eventHandler.GetEvent)
	mux.HandleFunc("GET /v1/events/log", eventHandler.GetEventByTxLog)

	return mux
}