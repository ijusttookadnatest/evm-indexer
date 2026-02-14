package repository

import (
	"database/sql"
	"github/ijusttookadnatest/indexer-evm/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (repo *TransactionRepository) GetByTransactionFilter(filter domain.TransactionFilter) ([]domain.Transaction,error) {
	rows, err := repo.db.Query(`
		SELECT block_id, tx_hash, from_addr, to_addr, gas_used
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
	txs, err := fetchTxs(rows)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

