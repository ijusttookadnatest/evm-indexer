package rest

import (
	"encoding/json"
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
	"net/http"
	"strconv"
)

type EventHandler struct {
	service ports.EventService
}

func NewEventHandler(service ports.EventService) *EventHandler {
	return &EventHandler{service: service}
}

func extractEventFilter(r *http.Request) (domain.EventFilter, error) {
	query := r.URL.Query()
	filter := domain.EventFilter{}

	emitter := query.Get("address")
	topic0 := query.Get("topic0")
	topic1 := query.Get("topic1")
	topic2 := query.Get("topic2")
	topic3 := query.Get("topic3")
	fromBlock := query.Get("fromBlock")
	toBlock := query.Get("toBlock")
	limit := query.Get("limit")

	hasEmitter := emitter != ""
	hasTopics := topic0 != "" || topic1 != "" || topic2 != "" || topic3 != ""

	if !hasEmitter && !hasTopics {
		return filter, errInvalidParams
	}

	if (fromBlock != "") != (toBlock != "") {
		return filter, errInvalidParams
	}

	if hasEmitter {
		filter.Emitter = &emitter
	}

	var topics []string
	for _, t := range []string{topic0, topic1, topic2, topic3} {
		if t != "" {
			topics = append(topics, t)
		}
	}
	filter.Topics = topics

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

func (handler *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	filter, err := extractEventFilter(r)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	events, err := handler.service.GetByEventFilter(filter)
	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (handler *EventHandler) GetEventByTxLog(w http.ResponseWriter, r *http.Request) {
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

	event, err := handler.service.GetByTxHashLogIndex(txHash, logIndex)
	if err != nil {
		writeResErrorToHTTP(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}
