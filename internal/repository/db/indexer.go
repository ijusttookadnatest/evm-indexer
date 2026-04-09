package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github/ijusttookadnatest/evm-indexer/internal/core/domain"

	"github.com/lib/pq"
)

type IndexerRepository struct {
	db *sql.DB
}

func NewIndexerRepository(db *sql.DB) *IndexerRepository {
	return &IndexerRepository{db: db}
}

func (repo *IndexerRepository) Create(ctx context.Context, block domain.Block, txs []domain.Transaction, events []domain.Event) error {
	tx, err := repo.db.BeginTx(ctx, nil)
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

func (repo *IndexerRepository) BulkCreate(ctx context.Context, items []domain.BlockTxsEvents) error {
	start := time.Now()
	sqlTx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlTx.Rollback()

	blockHashes := make([]string, len(items))
	blockIds := make([]uint64, len(items))
	parentHashes := make([]string, len(items))
	gasLimits := make([]uint64, len(items))
	gasUseds := make([]uint64, len(items))
	miners := make([]string, len(items))
	timestamps := make([]uint64, len(items))
	for i, item := range items {
		blockHashes[i] = item.Block.Hash
		blockIds[i] = item.Block.Id
		parentHashes[i] = item.Block.ParentHash
		gasLimits[i] = item.Block.GasLimit
		gasUseds[i] = item.Block.GasUsed
		miners[i] = item.Block.Miner
		timestamps[i] = item.Block.Timestamp
	}
	_, err = sqlTx.Exec(`
		INSERT INTO blocks(block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp)
		SELECT * FROM UNNEST($1::text[], $2::bigint[], $3::text[], $4::bigint[], $5::bigint[], $6::text[], $7::bigint[])
		ON CONFLICT (block_hash) DO NOTHING;
	`, pq.Array(blockHashes), pq.Array(blockIds), pq.Array(parentHashes),
		pq.Array(gasLimits), pq.Array(gasUseds), pq.Array(miners), pq.Array(timestamps))
	if err != nil {
		return err
	}

	var flatTxs []domain.Transaction
	for _, item := range items {
		flatTxs = append(flatTxs, item.Txs...)
	}
	if len(flatTxs) > 0 {
		txBlockIds := make([]uint64, len(flatTxs))
		txHashes := make([]string, len(flatTxs))
		txFroms := make([]string, len(flatTxs))
		txTos := make([]*string, len(flatTxs))
		txGasUseds := make([]uint64, len(flatTxs))
		for i, t := range flatTxs {
			txBlockIds[i] = t.BlockId
			txHashes[i] = t.Hash
			txFroms[i] = t.From
			txTos[i] = t.To
			txGasUseds[i] = t.GasUsed
		}
		_, err = sqlTx.Exec(`
			INSERT INTO transactions(block_id, tx_hash, from_addr, to_addr, gas_used)
			SELECT * FROM UNNEST($1::bigint[], $2::text[], $3::text[], $4::text[], $5::bigint[])
			ON CONFLICT (tx_hash) DO NOTHING;
		`, pq.Array(txBlockIds), pq.Array(txHashes), pq.Array(txFroms), pq.Array(txTos), pq.Array(txGasUseds))
		if err != nil {
			return err
		}
	}

	var flatEvents []domain.Event
	for _, item := range items {
		flatEvents = append(flatEvents, item.Events...)
	}
	if len(flatEvents) > 0 {
		evBlockIds := make([]uint64, len(flatEvents))
		evLogIndexes := make([]uint64, len(flatEvents))
		evTxHashes := make([]string, len(flatEvents))
		evEmitters := make([]string, len(flatEvents))
		evDatas := make([]string, len(flatEvents))
		evTopics := make([]string, len(flatEvents))
		for i, e := range flatEvents {
			evBlockIds[i] = e.BlockId
			evLogIndexes[i] = e.LogIndex
			evTxHashes[i] = e.TxHash
			evEmitters[i] = e.Emitter
			evDatas[i] = e.Datas
			val, err := pq.Array(e.Topics).Value()
			if err != nil {
				return err
			}
			switch v := val.(type) {
			case []byte:
				evTopics[i] = string(v)
			case string:
				evTopics[i] = v
			default:
				evTopics[i] = "{}"
			}
		}
		_, err = sqlTx.Exec(`
			INSERT INTO events(block_id, log_index, tx_hash, emitter, datas, topics)
			SELECT block_id, log_index, tx_hash, emitter, datas, topics::text[]
			FROM UNNEST($1::bigint[], $2::bigint[], $3::text[], $4::text[], $5::text[], $6::text[])
			AS t(block_id, log_index, tx_hash, emitter, datas, topics)
			ON CONFLICT (tx_hash, log_index) DO NOTHING;
		`, pq.Array(evBlockIds), pq.Array(evLogIndexes), pq.Array(evTxHashes),
			pq.Array(evEmitters), pq.Array(evDatas), pq.Array(evTopics))
		if err != nil {
			return err
		}
	}

	if err := sqlTx.Commit(); err != nil {
		return err
	}

	totalTxs := 0
	totalEvents := 0
	for _, item := range items {
		totalTxs += len(item.Txs)
		totalEvents += len(item.Events)
	}
	slog.Info("BulkCreate done",
		"blocks", len(items),
		"txs", totalTxs,
		"events", totalEvents,
		"duration", time.Since(start),
	)
	return nil
}

func (repo *IndexerRepository) GetBackfillCursor(ctx context.Context) (uint64, error) {
	var id uint64
	err := repo.db.QueryRowContext(ctx, `SELECT last_block_id FROM backfill_cursor;`).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *IndexerRepository) UpdateBackfillCursor(ctx context.Context, blockId uint64) error {
	_, err := repo.db.ExecContext(ctx, `UPDATE backfill_cursor SET last_block_id = $1;`, blockId)
	return err
}

func (repo *IndexerRepository) ResetBackfillCursor(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `UPDATE backfill_cursor SET last_block_id = 0;`)
	return err
}

