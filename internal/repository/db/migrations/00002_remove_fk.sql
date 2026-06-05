-- +goose Up
-- +goose StatementBegin

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_block_id_fkey;
ALTER TABLE transactions ALTER COLUMN block_id SET NOT NULL;

ALTER TABLE events DROP CONSTRAINT IF EXISTS events_tx_hash_fkey;
ALTER TABLE events ALTER COLUMN tx_hash SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE transactions ADD CONSTRAINT transactions_block_id_fkey
    FOREIGN KEY (block_id) REFERENCES blocks(block_id) ON DELETE CASCADE;

ALTER TABLE events ADD CONSTRAINT events_tx_hash_fkey
    FOREIGN KEY (tx_hash) REFERENCES transactions(tx_hash) ON DELETE CASCADE;

-- +goose StatementEnd
