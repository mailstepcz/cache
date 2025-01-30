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
	data sync.Map
}

func (c *Cache[K, V]) Put(key K, value *V) {
	c.data.Store(key, weak.Make(value))
	runtime.AddCleanup(value, func(key K) {
		c.data.Delete(key)
	}, key)
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	if ptr, ok := c.data.Load(key); ok {
		if value := ptr.(weak.Pointer[V]).Value(); value != nil {
			return value, true
		}
	}
	return nil, false
}
