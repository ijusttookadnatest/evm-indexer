package rest

import (
	"fmt"
	"github/ijusttookadnatest/indexer-evm/core/ports"
	"net/http"
	"time"
)

type Server struct {
	server       *http.Server
}

func NewServer(port int, blockSc ports.BlockService, txSc ports.TransactionService, eventSc ports.EventService) *Server {
	return &Server{
		server: &http.Server{
			ReadTimeout: 10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Addr:fmt.Sprintf(":%v", port),
			Handler: newRouter(blockSc, txSc, eventSc),
		},
	}
}

func (server *Server) Run() error {
	if err := server.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
