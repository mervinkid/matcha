package parallel_test

import (
	"github.com/mervinkid/allspark/parallel"
	"testing"
)

func TestNewGoroutine(t *testing.T) {

	parallelism := 100

	goroutines := make([]parallel.Goroutine, parallelism)

	for i := 0; i < 100; i ++ {
		in := i
		goroutine := parallel.NewGoroutine(func() {
			gId, err := parallel.GetGoroutineId()
			if err != nil {
				gId = 0
			}
			t.Log("Goroutine ", gId, ":", in)
		})
		goroutines[i] = goroutine
	}

	for _, g := range goroutines {
		g.Start()
	}

	for _, g := range goroutines {
		g.Join()
	}
}
