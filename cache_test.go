package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectCache(t *testing.T) {
	req := require.New(t)

	type person struct {
		Name string
		Age  int
	}

	var c ObjectCache[string, person]

	err := c.Put("p1", &person{"Oisin", 18})
	req.NoError(err)

	p, err := c.Get("p1")
	req.NoError(err)
	req.Equal("Oisin", p.Name)
	req.Equal(18, p.Age)
}
