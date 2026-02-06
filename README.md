# evm-indexer

## Why block-by-block instead of FilterLogs?

FilterLogs would be simpler for this use case, but I deliberately
chose block-by-block processing to implement production patterns:

- **Message queue** (Redis streams) for backpressure handling
- **Worker pool** with configurable concurrency
- **Idempotent processing** for crash recovery
- **Dead letter queue** for failed blocks

This architecture scales horizontally and mirrors real-world
indexers like Etherscan or The Graph.