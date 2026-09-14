# shared-lru-cache

A thread-safe, sharded LRU cache for Go, built with generics.

## Features

- **Generic** — works with any `comparable` key and `any` value type
- **Thread-safe** — sharded design with per-shard `RWMutex` for high concurrency
- **LRU eviction** — least recently used entries are evicted when capacity is reached
- **Configurable** — adjustable shard count and capacity
- **Zero dependencies** — standard library only

## Install

```bash
go get github.com/s-Himansh/shared-lru-cache
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/s-Himansh/shared-lru-cache/cache"
)

func main() {
	// Create a cache with 1000 total capacity and string keys
	c := cache.NewShardedCache[string, string](1000, cache.StringHasher)

	// Put values
	c.Put("user:1", "Alice")
	c.Put("user:2", "Bob")

	// Get values
	if val, ok := c.Get("user:1"); ok {
		fmt.Println(val) // "Alice"
	}

	// Delete
	c.Delete("user:1")

	// Check size
	fmt.Println(c.Len()) // 1

	// Custom shard count
	c2 := cache.NewShardedCache[string, int](10000, cache.StringHasher,
		cache.WithNumShards[string, int](32),
	)
	_ = c2
}
```

## API

| Method | Description |
|--------|-------------|
| `NewShardedCache[K, V](capacity, hasher, ...Option)` | Create a new sharded cache |
| `Get(key) (V, bool)` | Retrieve a value (updates recency) |
| `Put(key, value)` | Add or update a value |
| `Delete(key) bool` | Remove a key |
| `Len() int` | Total items across all shards |
| `Clear()` | Remove all items |
| `Keys() []K` | Snapshot of all keys |
| `StringHasher(s string) uint32` | FNV-1a hasher for strings |
| `IntHasher(i int) uint32` | splitmix32 hasher for integers |

## How it Works

Keys are hashed and distributed across 16 (default) independent LRU sub-caches, each protected by its own mutex. This reduces lock contention compared to a single global lock, enabling higher throughput under concurrent access.

Each shard uses a `container/list` (doubly-linked list) for O(1) LRU ordering and a `map` for O(1) lookups.

## Testing

```bash
go test -race ./...
```

## Benchmarks

```bash
go test -bench=. -benchmem ./...
```

