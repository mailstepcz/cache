package cache

import (
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// TransientPtr is a transient pointer.
type TransientPtr uintptr

// UnsafePointer returns the pointer as an unsafe pointer.
func (tp TransientPtr) UnsafePointer() unsafe.Pointer {
	return unsafe.Pointer(tp)
}

type cacheObject[T any] struct {
	ptr TransientPtr
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	data sync.Map
}

func (c *Cache[K, V]) Put(key K, value *V) {
	obj := &cacheObject[V]{ptr: TransientPtr(unsafe.Pointer(value))}
	c.data.Store(key, obj)
	runtime.SetFinalizer(value, func(_ *V) {
		// use 'add cleanup' here once available
		go func() {
			c.data.Delete(key)
		}()
	})
}

func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	c.Put(key, value)
	time.AfterFunc(exp, func() {
		c.data.Delete(key)
	})
}

func (c *Cache[K, V]) PutValue(key K, value V) V {
	c.Put(key, &value)
	return value
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	if obj, ok := c.data.Load(key); ok {
		obj := obj.(*cacheObject[V])
		return (*V)(obj.ptr.UnsafePointer()), true
	}

	return nil, false
}
