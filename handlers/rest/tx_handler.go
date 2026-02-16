package rest

import (
	"encoding/json"
	"errors"
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
	"net/http"
	"strconv"
)

var errInvalidParams = errors.New("invalid or conflicting query parameters")

type TransactionHandler struct {
	service ports.TransactionService
}

func NewTransactionHandler(service ports.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func extractTxFilter(r *http.Request) (domain.TransactionFilter, error) {
	query := r.URL.Query()
	filter := domain.TransactionFilter{}

	hash := query.Get("hash")
	from := query.Get("from")
	to := query.Get("to")
	fromBlock := query.Get("fromBlock")
	toBlock := query.Get("toBlock")
	limit := query.Get("limit")

	hasHash := hash != ""
	hasRange := fromBlock != "" || toBlock != ""
	hasAddressFilter := from != "" || to != ""

	if hasHash && (hasRange || hasAddressFilter) {
		return filter, errInvalidParams
	}

	if (fromBlock != "") != (toBlock != "") {
		return filter, errInvalidParams
	}

	if hasAddressFilter && !hasRange {
		return filter, errInvalidParams
	}

	if !hasHash && !hasRange {
		return filter, errInvalidParams
	}

	if hasHash {
		filter.Hash = &hash
	}
	if from != "" {
		filter.From = &from
	}
	if to != "" {
		filter.To = &to
	}
	if fromBlock != "" {
		v, err := strconv.ParseUint(fromBlock, 10, 64)
		if err != nil {
			return filter, err
		}
		filter.FromBlock = &v
	}
	if toBlock != "" {
		v, err := strconv.ParseUint(toBlock, 10, 64)
		if err != nil {
			return filter, err
		}
		filter.ToBlock = &v
	}
	if limit != "" {
		v, err := strconv.Atoi(limit)
		if err != nil {
			return filter, err
		}
		filter.Limit = &v
	}

	return filter, nil
}

func (handler *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	filter, err := extractTxFilter(r)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	txs, err := handler.service.GetByTransactionFilter(filter)
	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}
