package psimgo

import (
	"testing"
	"fmt"
)

func acceptanceTest(t *testing.T) {
	comm := PSim{
		p: 5,
		topology: SWITCH,
	}
	runTest := func (rank int, comm PSim) {
		if rank == 0 {
			fmt.Printf("start test\n")
		}

		var f int
		a := comm.all2all_broadcast(rank, 0)
		for _, e := range a {
			if v, ok := e.(int); ok {
				f += v
			}
		}

		comm.barrier()

		op := func(a, b interface {}) interface {} {
			if x, ok := a.(int); ok {
				if y, ok := b.(int); ok {
					return x + y
				}
			}
			return 0
		}

		b := comm.all2all_reduce(rank, 0, op);

		if f!=10 || f!=b {
			fmt.Printf(rank, fmt.Sprintf("from process %d\n", rank))
		}

		if rank==0 {
			fmt.Printf(rank, "test passed\n")
		}
	}
	comm.run(runTest)
}


