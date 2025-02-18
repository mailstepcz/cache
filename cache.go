package cache

import (
	"sync"
	"unsafe"
)

type TransientPtr uintptr

func (tp TransientPtr) UnsafePointer() unsafe.Pointer {
	return unsafe.Pointer(tp)
}

type cacheObject[T any] struct {
	ptr TransientPtr
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	lock sync.RWMutex
	data map[K]*cacheObject[V]
	once sync.Once
}

func (c *Cache[K, V]) Put(key K, value *V) {
	c.once.Do(func() {
		c.data = make(map[K]*cacheObject[V])
	})

	c.lock.Lock()
	defer c.lock.Unlock()

	obj := &cacheObject[V]{ptr: TransientPtr(unsafe.Pointer(value))}
	c.data[key] = obj
	// use AddCleanup here
	// runtime.SetFinalizer(value, func(_ *V) {
	// 	c.lock.Lock()
	// 	defer c.lock.Unlock()
	// 	delete(c.data, key)
	// })
}

func (c *Cache[K, V]) PutValue(key K, value V) V {
	c.Put(key, &value)
	return value
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if obj, ok := c.data[key]; ok {
		return (*V)(obj.ptr.UnsafePointer()), true
	}

	return nil, false
}
