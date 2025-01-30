package cache

import (
	"errors"
	"runtime"
	"sync"
	"weak"
)

var (
	// ErrCacheMiss indicates a cache miss.
	ErrCacheMiss = errors.New("cache miss")
)

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	lock sync.RWMutex
	data map[K]weak.Pointer[V]
}

func (c *Cache[K, V]) Put(key K, value *V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.data == nil {
		c.data = make(map[K]weak.Pointer[V])
	}
	c.data[key] = weak.Make(value)
	runtime.AddCleanup(value, func(key K) {
		c.lock.Lock()
		defer c.lock.Unlock()
		delete(c.data, key)
	}, key)
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if ptr, ok := c.data[key]; ok {
		if value := ptr.Value(); value != nil {
			return value, true
		}
	}
	return nil, false
}
