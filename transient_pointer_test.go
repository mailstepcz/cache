package cache

import (
	"runtime"
	"testing"
	"time"

	"github.com/mailstepcz/testutils/testcond"
	"github.com/stretchr/testify/require"
)

func TestTransientPtr(t *testing.T) {
	req := require.New(t)
	p := &person{"Oisin", 18}

	tp := makeTransientPtrGC(p)

	testcond.Equal(t, "Oisin", tp.Pointer().Name)

	runtime.GC()

	req.Nil(tp.Pointer())
}

func TestTransientPtr2(t *testing.T) {
	req := require.New(t)

	iters := 1_000_000

	var tps []transientPtr[person]
	for i := 1; i <= iters; i++ {
		p := &person{"Oisin", 18}
		tp := makeTransientPtrGC(p)
		tps = append(tps, tp)

		if i%1_000 == 0 {
			runtime.GC()
		}

	}

	var nils int

	for _, x := range tps {
		if x.Pointer() == nil {
			nils++
		}
	}

	req.Equal(iters, nils)
}

func TestTransientPtrExpiring(t *testing.T) {
	req := require.New(t)
	p := &person{"Oisin", 18}

	tp := makeTransientPtrExpiring(p, time.Second)

	testcond.Equal(t, "Oisin", tp.Pointer().Name)

	runtime.GC()

	req.NotNil(tp.Pointer())

	time.Sleep(time.Second)

	runtime.GC()

	req.Nil(tp.Pointer())
}
