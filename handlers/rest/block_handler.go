package rest

import (
	"encoding/json"
	"errors"
	"github/ijusttookadnatest/indexer-evm/core/domain"
	"github/ijusttookadnatest/indexer-evm/core/ports"
	"net/http"
	"strconv"
)

type BlockHandler struct {
	service ports.BlockService
}

func NewBlockHandler(service ports.BlockService) *BlockHandler {
	return &BlockHandler{service: service}
}

var (
	IdParam = 1
	HashParam = 2
	FromToBlockParam = 3
	FromToTimeParam = 4
)

type groupParam = int 

type blockDTO struct {
	id uint64
	hash string
	fromBlock uint64
	toBlock uint64
	fromTime uint64
	toTime uint64
	tx bool
	groupParam groupParam
}

func extractDTO(r *http.Request) (*blockDTO,error) {
	query := r.URL.Query()

	var err error = errors.New("invalid parameters")
	block := blockDTO{}
	group := 0

	id := query.Get("id")
	hash := query.Get("hash")
	fromBlock := query.Get("fromBlock")
	toBlock := query.Get("toBlock")
	fromTime := query.Get("fromTime")
	toTime := query.Get("toTime")
	tx := query.Get("tx")

	if id != "" { group++ ; block.groupParam = IdParam }
	if hash != "" { group++ ; block.groupParam = HashParam }
	if fromBlock != "" && toBlock != "" { group++ ; block.groupParam = FromToBlockParam }
	if fromTime != "" && toTime != "" { group++ ; block.groupParam = FromToTimeParam }

	if group != 1 {
		return nil, err
	}
	
	if block.groupParam == HashParam {
		block.hash = hash
	} else if block.groupParam == IdParam {
		block.id, err = strconv.ParseUint(id, 10, 64)
	} else if block.groupParam == FromToBlockParam {
		block.fromBlock, err = strconv.ParseUint(fromBlock, 10, 64)
		block.toBlock, err = strconv.ParseUint(toBlock, 10, 64)
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

func writeResErrorToHTTP(err error, w http.ResponseWriter) {
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, "block not found", http.StatusNotFound)
	} else if errors.Is(err, domain.ErrInvalidHash) || errors.Is(err, domain.ErrInvalidId) || errors.Is(err, domain.ErrInvalidBlockRange) || errors.Is(err, domain.ErrInvalidTimeRange) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *BlockHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	blockDTO, err := extractDTO(r)
	if err != nil {
		http.Error(w,"invalid query parameters", http.StatusBadRequest)
		return
	}

	var blocks []domain.BlockTxs
	var block *domain.BlockTxs

	switch blockDTO.groupParam {
		case IdParam: block, err = handler.service.GetById(blockDTO.id, blockDTO.tx)
		case HashParam: block, err = handler.service.GetByHash(blockDTO.hash, blockDTO.tx)
		case FromToBlockParam: blocks, err = handler.service.GetByRangeId(blockDTO.fromBlock, blockDTO.toBlock, blockDTO.tx)
		case FromToTimeParam: blocks, err = handler.service.GetByRangeTime(blockDTO.fromTime, blockDTO.toTime, blockDTO.tx)
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