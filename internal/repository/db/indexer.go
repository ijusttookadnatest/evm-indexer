package repository

import (
	"database/sql"
	"github/ijusttookadnatest/indexer-evm/internal/core/domain"

	"github.com/lib/pq"
)

type IndexerRepository struct {
	db *sql.DB
}

func NewIndexerRepository(db *sql.DB) *IndexerRepository {
	return &IndexerRepository{db: db}
}

func (repo *IndexerRepository) Create(block domain.Block, txs []domain.Transaction, events []domain.Event) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT into blocks(block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (block_hash) DO NOTHING;
	`, block.Hash, block.Id, block.ParentHash, block.GasLimit, block.GasUsed, block.Miner, block.Timestamp)
	if err != nil {
		return err
	}

	stmtTx, err := tx.Prepare(`
		INSERT into transactions(block_id, tx_hash, from_addr, to_addr, gas_used)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tx_hash) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	defer stmtTx.Close()
	for _, tx := range txs {
		_, err = stmtTx.Exec(tx.BlockId, tx.Hash, tx.From, tx.To, tx.GasUsed)
		if err != nil {
			return err
		}
	}

	stmtEvent, err := tx.Prepare(`
		INSERT into events(block_id, log_index, tx_hash, emitter, datas, topics)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tx_hash, log_index) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	defer stmtEvent.Close()
	for _, event := range events {
		_, err = stmtEvent.Exec(event.BlockId, event.LogIndex, event.TxHash, event.Emitter, event.Datas, pq.Array(event.Topics))
		if err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func (repo *IndexerRepository) GetBackfillCursor() (uint64, error) {
	var id uint64
	err := repo.db.QueryRow(`SELECT last_block_id FROM backfill_cursor;`).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *IndexerRepository) UpdateBackfillCursor(blockId uint64) error {
	_, err := repo.db.Exec(`UPDATE backfill_cursor SET last_block_id = $1;`, blockId)
	return err
}

func (repo *IndexerRepository) Delete(blockId int) error {
	_, err := repo.db.Exec(`
		DELETE FROM blocks WHERE block_id = $1;
	`, blockId)
	if err != nil {
		return err
	}
	return nil
}
