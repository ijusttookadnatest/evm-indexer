package rest

import (
	"encoding/json"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
	"github/ijusttookadnatest/indexer-evm/internal/core/ports"
	"net/http"
	"strconv"
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

	var blocks []domain.BlockTxs
	var block *domain.BlockTxs

	switch blockDTO.groupParam {
	case IdParam:
		block, err = handler.service.GetBlockById(blockDTO.id, blockDTO.tx)
	case HashParam:
		block, err = handler.service.GetBlockByHash(blockDTO.hash, blockDTO.tx)
	case FromOffsetParam:
		blocks, err = handler.service.GetBlocksWithOffset(blockDTO.from, blockDTO.offset, blockDTO.tx)
	case FromToTimeParam:
		blocks, err = handler.service.GetBlocksByRangeTime(blockDTO.fromTime, blockDTO.toTime, blockDTO.tx)
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

	events, err := handler.service.GetEventsByFilter(filter)
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

	event, err := handler.service.GetEventByTxHashLogIndex(txHash, logIndex)
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

	txs, err := handler.service.GetTransactionsByFilter(filter)
	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}
