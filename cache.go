package cache

import (
	"iter"
	"sync"
	"sync/atomic"
	"time"
)

type cacheObject[T any] struct {
	ptr       transientPtr[T]
	timestamp atomic.Int64
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	data sync.Map
}

// Put puts new value into the cache.
func (c *Cache[K, V]) Put(key K, value *V) *cacheObject[V] {
	timestamp := time.Now().UnixNano()

	obj := &cacheObject[V]{ptr: makeTransientPtrGC(value, func() {
		if obj, ok := c.data.Load(key); ok {
			obj := obj.(*cacheObject[V])
			if timestamp == obj.timestamp.Load() {
				c.data.Delete(key)
			}
		}
	})}
	obj.timestamp.Store(timestamp)

	c.data.Store(key, obj)

	return obj
}

// PutExpiring puts new value to the cache. Value will be removed after specified expiration.
func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	timestamp := time.Now().UnixNano()

	obj := &cacheObject[V]{ptr: makeTransientPtrExpiring(value, exp, func() {
		if obj, ok := c.data.Load(key); ok {
			obj := obj.(*cacheObject[V])
			if timestamp == obj.timestamp.Load() {
				c.data.Delete(key)
			}
		}

	})}
	obj.timestamp.Store(timestamp)

	c.data.Store(key, obj)
}

// PutValue inserts new value to the cache. And returns inserted value.
func (c *Cache[K, V]) PutValue(key K, value V) V {
	c.Put(key, &value)
	return value
}

// Get returns value in cache and flag which indicates that value is available.
func (c *Cache[K, V]) Get(key K) (*V, bool) {
	if obj, ok := c.data.Load(key); ok {
		obj := obj.(*cacheObject[V])
		return obj.ptr.Pointer(), true
	}

	return nil, false
}

// GetAll returns all values stored in cache.
func (c *Cache[K, V]) GetAll() iter.Seq2[K, *V] {
	return func(yield func(K, *V) bool) {
		c.data.Range(func(key, value any) bool {
			return yield(key.(K), value.(*cacheObject[V]).ptr.Pointer())
		})
	}

}
