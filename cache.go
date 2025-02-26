package cache

import (
	"iter"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// TransientPtr is a transient pointer.
type TransientPtr[T any] uintptr

// Pointer returns the pointer as an unsafe pointer.
func (tp TransientPtr[T]) Pointer() *T {
	return (*T)(unsafe.Pointer(&tp))
}

type cacheObject[T any] struct {
	ptr     TransientPtr[T]
	timer   *time.Timer
	version int
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	data sync.Map
}

// Put puts new value into the cache.
func (c *Cache[K, V]) Put(key K, value *V) *cacheObject[V] {
	var ver int
	if obj, ok := c.data.Load(key); ok {
		obj := obj.(*cacheObject[V])
		if obj.timer != nil {
			obj.timer.Stop()
		}
		ver = obj.version + 1
	}

	obj := &cacheObject[V]{ptr: TransientPtr[V](unsafe.Pointer(value)), version: ver, timer: nil}
	c.data.Store(key, obj)
	runtime.SetFinalizer(value, func(_ *V) {
		c.data.Delete(key)
	})

	return obj
}

// PutExpiring puts new value to the cache. Value will be removed after specified expiration.
// If value is cleaned by GC value can be removed sooner.
func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	obj := c.Put(key, value)
	obj.timer = time.AfterFunc(exp, func() {
		c.data.Delete(key)
	})
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
