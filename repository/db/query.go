package repository

import (
	"database/sql"
	"github/ijusttookadnatest/indexer-evm/core/domain"

	"github.com/lib/pq"
)

type QueryRepository struct {
	db *sql.DB
}

func NewQueryRepository(db *sql.DB) *QueryRepository {
	return &QueryRepository{db :db}
}

func (repo *QueryRepository) GetBlockById(id uint64) (*domain.Block,error) {
	rows, err := repo.db.Query(`SELECT * FROM blocks WHERE block_id = $1;`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks, err := fetchBlocks(rows)
	if err != nil {
		return nil, err
	}
	return &blocks[0], nil
}

func (repo *QueryRepository) GetBlockByHash(hash string) (*domain.Block,error) {
	rows, err := repo.db.Query(`SELECT * FROM blocks WHERE block_hash = $1;`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks, err := fetchBlocks(rows)
	if err != nil {
		return nil, err
	}
	return &blocks[0], nil
}

func (repo *QueryRepository) GetBlocksByRangeId(from, to uint64) ([]domain.Block,error) {
	rows, err := repo.db.Query(`SELECT * FROM blocks  WHERE block_id > $1 AND block_id < $2;`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchBlocks(rows)
}

func (repo *QueryRepository) GetBlocksByRangeTime(from, to uint64) ([]domain.Block,error) {
	rows, err := repo.db.Query(`SELECT * FROM blocks WHERE block_timestamp > $1 AND block_timestamp < $2;`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchBlocks(rows)
}

func (repo *QueryRepository) GetEventsByFilter(filter domain.EventFilter) ([]domain.Event,error) {
	rows, err := repo.db.Query(`
		SELECT *
		FROM events
		WHERE
			($1::TEXT IS NULL OR tx_hash = $1)
			AND ($2::TEXT IS NULL OR emitter = $2)
			AND ($3::TEXT[] IS NULL OR topics @> $3)
			AND ($4::BIGINT IS NULL OR block_id > $4)
			AND ($5::BIGINT IS NULL OR block_id < $5);
		`, filter.TxHash, filter.Emitter, pq.Array(filter.Topics), filter.FromBlock, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchEvents(rows)
}

func (repo *QueryRepository) GetEventsByTxHashLogIndex(txHash string, logIndex int) ([]domain.Event,error) {
	rows, err := repo.db.Query(`SELECT * FROM events  WHERE tx_hash = $1 AND log_index = $2;`, txHash, logIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchEvents(rows)
}


func (repo *QueryRepository) GetTransactionsByFilter(filter domain.TransactionFilter) ([]domain.Transaction,error) {
	rows, err := repo.db.Query(`
		SELECT *
		FROM transactions
		WHERE
			($1::BIGINT IS NULL OR block_id = $1)
			AND ($2::TEXT IS NULL OR tx_hash = $2)
			AND ($3::TEXT IS NULL OR from_addr = $3)
			AND ($4::TEXT IS NULL OR to_addr = $4)
			AND ($5::BIGINT IS NULL OR block_id > $5)
			AND ($6::BIGINT IS NULL OR block_id < $6)
			AND ($7::INT IS NULL OR block_id > $7);
		`, filter.BlockId, filter.Hash, filter.From, filter.To, filter.FromBlock, filter.ToBlock, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchTxs(rows)
}

func (repo *QueryRepository) GetTransactionsByBatchBlocksId(blocksId []uint64) ([]domain.Transaction,error) {
	rows, err := repo.db.Query(`SELECT * FROM transactions WHERE block_id = ANY($1);`, pq.Array(blocksId))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchTxs(rows)
}

func (repo *QueryRepository) GetEventsByBatchTxsHash(txsHash []string) ([]domain.Event,error) {
	rows, err := repo.db.Query(`SELECT * FROM events WHERE tx_hash = ANY($1);`, pq.Array(txsHash))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchEvents(rows)
}
