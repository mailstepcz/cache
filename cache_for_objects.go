package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
)

var (
	// ErrCacheMiss indicates a cache miss.
	ErrCacheMiss = errors.New("cache miss")
)

type ocachedObject struct {
	data []byte
}

// ObjectCache is a cache for transient objects.
type ObjectCache[K comparable, V any] struct {
	data Cache[K, ocachedObject]
}

func (c *ObjectCache[K, V]) Put(key K, object *V) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(object); err != nil {
		return err
	}
	c.data.Put(key, &ocachedObject{buf.Bytes()})
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
