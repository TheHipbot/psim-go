package psimgo

import (
	"testing"
	"fmt"
	"math/rand"
)

func TestMessagePassing(t *testing.T) {
	comm := PSim{
		P: 5,
		Topology: SWITCH,
	}

	runTest := func (rank int, comm PSim) {
		if rank == 0 {
			fmt.Printf("start test\n")
		}

		if rank == 2 {
			comm.Send(rank, 3, 11)
		} else if rank == 3 {
			c := comm.RecvInt(rank, 2)
			d := c + 1
			if d != 12 {
				t.Errorf("Wrong number received or number not received as int")
			}
		}

		var f int
		a := ToIntArray(comm.All2all_broadcast(rank, rank))

		for _, e := range a {
			f += e
		}

		fmt.Printf("%d\n", f)

		comm.Barrier(rank)

		op := func(a, b interface {}) interface {} {
			if x, ok := a.(int); ok {
				if y, ok := b.(int); ok {
					return x + y
				}
			}
			return 0
		}

		b := comm.All2all_reduce(rank, rank, op)

		fmt.Printf("%d\n", b)

		if f!=10 || f!=b {
			t.Error(fmt.Sprintf("from process %d\n", rank))
		}

		if rank==0 {
			fmt.Printf("test passed\n")
		}
	}

	comm.Run(runTest)
}

func TestMergeSort(t *testing.T) {

	comm := PSim{
		P: 2,
		Topology: SWITCH,
	}

	p_mergesort_test := func(rank int, comm PSim) {
			data := make([]float64, 16, 16)
		if rank == 0 {
			for i := range data {
				data[i] = rand.Float64() * 10
			}
			mid := int(16/2)
			comm.Send(rank, 1, data[mid+1:])
			mergesort(data, 0, mid)
			second := comm.RecvFloatArray(rank, 1)

			for k := range second {
				data[k+mid+1] = second[k]
			}
			fmt.Printf("%v\n", data)
			merge(data, 0, mid+1, 15)
			fmt.Printf("%v\n", data)

			for j := range data {
				if j < len(data)-1 && data[j+1] < data[j] {
					t.Errorf("Data not sorted correctly")
				}
			}
		} else {
			data2 := comm.RecvFloatArray(rank, 0)


			//		reflect.ValueOf(data2)
			mergesort(data2, 0, len(data2)-1)
			comm.Send(rank, 0, data2)
		}
	}

	comm.Run(p_mergesort_test)
}

func mergesort(A []float64, p, r int) {
	if p < r {
		q := int((p + r) / 2)
		mergesort(A, p, q)
		mergesort(A, q+1, r)
		merge(A, p, q+1, r)
	}
}

func merge(A []float64, p, q, r int) {
	B := make([]float64, 0)
	//	fmt.Printf("%d %d %d\n", p, q, r)
	i := p
	j := q
	for {
		if A[i] <= A[j] {
			B = append(B, A[i])
			i = i+1
		} else {
			B = append(B, A[j])
			j = j+1
		}

		if i == q {
			for j<=r {
				B = append(B, A[j])
				j = j+1
			}
			break
		}
		if j == r+1 {
			for i<q {
				B = append(B, A[i])
				i = i+1
			}
			break
		}
	}

	for k := range B {
		A[p+k] = B[k]
	}
}

func mergesort_test() {
	data := make([]float64, 16, 16)
	for i := range data {
		data[i] = rand.Float64() * 10
	}
	fmt.Printf("%v\n", data)
	mergesort(data, 0, 15)
	fmt.Printf("%v\n", data)
}



