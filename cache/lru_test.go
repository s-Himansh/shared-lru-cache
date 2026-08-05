package cache

import "testing"

func TestLRUCache(t *testing.T) {
	// Define the test cases
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
	}

	// Iterate over test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			c := newLRU[string, string](tt.capacity)

			// Execute actions
			if tt.actions != nil {
				tt.actions(c)
			}

			// Check length
			if gotLen := c.Len(); gotLen != tt.wantLen {
				t.Errorf("Len() = %v, want %v", gotLen, tt.wantLen)
			}

			// Check Get result
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
