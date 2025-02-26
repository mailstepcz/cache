package cache

import (
	"iter"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

type innerTransientPtr[T any] interface {
	Pointer() *T
}

type innerTransientPtrGC[T any] struct {
	mtx sync.RWMutex
	ptr uintptr
}

func (ptr *innerTransientPtrGC[T]) Pointer() *T {
	ptr.mtx.RLock()
	defer ptr.mtx.RUnlock()
	return *(**T)(unsafe.Pointer(&ptr.ptr))
}

type innerTransientPtrExp[T any] struct {
	mtx sync.RWMutex
	ptr *T
}

func (ptr *innerTransientPtrExp[T]) Pointer() *T {
	ptr.mtx.RLock()
	defer ptr.mtx.RUnlock()
	return ptr.ptr
}

// transientPtr is a transient pointer.
type transientPtr[T any] struct {
	ptr innerTransientPtr[T]
}

type deleteCallback = func()

func makeTransientPtr[T any](p *T, cbs ...deleteCallback) transientPtr[T] {
	innerPtr := &innerTransientPtrGC[T]{
		ptr: uintptr(unsafe.Pointer(p)),
	}

	runtime.AddCleanup(p, func(ptr *innerTransientPtrGC[T]) {
		for _, cb := range cbs {
			cb()
		}
		ptr.mtx.Lock()
		defer ptr.mtx.Unlock()
		ptr.ptr = 0
	}, innerPtr)

	return transientPtr[T]{
		ptr: innerPtr,
	}

}

func makeTransientPtrExpiring[T any](p *T, exp time.Duration, cbs ...deleteCallback) transientPtr[T] {
	innerPtr := &innerTransientPtrExp[T]{
		ptr: p,
	}

	time.AfterFunc(exp, func() {
		for _, cb := range cbs {
			cb()
		}
		innerPtr.mtx.Lock()
		defer innerPtr.mtx.Unlock()
		innerPtr.ptr = nil
	})

	return transientPtr[T]{
		ptr: innerPtr,
	}
}

// Pointer returns the pointer as an unsafe pointer.
func (tp transientPtr[T]) Pointer() *T {
	return tp.ptr.Pointer()
}

type cacheObject[T any] struct {
	ptr       transientPtr[T]
	timestamp time.Time
}

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	data sync.Map
}

// Put puts new value into the cache.
func (c *Cache[K, V]) Put(key K, value *V) *cacheObject[V] {
	timestamp := time.Now()

	obj := &cacheObject[V]{ptr: makeTransientPtr(value, func() {
		if obj, ok := c.data.Load(key); ok {
			obj := obj.(*cacheObject[V])
			if timestamp.Equal(obj.timestamp) {
				c.data.Delete(key)
			}
		}
	}), timestamp: timestamp}

	c.data.Store(key, obj)

	return obj
}

// PutExpiring puts new value to the cache. Value will be removed after specified expiration.
func (c *Cache[K, V]) PutExpiring(key K, value *V, exp time.Duration) {
	timestamp := time.Now()

	obj := &cacheObject[V]{ptr: makeTransientPtrExpiring(value, exp, func() {
		if obj, ok := c.data.Load(key); ok {
			obj := obj.(*cacheObject[V])
			if timestamp.Equal(obj.timestamp) {
				c.data.Delete(key)
			}
		}

	}), timestamp: timestamp}

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
