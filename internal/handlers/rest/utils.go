package rest

import (
	"errors"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"
	"net/http"
	"strconv"
)

var (
	IdParam          = 1
	HashParam        = 2
	FromOffsetParam  = 3
	FromToTimeParam  = 4
)

type groupParam = int

type blockDTO struct {
	id         uint64
	hash       string
	from       uint64
	offset     uint64
	fromTime   uint64
	toTime     uint64
	tx         bool
	groupParam groupParam
}

func extractBlockDTO(r *http.Request) (*blockDTO, error) {
	query := r.URL.Query()

	var err error = errors.New("invalid parameters")
	block := blockDTO{}
	group := 0

	id := query.Get("id")
	hash := query.Get("hash")
	from := query.Get("from")
	offset := query.Get("offset")
	fromTime := query.Get("fromTime")
	toTime := query.Get("toTime")
	tx := query.Get("tx")

	if id != "" {
		group++
		block.groupParam = IdParam
	}
	if hash != "" {
		group++
		block.groupParam = HashParam
	}
	if from != "" {
		group++
		block.groupParam = FromOffsetParam
	}
	if fromTime != "" && toTime != "" {
		group++
		block.groupParam = FromToTimeParam
	}

	if group != 1 {
		return nil, err
	}

	if block.groupParam == HashParam {
		block.hash = hash
	} else if block.groupParam == IdParam {
		block.id, err = strconv.ParseUint(id, 10, 64)
	} else if block.groupParam == FromOffsetParam {
		block.from, err = strconv.ParseUint(from, 10, 64)
		if offset == "" {
			block.offset = 0
		} else {
			block.offset, err = strconv.ParseUint(offset, 10, 64)
		}
	} else if block.groupParam == FromToTimeParam {
		block.fromTime, err = strconv.ParseUint(fromTime, 10, 64)
		block.toTime, err = strconv.ParseUint(toTime, 10, 64)
	}

	if err != nil {
		return nil, err
	}

	block.tx = (tx == "yes")
	return &block, nil
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

func writeResErrorToHTTP(err error, w http.ResponseWriter) {
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, "block not found", http.StatusNotFound)
	} else if errors.Is(err, domain.ErrInvalidHash) || errors.Is(err, domain.ErrInvalidId) || errors.Is(err, domain.ErrInvalidBlockRange) || errors.Is(err, domain.ErrInvalidTimeRange) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var errInvalidParams = errors.New("invalid or conflicting query parameters")

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
