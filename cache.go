package cache

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// TransientPtr is a transient pointer.
type TransientPtr[T any] struct {
	ptr unsafe.Pointer
}

// UnsafePointer returns the pointer as an unsafe pointer.
func (tp TransientPtr[T]) Pointer() *T {
	return (*T)(tp.ptr)
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

func (c *Cache[K, V]) Put(key K, value *V) *cacheObject[V] {
	var ver int
	if obj, ok := c.data.Load(key); ok {
		obj := obj.(*cacheObject[V])
		if obj.timer != nil {
			obj.timer.Stop()
		}
		ver = obj.version + 1
	}

	obj := &cacheObject[V]{ptr: TransientPtr[V]{ptr: unsafe.Pointer(value)}, version: ver, timer: nil}
	c.data.Store(key, obj)

	fmt.Println(1)

	x, y := c.data.Load(key)
	fmt.Println("x", x, "y", y)
	runtime.SetFinalizer(value, func(_ *V) {
		/* use 'add cleanup' here once available
		go func() {
			c.data.Delete(key)
		}()*/
	})

	return obj
}

func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	fmt.Println("putting new expiring", key, value, exp)
	obj := c.Put(key, value)
	obj.timer = time.AfterFunc(exp, func() {
		c.data.Delete(key)
	})
	fmt.Println("print after put before get")
	val, found := c.Get(key)
	fmt.Println("print after put", val, found)
}

func (c *Cache[K, V]) PutValue(key K, value V) V {
	c.Put(key, &value)
	return value
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	// c.data.Range(func(key, value any) bool {

	// 	fmt.Println("iter", key, value.(*cacheObject[V]))

	// 	return true
	// })

	if obj, ok := c.data.Load(key); ok {
		fmt.Println("1-1")
		obj, ok := obj.(*cacheObject[V])
		fmt.Println("1-2", ok)
		return obj.ptr.Pointer(), true
	}

	fmt.Println("1-3")

	return nil, false
}

func (c *Cache[K, V]) GetAll(cb func(K, *V)) {
	c.data.Range(func(key, value any) bool {
		obj := value.(*cacheObject[V])

		cb(key.(K), obj.ptr.Pointer())
		return true
	})
}
