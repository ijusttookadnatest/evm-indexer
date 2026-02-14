  INSERT INTO blocks (block_id, block_hash, parent_hash, gas_limit,
  gas_used, miner, block_timestamp)
  VALUES
      (100, '0xblock100', '0xparent99',  30000000, 15000000, '0xminer1',
  1700000000),
      (101, '0xblock101', '0xblock100',  30000000, 12000000, '0xminer2',
  1700000012),
      (102, '0xblock102', '0xblock101',  30000000, 18000000, '0xminer1',
  1700000024);

  -- Transactions liées aux blocks
  INSERT INTO transactions (block_id, tx_hash, from_addr, to_addr, gas_used)
  VALUES
      (100, '0xtx1', '0xAlice', '0xBob',      21000),
      (100, '0xtx2', '0xAlice', '0xContract',  50000),
      (101, '0xtx3', '0xBob',   '0xAlice',     21000);

  -- Events liés aux transactions
  INSERT INTO events (block_id, log_index, tx_hash, emitter, datas, topics)
  VALUES
      (100, 0, '0xtx2', '0xContract', '"0xdata1"',
  '{"0xTransferSig","0xAlice","0xBob"}'),
      (100, 1, '0xtx2', '0xContract', '"0xdata2"',
  '{"0xApprovalSig","0xAlice"}');