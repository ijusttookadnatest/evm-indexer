# evm-indexer

EVM blockchain indexer (Ethereum/Arbitrum) written in Go. Indexes blocks, transactions, and events from an RPC node into PostgreSQL, and exposes them via REST, GraphQL, and WebSocket APIs.

## Architecture

Two separate binaries with hexagonal architecture:

- **`cmd/indexer`** — pulls data from the chain (backfill + live) and stores it
- **`cmd/api`** — serves the indexed data via HTTP

```
RPC Node ──► Indexer ──► PostgreSQL ──► REST / GraphQL
                  └──► Redis (pub/sub) ──► Websocket
```

## Prerequisites

- Go 1.25+
- PostgreSQL
- Redis
- An EVM RPC endpoint (HTTP + WebSocket)

## Setup

**1. Start dependencies**

```bash
docker compose -f docker-compose.db.yml up -d
```

**2. Configure environment**

Edit `.env` in each binary directory:

**Indexer** (`cmd/indexer/.env`):

| Variable             | Description                         |
|----------------------|-------------------------------------|
| `POSTGRES_*`         | PostgreSQL connection               |
| `REDIS_*`            | Redis connection                    |
| `RPC_HTTP`           | HTTP RPC endpoint                   |
| `RPC_WS`             | WebSocket RPC endpoint              |
| `RPC_RATE_LIMIT`     | Requests/sec cap                    |
| `FROM`               | Start block number for backfill     |
| `CONCURRENCY_FACTOR` | Parallel block fetch workers        |

**API** (`cmd/api/.env`):

| Variable             | Description                         |
|----------------------|-------------------------------------|
| `POSTGRES_*`         | PostgreSQL connection               |
| `REDIS_*`            | Redis connection                    |
| `PORT`               | HTTP listen port (default 8080)     |
| `PLAYGROUND_ENABLED` | Enable GraphQL playground           |
| `MAX_TIME`           | Max time range for queries (secs)   |
| `MAX_OFFSET`         | Max pagination offset               |

## Run

```bash
# Start indexer
cd cmd/indexer && go run main.go

# Re-index from scratch
cd cmd/indexer && go run main.go -reindex

# Start API
cd cmd/api && go run main.go
```

Migrations run automatically on startup.

## API

### REST

| Endpoint            | Description                          |
|---------------------|--------------------------------------|
| `GET /blocks`       | Query blocks (by id, time range)     |
| `GET /transactions` | Query transactions (by from/to addr) |
| `GET /events`       | Query events (by emitter, topics)    |
| `GET /events/log`   | Query event by tx hash + log index   |

### GraphQL (`/query`)

Playground at `/` when `PLAYGROUND_ENABLED=true`. Supports queries on `blocks`, `transactions`, `events` with filters.

### WebSocket (`/ws`)

Real-time stream of new blocks/transactions/events via Redis pub/sub.

## Metrics

Prometheus metrics exposed on:

- Indexer: `:2112/metrics`
- API: `:2113/metrics`

## Tests

```bash
# Unit tests
go test ./...

# Integration tests (require running DB)
docker compose -f docker-compose.db.test.yml up -d
go test ./internal/repository/db/...
```

## Schema

Three tables: `blocks → transactions → events` (cascade deletes).  
GIN index on `events.topics` for efficient topic filtering.
