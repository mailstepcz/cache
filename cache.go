package cache

import (
	"runtime"
	"sync"
	"unsafe"
)

type cacheObject[T any] struct {
	ptr uintptr
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

	obj := &cacheObject[V]{ptr: uintptr(unsafe.Pointer(value))}
	c.data[key] = obj
	runtime.SetFinalizer(value, func(_ *V) {
		c.lock.Lock()
		defer c.lock.Unlock()
		delete(c.data, key)
	})
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if obj, ok := c.data[key]; ok {
		return (*V)(unsafe.Pointer(obj.ptr)), true
	}
	return nil, false
}
