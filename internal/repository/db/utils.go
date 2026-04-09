package repository

import (
	"database/sql"
	"github/ijusttookadnatest/evm-indexer/internal/core/domain"

	"github.com/lib/pq"
)

func fetchBlocksTxs(rows *sql.Rows) ([]domain.Block, error) {
	blocks := []domain.Block{}
	for rows.Next() {
		block, err := scanBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, *block)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(blocks) == 0 {
		return nil, domain.ErrNotFound
	}
	return blocks, nil
}

func fetchBlocks(rows *sql.Rows) ([]domain.Block, error) {
	blocks := []domain.Block{}
	for rows.Next() {
		block, err := scanBlock(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, *block)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(blocks) == 0 {
		return nil, domain.ErrNotFound
	}
	return blocks, nil
}

func fetchTxs(rows *sql.Rows) ([]domain.Transaction, error) {
	txs := []domain.Transaction{}
	for rows.Next() {
		tx, err := scanTx(rows)
		if err != nil {
			return nil, err
		}
		txs = append(txs, *tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(txs) == 0 {
		return nil, domain.ErrNotFound
	}
	return txs, nil
}

func fetchEvents(rows *sql.Rows) ([]domain.Event, error) {
	events := []domain.Event{}
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(events) == 0 {
		return nil, domain.ErrNotFound
	}
	return events, nil
}

func fetchLogs(rows *sql.Rows) ([]domain.Log, error) {
	events := []domain.Log{}
	for rows.Next() {
		event, err := scanLog(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	return events, nil
}

func scanBlock(row *sql.Rows) (*domain.Block, error) {
	block := new(domain.Block)
	err := row.Scan(
		&block.Hash,
		&block.Id,
		&block.ParentHash,
		&block.GasLimit,
		&block.GasUsed,
		&block.Miner,
		&block.Timestamp,
	)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func scanTx(row *sql.Rows) (*domain.Transaction, error) {
	tx := new(domain.Transaction)
	err := row.Scan(
		&tx.BlockId,
		&tx.Hash,
		&tx.From,
		&tx.To,
		&tx.GasUsed,
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func scanEvent(row *sql.Rows) (*domain.Event, error) {
	event := new(domain.Event)
	err := row.Scan(
		&event.BlockId,
		&event.LogIndex,
		&event.TxHash,
		&event.Emitter,
		&event.Datas,
		pq.Array(&event.Topics),
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func scanLog(row *sql.Rows) (*domain.Log, error) {
	event := new(domain.Log)
	err := row.Scan(
		&event.Id,
		&event.Emitter,
		&event.Datas,
		pq.Array(&event.Topics),
	)
	if err != nil {
		return nil, err
	}
	return event, nil
}
