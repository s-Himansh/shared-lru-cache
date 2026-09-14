package cache

import "testing"

func TestLRUCache(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		actions  func(c *lruCache[string, string])
		getKey   string
		wantVal  string
		wantOk   bool
		wantLen  int
	}{
		{
			name:     "Put and Get single item",
			capacity: 2,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
			},
			getKey:  "a",
			wantVal: "1",
			wantOk:  true,
			wantLen: 1,
		},
		{
			name:     "Evict Least Recently Used",
			capacity: 2,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("b", "2")
				c.Put("c", "3") // "a" should be evicted
			},
			getKey:  "a",
			wantVal: "",
			wantOk:  false,
			wantLen: 2,
		},
		{
			name:     "Get updates recency (prevents eviction)",
			capacity: 2,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("b", "2")
				c.Get("a")      // "a" is now most recently used
				c.Put("c", "3") // "b" should be evicted, not "a"
			},
			getKey:  "a",
			wantVal: "1",
			wantOk:  true,
			wantLen: 2,
		},
		{
			name:     "Update existing key does not increase length",
			capacity: 2,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("a", "updated")
			},
			getKey:  "a",
			wantVal: "updated",
			wantOk:  true,
			wantLen: 1,
		},
		{
			name:     "Get missing key",
			capacity: 2,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
			},
			getKey:  "b",
			wantVal: "",
			wantOk:  false,
			wantLen: 1,
		},
		{
			name:     "Delete existing key",
			capacity: 3,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("b", "2")
				c.Delete("a")
			},
			getKey:  "a",
			wantVal: "",
			wantOk:  false,
			wantLen: 1,
		},
		{
			name:     "Delete missing key returns false",
			capacity: 3,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Delete("z")
			},
			getKey:  "a",
			wantVal: "1",
			wantOk:  true,
			wantLen: 1,
		},
		{
			name:     "Clear removes all items",
			capacity: 3,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("b", "2")
				c.Clear()
			},
			getKey:  "a",
			wantVal: "",
			wantOk:  false,
			wantLen: 0,
		},
		{
			name:     "Keys returns all keys",
			capacity: 3,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("b", "2")
				c.Put("c", "3")
			},
			wantLen: 3,
		},
		{
			name:     "Capacity 1",
			capacity: 1,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("b", "2") // "a" evicted
			},
			getKey:  "a",
			wantVal: "",
			wantOk:  false,
			wantLen: 1,
		},
		{
			name:     "Capacity 0 evicts everything",
			capacity: 0,
			actions: func(c *lruCache[string, string]) {
				c.Put("a", "1")
				c.Put("b", "2")
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newLRU[string, string](tt.capacity)

			if tt.actions != nil {
				tt.actions(c)
			}

			if gotLen := c.Len(); gotLen != tt.wantLen {
				t.Errorf("Len() = %v, want %v", gotLen, tt.wantLen)
			}

			if tt.getKey != "" {
				gotVal, gotOk := c.Get(tt.getKey)
				if gotOk != tt.wantOk {
					t.Errorf("Get(%q) ok = %v, want %v", tt.getKey, gotOk, tt.wantOk)
				}
				if gotVal != tt.wantVal {
					t.Errorf("Get(%q) val = %v, want %v", tt.getKey, gotVal, tt.wantVal)
				}
			}
		})
	}
}

func TestKeys(t *testing.T) {
	c := newLRU[string, string](5)
	c.Put("a", "1")
	c.Put("b", "2")
	c.Put("c", "3")

	keys := c.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for _, expected := range []string{"a", "b", "c"} {
		if !keySet[expected] {
			t.Errorf("Keys() missing key %q", expected)
		}
	}
}

func BenchmarkLRUPut(b *testing.B) {
	c := newLRU[string, string](1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Put("key", "value")
	}
}

func BenchmarkLRUGetMiss(b *testing.B) {
	c := newLRU[string, string](1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("nonexistent")
	}
}

func BenchmarkLRUGetHit(b *testing.B) {
	c := newLRU[string, string](1000)
	c.Put("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}
