package cache

import (
	"runtime"
	"testing"

	"github.com/mailstepcz/testutils/testcond"
)

func TestExpulser(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	var ex expulser[person]

	var expelled bool
	ex.add(&person{"Saoirse", 25}, func() {
		expelled = true
	})
	runtime.GC()
	testcond.Equal(t, true, expelled)
}