func (repo *IndexerRepository) Delete(ctx context.Context, blockId uint64) error {
	_, err := repo.db.ExecContext(ctx, `DELETE FROM blocks WHERE block_id = $1;`, blockId)
	return err
}

func (repo *IndexerRepository) GetBlockById(ctx context.Context, id uint64) (*domain.Block, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT block_hash, block_id, parent_hash, gas_limit, gas_used, miner, block_timestamp
		FROM blocks WHERE block_id = $1;
	`, id)
	var b domain.Block
	err := row.Scan(&b.Hash, &b.Id, &b.ParentHash, &b.GasLimit, &b.GasUsed, &b.Miner, &b.Timestamp)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (repo *IndexerRepository) GetBalancefillCursor(ctx context.Context) (uint64, error) {
	var id uint64
	err := repo.db.QueryRowContext(ctx, `SELECT last_block_id FROM balancefill_cursor;`).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *IndexerRepository) UpdateBalancefillCursor(ctx context.Context, blockId uint64) error {
	_, err := repo.db.ExecContext(ctx, `UPDATE balancefill_cursor SET last_block_id = $1;`, blockId)
	return err
}

func (repo *IndexerRepository) ResetBalancefillCursor(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `UPDATE balancefill_cursor SET last_block_id = 0;`)
	return err
}

func (repo *IndexerRepository) BatchUpsertBalance(ctx context.Context, entries []domain.BalanceEntry) error {
	var walletAddress, tokenAddress, tokenId, amount []string
	for _, entry := range entries {
		walletAddress = append(walletAddress, entry.WalletAddress)
		tokenAddress = append(tokenAddress, entry.TokenAddress)
		tokenId = append(tokenId, entry.TokenId)
		amount = append(amount, entry.Amount.String())
	}

	sqlTx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer sqlTx.Rollback()

	batchUpsert := `
		INSERT INTO wallet_balance (wallet_address, token_address, token_id, amount)
		SELECT * FROM UNNEST($1::text[], $2::text[], $3::text[], $4::numeric[])
		ON CONFLICT (wallet_address, token_address, token_id)
		DO UPDATE SET amount = wallet_balance.amount + EXCLUDED.amount;`

	if _, err := sqlTx.ExecContext(ctx, batchUpsert, pq.Array(walletAddress), pq.Array(tokenAddress), pq.Array(tokenId), pq.Array(amount)); err != nil {
		return err
	}

	return sqlTx.Commit()
}

func (repo *IndexerRepository) GetMaxIndexedBlock(ctx context.Context) (uint64, error) {
	var maxId uint64
	err := repo.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(id), 0) FROM blocks;`).Scan(&maxId)
	return maxId, err
}

func (repo *IndexerRepository) GetLogsByTopic(ctx context.Context, filter domain.LogFilter) ([]domain.Log, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, emitter, datas, topics
		FROM events
		WHERE topics[1] = ANY($1) AND id >= $2
		LIMIT $3;`, pq.Array(filter.Topics), filter.From, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return fetchLogs(rows)
}