package cache

import (
	"runtime"
	"sync"
	"time"
	"weak"
)

type innerTransientPtr[T any] interface {
	Value() *T
}

type innerTransientPtrExp[T any] struct {
	mtx sync.RWMutex
	ptr *T
}

func (ptr *innerTransientPtrExp[T]) Value() *T {
	ptr.mtx.RLock()
	defer ptr.mtx.RUnlock()
	return ptr.ptr
}

// transientPtr is a transient pointer.
type transientPtr[T any] struct {
	ptr innerTransientPtr[T]
}

type deleteCallback = func()

func makeTransientPtrGC[T any](p *T, cbs ...deleteCallback) transientPtr[T] {
	runtime.AddCleanup(p, func(_ int) {
		for _, cb := range cbs {
			cb()
		}

	}, 0)

	return transientPtr[T]{
		ptr: weak.Make(p),
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
	return ptr.ptr.Value()
}
