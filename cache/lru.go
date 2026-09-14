package cache

import (
	"container/list"
)

// lruCache is a non-thread-safe LRU cache. It is wrapped by ShardedCache for concurrency.
type lruCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*list.Element
	ll       *list.List
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func newLRU[K comparable, V any](capacity int) *lruCache[K, V] {
	return &lruCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*list.Element),
		ll:       list.New(),
	}
}

func (c *lruCache[K, V]) Get(key K) (V, bool) {
	if el, ok := c.cache[key]; ok {
		c.ll.MoveToFront(el)
		return el.Value.(*entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (c *lruCache[K, V]) Put(key K, value V) {
	if el, ok := c.cache[key]; ok {
		c.ll.MoveToFront(el)
		el.Value.(*entry[K, V]).value = value
		return
	}

	el := c.ll.PushFront(&entry[K, V]{key, value})
	c.cache[key] = el

	if c.ll.Len() > c.capacity {
		oldest := c.ll.Back()
		if oldest != nil {
			c.ll.Remove(oldest)
			delete(c.cache, oldest.Value.(*entry[K, V]).key)
		}
	}
}

func (c *lruCache[K, V]) Delete(key K) bool {
	if el, ok := c.cache[key]; ok {
		c.ll.Remove(el)
		delete(c.cache, key)
		return true
	}
	return false
}

func (c *lruCache[K, V]) Clear() {
	c.cache = make(map[K]*list.Element)
	c.ll.Init()
}

func (c *lruCache[K, V]) Len() int {
	return c.ll.Len()
}

func (c *lruCache[K, V]) Keys() []K {
	keys := make([]K, 0, c.ll.Len())
	for el := c.ll.Front(); el != nil; el = el.Next() {
		keys = append(keys, el.Value.(*entry[K, V]).key)
	}
	return keys
}
