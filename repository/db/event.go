package repository

import (
	"database/sql"
	"github/ijusttookadnatest/indexer-evm/core/domain"

	"github.com/lib/pq"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

func (repo *EventRepository) GetByEventFilter(filter domain.EventFilter) ([]domain.Event,error) {
	rows, err := repo.db.Query(`
		SELECT block_id, log_index, tx_hash, emitter, datas, topics
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
	events, err := fetchEvents(rows)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (repo *EventRepository) GetByTxHashLogIndex(txHash string, logIndex int) ([]domain.Event,error) {
	rows, err := repo.db.Query(`
		SELECT block_id, log_index, tx_hash, emitter, datas, topics
		FROM events 
		WHERE tx_hash = $1 AND log_index = $2;
		`, txHash, logIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events, err := fetchEvents(rows)
	if err != nil {
		return nil, err
	}
	return events, nil
}