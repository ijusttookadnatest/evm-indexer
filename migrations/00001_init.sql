-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS tokens (
    addr VARCHAR(42) NOT NULL PRIMARY KEY,
    symbol VARCHAR(3) NOT NULL,
    decimals INT NOT NULL,
    deployment_block INT NOT NULL,
    deployment_timestamp BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS transfers (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL,
    token_addr VARCHAR(42) REFERENCES tokens(addr) ON DELETE CASCADE,
    from_addr VARCHAR(42),
    to_addr VARCHAR(42),
    amount BIGINT NOT NULL,
    block_timestamp BIGINT NOT NULL,
    block_number INT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS tokens;

-- +goose StatementEnd
