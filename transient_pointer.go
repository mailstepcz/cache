package cache

import (
	"runtime"
	"sync/atomic"
	"time"
	"weak"
)

type innerTransientPtr[T any] interface {
	Value() *T
}

type innerTransientPtrExp[T any] struct {
	ptr atomic.Pointer[T]
}

func (ptr *innerTransientPtrExp[T]) Value() *T {
	return ptr.ptr.Load()
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
	innerPtr := new(innerTransientPtrExp[T])
	innerPtr.ptr.Store(p)

	time.AfterFunc(exp, func() {
		for _, cb := range cbs {
			cb()
		}
		innerPtr.ptr.Store(nil)
	})

	return transientPtr[T]{
		ptr: innerPtr,
	}
}

// Pointer returns the pointer as an unsafe pointer.
func (ptr transientPtr[T]) Pointer() *T {
	return ptr.ptr.Value()
}
