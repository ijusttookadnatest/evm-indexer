package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github/ijusttookadnatest/evm-indexer/internal/core/domain"
	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
)

type Handler struct {
	service ports.QueryService
}

func NewHandler(service ports.QueryService) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) GetBlock(w http.ResponseWriter, r *http.Request) {
	blockDTO, err := extractBlockDTO(r)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var blocks []domain.BlockTxs
	var block *domain.BlockTxs

	switch blockDTO.groupParam {
	case IdParam:
		block, err = handler.service.GetBlockById(ctx, blockDTO.id, blockDTO.tx)
	case HashParam:
		block, err = handler.service.GetBlockByHash(ctx, blockDTO.hash, blockDTO.tx)
	case FromOffsetParam:
		blocks, err = handler.service.GetBlocksWithOffset(ctx, blockDTO.from, blockDTO.offset, blockDTO.tx)
	case FromToTimeParam:
		blocks, err = handler.service.GetBlocksByRangeTime(ctx, blockDTO.fromTime, blockDTO.toTime, blockDTO.tx)
	}

	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	if block != nil {
		blocks = []domain.BlockTxs{*block}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

func (handler *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	filter, err := extractEventFilter(r)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	events, err := handler.service.GetEventsByFilter(ctx, filter)
	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (handler *Handler) GetEventByTxLog(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	txHash := query.Get("txHash")
	logIndexStr := query.Get("logIndex")

	if txHash == "" || logIndexStr == "" {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	logIndex, err := strconv.Atoi(logIndexStr)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	event, err := handler.service.GetEventByTxHashLogIndex(ctx, txHash, logIndex)
	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (handler *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	filter, err := extractTxFilter(r)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	txs, err := handler.service.GetTransactionsByFilter(ctx, filter)
	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}
