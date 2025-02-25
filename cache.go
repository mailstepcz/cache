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
	ptr     TransientPtr
	timer   *time.Timer
	version int
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	data sync.Map
}

func (c *Cache[K, V]) Put(key K, value *V) {
	var ver int
  if obj, ok := c.data.Load(key); ok {
    obj := obj.(*cacheObject[V])
		obj.timer.Stop()
		ver = obj.version + 1
	}

	obj := &cacheObject[V]{ptr: TransientPtr(unsafe.Pointer(value)), version: ver}
	c.data.Store(key, obj)
	runtime.SetFinalizer(value, func(_ *V) {
		/* use 'add cleanup' here once available
		go func() {
			c.data.Delete(key)
		}()*/
	})

	return obj
}

func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	obj := c.Put(key, value)
	obj.timer = time.AfterFunc(exp, func() {
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
