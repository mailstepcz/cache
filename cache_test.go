package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type person struct {
	Name string
	Age  int
}

func TestObjectCache(t *testing.T) {
	req := require.New(t)

	var c ObjectCache[string, person]

	err := c.Put("p1", &person{"Oisin", 18})
	req.NoError(err)

	p, err := c.Get("p1")
	req.NoError(err)
	req.Equal("Oisin", p.Name)
	req.Equal(18, p.Age)
}

func TestCache(t *testing.T) {
	req := require.New(t)

	var c Cache[string, person]
	val := c.Put("p1", &person{"Oisin", 18})
	req.NotNil(val)

	p, found := c.Get("p1")
	req.True(found)
	req.Equal("Oisin", p.Name)
	req.Equal(18, p.Age)
}
