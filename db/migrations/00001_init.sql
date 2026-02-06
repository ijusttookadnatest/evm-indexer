-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS blocks (
    block_hash VARCHAR(42) NOT NULL PRIMARY KEY,
    block_id BIGINT,
    parent_hash VARCHAR(42) NOT NULL,
    gas_limit BIGINT NOT NULL,
    gas_used BIGINT NOT NULL,
    miner VARCHAR(66) NOT NULL,
    block_timestamp BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    block_hash VARCHAR(42) REFERENCES blocks(block_hash) ON DELETE CASCADE,
    tx_hash VARCHAR(66) NOT NULL PRIMARY KEY,
    from_addr VARCHAR(66) NOT NULL,
    to_addr  VARCHAR(66) NOT NULL,
    gas_used BIGINT NOT NULL,
    contract_addr VARCHAR(66) NOT NULL
);

CREATE TABLE IF NOT EXISTS logs (
    id BIGINT NOT NULL PRIMARY KEY,
    tx_hash VARCHAR(66) REFERENCES transactions(tx_hash) ON DELETE CASCADE,
    emitter_addr VARCHAR(66) NOT NULL,
    datas TEXT,
    topics TEXT
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS blocks;

-- +goose StatementEnd
