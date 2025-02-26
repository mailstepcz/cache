package cache

import (
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
func (ptr transientPtr[T]) Pointer() *T {
	return ptr.ptr.Pointer()
}
