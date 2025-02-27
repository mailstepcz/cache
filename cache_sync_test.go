package cache

import (
	"fmt"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"
)

func TestCacheConcurrent(t *testing.T) {
	synctest.Run(func() {
		req := require.New(t)

		var c Cache[string, person]

		for i := 0; i < 1_000; i++ {
			k := fmt.Sprintf("k%d", i)
			v := fmt.Sprintf("v%d", i)
			go func() {
				c.Put(k, &person{v, i})
			}()
		}

		synctest.Wait()

		for i := 0; i < 1_000; i++ {
			k := fmt.Sprintf("k%d", i)
			p, ok := c.Get(k)
			req.True(ok)
			req.Equal(fmt.Sprintf("v%d", i), p.Name)
			req.Equal(i, p.Age)
		}
	})
}
