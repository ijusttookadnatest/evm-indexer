-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS blocks (
    block_id BIGINT NOT NULL,
    block_hash VARCHAR(66) NOT NULL,
    parent_hash VARCHAR(66) NOT NULL,
    gas_limit BIGINT NOT NULL,
    gas_used BIGINT NOT NULL,
    miner VARCHAR(42) NOT NULL,
    block_timestamp BIGINT NOT NULL,
    PRIMARY KEY(block_id)
);

CREATE TABLE IF NOT EXISTS transactions (
    block_id INT REFERENCES blocks(block_id) ON DELETE CASCADE,
    tx_hash VARCHAR(66) NOT NULL,
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

CREATE INDEX idx_block_id ON blocks(block_id);
CREATE INDEX idx_block_timestamp ON blocks(block_timestamp);

CREATE INDEX idx_tx_block_id ON transactions(block_id);
CREATE INDEX idx_tx_to ON transactions(to_addr);
CREATE INDEX idx_tx_from ON transactions(from_addr);

CREATE INDEX idx_event_emitter ON events(emitter);
CREATE INDEX idx_event_gin_topics ON events USING GIN(topics);

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

DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS blocks;

-- +goose StatementEnd
