package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"runtime"
	"sync"
	"weak"
)

var (
	// ErrCacheMiss indicates a cache miss.
	ErrCacheMiss = errors.New("cache miss")
)

// Cache is a cache for transient values.
type Cache[K comparable, V any] struct {
	data sync.Map
}

func (c *Cache[K, V]) Put(key K, value *V) {
	c.data.Store(key, weak.Make(value))
	runtime.AddCleanup(value, func(key K) {
		c.data.Delete(key)
	}, key)
}

func (c *Cache[K, V]) Get(key K) (*V, bool) {
	if ptr, ok := c.data.Load(key); ok {
		if value := ptr.(weak.Pointer[V]).Value(); value != nil {
			return value, true
		}
	}
	return nil, false
}

type cachedObject struct {
	data []byte
}

// ObjectCache is a cache for transient objects.
type ObjectCache[K comparable, V any] struct {
	data Cache[K, cachedObject]
}

func (c *ObjectCache[K, V]) Put(key K, object *V) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(object); err != nil {
		return err
	}
	c.data.Put(key, &cachedObject{buf.Bytes()})
	return nil
}

func (c *ObjectCache[K, V]) Get(key K) (*V, error) {
	if co, ok := c.data.Get(key); ok {
		var obj V
		if err := gob.NewDecoder(bytes.NewReader(co.data)).Decode(&obj); err != nil {
			return nil, err
		}
		return &obj, nil
	}
	return nil, ErrCacheMiss
}
