# sharded-lru-cache

A high-performance, thread-safe, sharded LRU cache for Go with a real-time web dashboard.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Next.js   │────▶│  Go API      │────▶│  Sharded     │
│   Dashboard │◀────│  + WebSocket │◀────│  LRU Cache   │
└─────────────┘     └──────────────┘     └──────────────┘
   :3000                :8080                16 shards
```

### Cache Engine (`cache/`)
- **Sharded design** — 16 independent sub-caches with per-shard mutexes
- **O(1) operations** — doubly-linked list + hashmap
- **LRU eviction** — least recently used entries evicted at capacity
- **Generic** — works with any `comparable` key and `any` value type
- **Metrics** — hit rate, miss rate, eviction count, ops tracking

### API Server (`api/`)
- REST endpoints: `GET`, `PUT`, `DELETE`, `CLEAR`, `STATE`, `METRICS`
- WebSocket for real-time cache state updates (500ms broadcast)
- CORS enabled, configurable capacity/shards

### Dashboard (`ui/`)
- Live shard distribution visualization
- Real-time hit rate chart (sliding 60s window)
- Interactive key/value management
- Connection status indicator

## Quick Start

```bash
# Docker
docker compose up

# Or manually
go run ./cmd/server    # API on :8080
cd ui && npm run dev   # Dashboard on :3000
```

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/state` | GET | Full cache state (entries, shards, metrics) |
| `/api/keys` | GET | List all keys |
| `/api/put` | POST | Add/update key `{key, value}` |
| `/api/get` | POST | Lookup key `{key}` |
| `/api/delete` | POST | Remove key `{key}` |
| `/api/clear` | POST | Flush entire cache |
| `/api/metrics` | GET | Performance metrics |
| `/ws` | WebSocket | Real-time state stream |

## Configuration

| Env Var | Default | Description |
|---------|---------|-------------|
| `CACHE_CAPACITY` | 1000 | Total cache capacity |
| `CACHE_SHARDS` | 16 | Number of shards |
| `PORT` | 8080 | API server port |

## Benchmarks

```
BenchmarkLRUPut-10          125M    9.3 ns/op    0 B/op   0 allocs
BenchmarkShardedPut-10       7.6M  155   ns/op   65 B/op   3 allocs
BenchmarkShardedConcurrent   5.8M  207   ns/op   15 B/op   1 alloc
```

## Tech Stack

- **Backend:** Go 1.22, gorilla/websocket, standard library HTTP
- **Frontend:** Next.js 16, TypeScript, Tailwind CSS
- **Infra:** Docker, GitHub Actions CI/CD
