package cache

import (
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
	cleanup func()
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	data     sync.Map
	expulser expulser[cacheObject[V]]
}

func (c *Cache[K, V]) Put(key K, value *V) *cacheObject[V] {
	var ver int
	if obj, ok := c.data.Load(key); ok {
		obj := obj.(*cacheObject[V])
		obj.timer.Stop()
		obj.cleanup()
		ver = obj.version + 1
	}

	obj := &cacheObject[V]{ptr: TransientPtr(unsafe.Pointer(value)), version: ver}
	c.data.Store(key, obj)
	obj.cleanup = c.expulser.add(obj, func() {
		if obj, ok := c.data.Load(key); ok {
			obj := obj.(*cacheObject[V])
			if ver != obj.version {
				if obj.timer != nil {
					obj.timer.Stop()
				}
				return
			}
		}
		c.data.Delete(key)
	})

	return obj
}

func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	obj := c.Put(key, value)
	obj.timer = time.AfterFunc(exp, func() {
		c.data.Delete(key)
		obj.cleanup()
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
