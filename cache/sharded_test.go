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
			name:     "Eviction across shards",
			capacity: 16,
			hasher:   StringHasher,
			actions: func(c *ShardedCache[string, string]) {
				for i := 0; i < 20; i++ {
					c.Put(fmt.Sprintf("key-%d", i), "val")
				}
			},
			getKey:  "key-0",
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

func TestShardedCacheLen(t *testing.T) {
	c := NewShardedCache[string, string](160, StringHasher)
	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}
	c.Put("a", "1")
	c.Put("b", "2")
	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2", c.Len())
	}
}

func TestShardedCacheDelete(t *testing.T) {
	c := NewShardedCache[string, string](160, StringHasher)
	c.Put("a", "1")
	if !c.Delete("a") {
		t.Error("Delete(a) = false, want true")
	}
	if c.Delete("a") {
		t.Error("Delete(a) on missing = true, want false")
	}
	_, ok := c.Get("a")
	if ok {
		t.Error("Get(a) after delete returned ok=true")
	}
}

func TestShardedCacheClear(t *testing.T) {
	c := NewShardedCache[string, string](160, StringHasher)
	c.Put("a", "1")
	c.Put("b", "2")
	c.Clear()
	if c.Len() != 0 {
		t.Errorf("Len() after Clear = %d, want 0", c.Len())
	}
}

func TestShardedCacheKeys(t *testing.T) {
	c := NewShardedCache[string, string](160, StringHasher)
	c.Put("a", "1")
	c.Put("b", "2")
	c.Put("c", "3")

	keys := c.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}
}

func TestShardedCacheCustomShards(t *testing.T) {
	c := NewShardedCache[string, string](100, StringHasher, WithNumShards[string, string](4))
	c.Put("a", "1")
	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
}

func TestIntHasherDistribution(t *testing.T) {
	shardCounts := make(map[uint32]int)
	for i := 0; i < 10000; i++ {
		shard := IntHasher(i) % 16
		shardCounts[shard]++
	}
	for shard, count := range shardCounts {
		if count < 100 || count > 1200 {
			t.Errorf("IntHasher shard %d has %d items, expected even distribution ~625", shard, count)
		}
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

	wg.Wait()
}

func BenchmarkShardedPut(b *testing.B) {
	c := NewShardedCache[string, string](1000, StringHasher)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Put(fmt.Sprintf("key-%d", i%1000), "value")
	}
}

func BenchmarkShardedGet(b *testing.B) {
	c := NewShardedCache[string, string](1000, StringHasher)
	for i := 0; i < 1000; i++ {
		c.Put(fmt.Sprintf("key-%d", i), "value")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key-%d", i%1000))
	}
}

func BenchmarkShardedConcurrent(b *testing.B) {
	c := NewShardedCache[string, int](10000, StringHasher)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%5000)
			c.Put(key, i)
			c.Get(key)
			i++
		}
	})
}
