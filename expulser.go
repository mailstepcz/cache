package cache

import "runtime"

type expulser[T any] struct{}

func (ex *expulser[T]) add(obj *T, f func()) func() {
	cleanup := runtime.AddCleanup(obj, func(f func()) {
		f()
	}, f)
	return func() {
		cleanup.Stop()
	}
}
