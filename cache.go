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
	lock sync.RWMutex
	data map[K]*cacheObject[V]
	once sync.Once
}

func (c *Cache[K, V]) Put(key K, value *V) *cacheObject[V] {
	c.once.Do(func() {
		c.data = make(map[K]*cacheObject[V])
	})

	c.lock.Lock()
	defer c.lock.Unlock()

	var ver int
	if obj, ok := c.data[key]; ok {
		obj.timer.Stop()
		ver = obj.version + 1
	}

	obj := &cacheObject[V]{ptr: TransientPtr(unsafe.Pointer(value)), version: ver}
	c.data[key] = obj
	runtime.SetFinalizer(value, func(_ *V) {
		/* use 'add cleanup' here once available // go func() {
			c.lock.Lock()
			defer c.lock.Unlock()
			delete(c.data, key)
		}()*/
	})

	return obj
}

func (c *Cache[K, V]) PutValue(key K, value V) V {
	c.Put(key, &value)
	return value
}

func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	obj := c.Put(key, value)
	obj.timer = time.AfterFunc(exp, func() {
		c.lock.Lock()
		defer c.lock.Unlock()
		delete(c.data, key)
	})
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if obj, ok := c.data[key]; ok {
		return (*V)(obj.ptr.UnsafePointer()), true
	}

	return nil, false
}
