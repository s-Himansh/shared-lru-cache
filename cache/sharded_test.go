package cache

import (
	"fmt"
	"sync"
	"testing"
)

func TestShardedCache(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		hasher   Hasher[string]
		actions  func(c *ShardedCache[string, string])
		getKey   string
		wantVal  string
		wantOk   bool
	}{
		{
			name:     "Basic Put and Get with StringHasher",
			capacity: 160,
			hasher:   StringHasher,
			actions: func(c *ShardedCache[string, string]) {
				c.Put("user:1", "Alice")
				c.Put("user:2", "Bob")
			},
			getKey:  "user:1",
			wantVal: "Alice",
			wantOk:  true,
		},
		{
			name:     "Eviction across shards (capacity 1 per shard)",
			capacity: 16, // 16 shards * 1 capacity = 16 total
			hasher:   StringHasher,
			actions: func(c *ShardedCache[string, string]) {
				// Both keys hash to the same shard. The first should be evicted.
				// We use similar strings to try to force the same shard (though FNV-1a will distribute them).
				// To truly test, we add 2 items, then a 3rd that hashes to the same shard.
				// Since we can't easily guess hashes, we just test total capacity limits.
				for i := 0; i < 20; i++ {
					c.Put(fmt.Sprintf("key-%d", i), "val")
				}
			},
			getKey:  "key-0", // likely evicted because 20 items > 16 capacity
			wantVal: "",
			wantOk:  false,
		},
		{
			name:     "Missing key returns false",
			capacity: 160,
			hasher:   StringHasher,
			actions: func(c *ShardedCache[string, string]) {
				c.Put("exists", "yes")
			},
			getKey:  "doesnotexist",
			wantVal: "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewShardedCache[string, string](tt.capacity, tt.hasher)

			if tt.actions != nil {
				tt.actions(c)
			}

			gotVal, gotOk := c.Get(tt.getKey)
			if gotOk != tt.wantOk {
				t.Errorf("Get(%q) ok = %v, want %v", tt.getKey, gotOk, tt.wantOk)
			}
			if gotVal != tt.wantVal {
				t.Errorf("Get(%q) val = %v, want %v", tt.getKey, gotVal, tt.wantVal)
			}
		})
	}
}

// TestShardedCacheConcurrency ensures thread safety.
// Run with: go test -race
func TestShardedCacheConcurrency(t *testing.T) {
	c := NewShardedCache[string, int](160, StringHasher)

	var wg sync.WaitGroup
	workers := 100
	iterations := 100

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				key := fmt.Sprintf("key-%d", (workerID+i)%50)
				c.Put(key, i)
				_, _ = c.Get(key)
			}
		}(w)
	}

	wg.Wait() // If there's a race condition, `go test -race` will fail here
}
