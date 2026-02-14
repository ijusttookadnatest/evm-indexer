package repository

import (
	"database/sql"
	"github/ijusttookadnatest/indexer-evm/domain"

	"github.com/lib/pq"
)

type BlockRepository struct {
	db *sql.DB
}

func NewBlockRepository(db *sql.DB) *BlockRepository {
	return &BlockRepository{db :db}
}

func (repo *BlockRepository) GetByNumber(number uint64) (*domain.Block,error) {
	rows, err := repo.db.Query(`
		SELECT block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp
		FROM blocks 
		WHERE block_id = $1;
	`, number)
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

func (repo *BlockRepository) GetByHash(hash string) (*domain.Block,error) {
	rows, err := repo.db.Query(`
		SELECT block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp
		FROM blocks 
		WHERE block_hash = $1;
	`, hash)
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

func (repo *BlockRepository) GetByRangeNumber(from, to uint64) ([]domain.Block,error) {
	rows, err := repo.db.Query(`
		SELECT block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp
		FROM blocks 
		WHERE block_id > $1 AND block_id < $2;
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks, err := fetchBlocks(rows)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (repo *BlockRepository) GetByRangeTime(from, to uint64) ([]domain.Block,error) {
	rows, err := repo.db.Query(`
		SELECT block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp
		FROM blocks 
		WHERE block_timestamp > $1 AND block_timestamp < $2;
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks, err := fetchBlocks(rows)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (repo *BlockRepository) Create(block domain.Block, txs []domain.Transaction, events []domain.Event) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT into blocks(block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`, block.Hash, block.Id, block.ParentHash, block.GasLimit, block.GasUsed, block.Miner, block.Timestamp)
	if err != nil {
		return err
	}

	stmtTx, err := tx.Prepare(`
		INSERT into transactions(block_id, tx_hash, from_addr, to_addr, gas_used)
		VALUES ($1, $2, $3, $4, $5);
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
		VALUES ($1, $2, $3, $4, $5, $6);
	`)
	if err != nil {
		return err
	}
	defer stmtEvent.Close()
	for _, event := range  events {
		_, err = stmtEvent.Exec(event.BlockId, event.LogIndex, event.TxHash, event.Emitter, pq.Array(event.Datas), pq.Array(event.Topics))
		if err != nil {
			return err
		}
	}
	
	tx.Commit()
	return nil
}

func (repo *BlockRepository) Delete(blockId int) error {
	_, err := repo.db.Exec(`
		DELETE FROM blocks WHERE block_id = $1;
	`, blockId)
	if err != nil {
		return err
	}
	return nil
}