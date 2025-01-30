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

type cacheObject[T any] struct {
	ptr     weak.Pointer[T]
	cleanup runtime.Cleanup
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	lock sync.RWMutex
	data map[K]*cacheObject[V]
}

func (c *Cache[K, V]) Put(key K, value *V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.data == nil {
		c.data = make(map[K]*cacheObject[V])
	}

	if obj, ok := c.data[key]; ok {
		obj.cleanup.Stop()
	}

	obj := &cacheObject[V]{ptr: weak.Make(value)}
	c.data[key] = obj
	obj.cleanup = runtime.AddCleanup(value, func(key K) {
		c.lock.Lock()
		defer c.lock.Unlock()
		delete(c.data, key)
	}, key)
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if obj, ok := c.data[key]; ok {
		if value := obj.ptr.Value(); value != nil {
			return value, true
		}
	}
	return nil, false
}
