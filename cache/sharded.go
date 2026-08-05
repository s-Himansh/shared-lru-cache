package cache

import (
	"hash/fnv"
	"sync"
)

const numShards = 16

// Hasher is a function type used to hash keys into uint32.
type Hasher[K comparable] func(K) uint32

// ShardedCache is a thread-safe, sharded LRU cache.
type ShardedCache[K comparable, V any] struct {
	shards []*lruCache[K, V]
	locks  [numShards]sync.RWMutex
	hasher Hasher[K]
}

// NewShardedCache creates a new ShardedCache.
// totalCapacity is divided evenly across the 16 shards.
func NewShardedCache[K comparable, V any](totalCapacity int, hasher Hasher[K]) *ShardedCache[K, V] {
	if totalCapacity < numShards {
		totalCapacity = numShards
	}
	perShard := totalCapacity / numShards

	sc := &ShardedCache[K, V]{
		shards: make([]*lruCache[K, V], numShards),
		hasher: hasher,
	}

	for i := 0; i < numShards; i++ {
		sc.shards[i] = newLRU[K, V](perShard)
	}
	return sc
}

// getShardIndex determines which shard a key belongs to.
func (sc *ShardedCache[K, V]) getShardIndex(key K) uint32 {
	return sc.hasher(key) % numShards
}

// Get retrieves a value from the cache.
// It uses an RLock, allowing multiple concurrent readers.
func (sc *ShardedCache[K, V]) Get(key K) (V, bool) {
	idx := sc.getShardIndex(key)

	sc.locks[idx].RLock()
	defer sc.locks[idx].RUnlock()

	return sc.shards[idx].Get(key)
}

// Put adds or updates a value in the cache.
// It uses a full Lock, blocking other readers/writers for that specific shard.
func (sc *ShardedCache[K, V]) Put(key K, value V) {
	idx := sc.getShardIndex(key)

	sc.locks[idx].Lock()
	defer sc.locks[idx].Unlock()

	sc.shards[idx].Put(key, value)
}

// StringHasher uses FNV-1a hashing for string keys.
func StringHasher(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// IntHasher casts an int to uint32.
// Note: For high-throughput integer keys, a proper integer hashing function
// (like splitmix32) is recommended to ensure even distribution.
func IntHasher(i int) uint32 {
	return uint32(i)
}
