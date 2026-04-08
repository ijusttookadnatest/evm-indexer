-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS blocks (
    block_id BIGINT NOT NULL UNIQUE,
    block_hash VARCHAR(66) NOT NULL UNIQUE,
    parent_hash VARCHAR(66) NOT NULL,
    gas_limit BIGINT NOT NULL,
    gas_used BIGINT NOT NULL,
    miner VARCHAR(42) NOT NULL,
    block_timestamp BIGINT NOT NULL,
    PRIMARY KEY(block_id)
);

CREATE TABLE IF NOT EXISTS transactions (
    block_id INT REFERENCES blocks(block_id) ON DELETE CASCADE,
    tx_hash VARCHAR(66) NOT NULL UNIQUE,
    from_addr VARCHAR(42) NOT NULL,
    to_addr  VARCHAR(42),
    gas_used BIGINT NOT NULL,
    PRIMARY KEY(tx_hash)
);

CREATE TABLE IF NOT EXISTS events (
    block_id INT NOT NULL,
    log_index INT NOT NULL,
    tx_hash VARCHAR(66) REFERENCES transactions(tx_hash) ON DELETE CASCADE,
    emitter VARCHAR(42) NOT NULL,
    datas TEXT,
    topics TEXT[],
    PRIMARY KEY(tx_hash, log_index)
);

CREATE TABLE IF NOT EXISTS wallet_balance (
    wallet_address TEXT NOT NULL,
    token_address  TEXT NOT NULL,
    token_id       TEXT NOT NULL DEFAULT '',
    amount         NUMERIC NOT NULL,
    PRIMARY KEY (wallet_address, token_address, token_id)
);

CREATE INDEX idx_block_id ON blocks(block_id);
CREATE INDEX idx_block_timestamp ON blocks(block_timestamp);

CREATE INDEX idx_tx_block_id ON transactions(block_id);
CREATE INDEX idx_tx_to ON transactions(to_addr);
CREATE INDEX idx_tx_from ON transactions(from_addr);

CREATE INDEX idx_event_emitter ON events(emitter);
CREATE INDEX idx_event_gin_topics ON events USING GIN(topics);

CREATE TABLE IF NOT EXISTS backfill_cursor (
    last_block_id BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS balancefill_cursor (
    last_block_id BIGINT NOT NULL DEFAULT 0
);

INSERT INTO backfill_cursor (last_block_id) VALUES (0);
INSERT INTO balancefill_cursor (last_block_id) VALUES (0);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_block_id;
DROP INDEX IF EXISTS idx_block_timestamp;
DROP INDEX IF EXISTS idx_tx_block_id;
DROP INDEX IF EXISTS idx_tx_to;
DROP INDEX IF EXISTS idx_tx_from;
DROP INDEX IF EXISTS idx_event_emitter;
DROP INDEX IF EXISTS idx_event_gin_topics;

DROP TABLE IF EXISTS backfill_cursor;
DROP TABLE IF EXISTS balancefill_cursor;

DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS wallet_balance;

-- +goose StatementEnd
