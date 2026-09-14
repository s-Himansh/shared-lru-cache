package cache

import (
	"hash/fnv"
	"sync"
)

const defaultNumShards = 16

// Hasher is a function type used to hash keys into uint32.
type Hasher[K comparable] func(K) uint32

// ShardedCache is a thread-safe, sharded LRU cache.
type ShardedCache[K comparable, V any] struct {
	shards    []*lruCache[K, V]
	locks     []sync.RWMutex
	hasher    Hasher[K]
	numShards int
}

// Option configures a ShardedCache.
type Option[K comparable, V any] func(*ShardedCache[K, V])

// WithNumShards sets the number of shards. Must be called before using the cache.
// Shards are created in NewShardedCache, so this is only effective there.
func WithNumShards[K comparable, V any](n int) Option[K, V] {
	return func(sc *ShardedCache[K, V]) {
		if n > 0 {
			sc.numShards = n
		}
	}
}

// NewShardedCache creates a new ShardedCache.
// totalCapacity is divided evenly across the shards.
func NewShardedCache[K comparable, V any](totalCapacity int, hasher Hasher[K], opts ...Option[K, V]) *ShardedCache[K, V] {
	sc := &ShardedCache[K, V]{
		numShards: defaultNumShards,
		hasher:    hasher,
	}
	for _, opt := range opts {
		opt(sc)
	}
	if sc.numShards < 1 {
		sc.numShards = defaultNumShards
	}
	if totalCapacity < sc.numShards {
		totalCapacity = sc.numShards
	}
	perShard := totalCapacity / sc.numShards

	sc.shards = make([]*lruCache[K, V], sc.numShards)
	sc.locks = make([]sync.RWMutex, sc.numShards)
	for i := 0; i < sc.numShards; i++ {
		sc.shards[i] = newLRU[K, V](perShard)
	}
	return sc
}

// getShardIndex determines which shard a key belongs to.
func (sc *ShardedCache[K, V]) getShardIndex(key K) int {
	return int(sc.hasher(key) % uint32(sc.numShards))
}

// Get retrieves a value from the cache.
// Uses a full Lock because Get mutates the LRU list (MoveToFront).
func (sc *ShardedCache[K, V]) Get(key K) (V, bool) {
	idx := sc.getShardIndex(key)

	sc.locks[idx].Lock()
	defer sc.locks[idx].Unlock()

	return sc.shards[idx].Get(key)
}

// Put adds or updates a value in the cache.
func (sc *ShardedCache[K, V]) Put(key K, value V) {
	idx := sc.getShardIndex(key)

	sc.locks[idx].Lock()
	defer sc.locks[idx].Unlock()

	sc.shards[idx].Put(key, value)
}

// Delete removes a key from the cache. Returns true if the key existed.
func (sc *ShardedCache[K, V]) Delete(key K) bool {
	idx := sc.getShardIndex(key)

	sc.locks[idx].Lock()
	defer sc.locks[idx].Unlock()

	return sc.shards[idx].Delete(key)
}

// Len returns the total number of items across all shards.
func (sc *ShardedCache[K, V]) Len() int {
	total := 0
	for i := 0; i < sc.numShards; i++ {
		sc.locks[i].RLock()
		total += sc.shards[i].Len()
		sc.locks[i].RUnlock()
	}
	return total
}

// Clear removes all items from all shards.
func (sc *ShardedCache[K, V]) Clear() {
	for i := 0; i < sc.numShards; i++ {
		sc.locks[i].Lock()
		sc.shards[i].Clear()
		sc.locks[i].Unlock()
	}
}

// Keys returns all keys in the cache (snapshot, order may vary).
func (sc *ShardedCache[K, V]) Keys() []K {
	var keys []K
	for i := 0; i < sc.numShards; i++ {
		sc.locks[i].RLock()
		keys = append(keys, sc.shards[i].Keys()...)
		sc.locks[i].RUnlock()
	}
	return keys
}

// StringHasher uses FNV-1a hashing for string keys.
func StringHasher(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// IntHasher uses splitmix32 to hash integer keys for even distribution.
func IntHasher(i int) uint32 {
	x := uint32(i)
	x ^= x >> 16
	x *= 0x45d9f3b
	x ^= x >> 16
	return x
}
