/*
Package psimgo is a go implementation of the library PSim for python.
PSim is used to simulate parallel processing on a single machine.

 */
package psimgo

import (
	"math"
	"sync"
	"fmt"
)

// ====== Helper Functions ======

// Go booleans cannot be treated as integers using
// the ^ xor operation. Created a simple function
// to be used in following code
func xor(i, j, v int) bool {
	return (i == v || j == v) && i != j
}

func divmod(i, j int) (d, r int) {
	result := int(math.Floor(float64(i/j)))
	remainder := i - (result*j)
	return result, remainder
}

// ====== Topology Communication functions ======
// boolean functions to test if node i can
// communicate directly with node j. Some
// functions first require input of p (number of nodes)
// and will then return test function

// Bus topology, all nodes can communicate
// with all others. Always true
func BUS(i, j int) bool {
	return true
}

// Switch topology, all nodes can communicate
// with all others. Always true
func SWITCH(i,j int) bool {
	return true
}

// 1-D Mesh topology. Returns function to test.
// if i and j are neighbors, communication is available
func MESH1(p int) func(i, j int) bool {
	return func(i, j int) bool {
		return int(math.Abs(float64(i - j))) == 1
	}
}

// 1-D Torus topology. Like 1-D mesh, but circularly
// linked. Function must account for first and last
// nodes connected
func TORUS1(p int) func(i, j int) bool {
	return func(i, j int) bool {
		return (i - j + p) % p == 1 || (j - i + p) % p == 1
	}
}

//2-D Mesh.
func MESH2(p int) func(i, j int) bool {
	q := int(math.Floor(math.Sqrt(float64(p)) + 0.1))
	return func(i, j int) bool {
		return xor(int(math.Abs(float64(i%q-j%q))), int(math.Abs(float64(i/q-j/q))), 1)
	}
}

func TORUS2(p int) func(i, j int) bool {
	q := int(math.Floor(math.Sqrt(float64(p)) + 0.1))
	return func(i, j int) bool {
		return xor(((i%q - j%q + q) %q), ((i/q - j/q + q) %q), 1) || xor(((j%q - i%q + q) %q), ((j/q - i/q + q) %q), 1)
	}
}

func TREE(i, j int) bool {
	return i == int(math.Ceil(float64(j-1) /2.0)) || j == int(math.Ceil(float64(i-1) /2.0))
}

// ====== PSim type and Functions ======

type PSim struct {
	P int
	Topology func(i, j int) (bool)
	initialized bool
	Pipes [][]chan interface {}
}

func (psim PSim) Run(f func(rank int, comm PSim)) {
	var wg sync.WaitGroup

	// init arrays
	psim.Pipes = make([][]chan interface {}, psim.P)

	// default p num procs
	if psim.P == 0 {
		psim.P = 1
	}

	// default topology
	if psim.Topology == nil {
		psim.Topology = SWITCH
	}

	// initialize channels
	for i := 0; i < psim.P; i++ {
		psim.Pipes[i] = make([]chan interface {}, psim.P)
		for j := 0; j < psim.P; j++ {
			psim.Pipes[i][j] = make(chan interface {})
		}
	}

	wg.Add(psim.P)

	// start go threads
	for r := 0; r < psim.P; r++ {
		go func (r int) {
			defer wg.Done()
			f(r, psim)
		}(r)
	}

	// psim object has been initialized
	psim.initialized = true

	wg.Wait()
}

func (psim PSim) Send(source, dest int, data interface {}) {
	// if i or j less than 0 or greater then nprocs error
	// if i -> j communication unavailable, error
	if source < 0 || source > psim.P-1 || dest < 0 || dest > psim.P-1 || !psim.Topology(source, dest) {
		fmt.Printf("Send ERR:\nOut of range, i: %d; j: %d\n", source, dest);
	} else {
		// send data
		psim.Pipes[source][dest] <-data
	}
}

func (psim PSim) Recv(rank, source int) interface {} {
	// if i or j less than 0 or greater then nprocs error
	// if i -> j communication unavailable, error
	if rank < 0 || rank > psim.P-1 || source < 0 || source > psim.P-1 || !psim.Topology(rank, source) {
		fmt.Printf("Recv ERR:\nOut of range, i: %d; j: %d\n", rank, source);
	} else {
		// recv data
		return <-psim.Pipes[source][rank]
	}
	return nil
}

func (psim PSim) One2all_broadcast(rank, source int, data interface {}) interface {} {
	if rank == source {
		for i := 0; i < psim.P; i++ {
			if i != source {
				psim.Send(source, i, data)
			}
		}
		return data
	} else {
		return psim.Recv(rank, source)
	}
}

func (psim PSim) All2all_broadcast(rank int, data interface {}) []interface {} {
	vector := psim.All2one_collect(rank, 0, data)
	v := psim.One2all_broadcast(rank, 0, vector)
	switch t:= v.(type){
	case []interface {}:
		return t
	case interface {}:
		return []interface {} {t}
	default:
		// TODO error
	}
	return nil
}

func (psim PSim) One2all_scatter(rank, source int, data []interface {}) []interface {} {
	if rank == source {
		h, reminder := divmod(len(data), psim.P)

		if reminder > 0 {
			h += 1
		}

		for i := 1; i < psim.P; i++ {
			psim.Send(rank, i, data[i*h:i*h+h])
		}
		return data[0:h]
	} else {
		v := psim.Recv(rank, source)
		if vector, ok := v.([]interface {}); ok {
			return vector	
		} else {
			return nil
		}
	}
}

func (psim PSim) All2one_collect(rank, dest int, data interface {}) []interface {} {
	var result []interface {}
	
	if rank == dest {
		for i := 0; i < psim.P; i++ {
			if i == rank {
				result = append(result, data)
			} else {
				result = append(result, psim.Recv(rank, i))
			}
		}
	} else {
		psim.Send(rank, dest, data)
	}
	
	return result
}

func (psim PSim) All2one_reduce(
	rank, dest int,
	data interface {},
	op func(a, b interface {}) interface {}) interface {} {
	if rank == dest {
		result := data
		for i := 1; i< psim.P; i++ {
			if i != rank {
				result = op(result, psim.Recv(rank, i))
			}
		}
		return result
	} else {
		psim.Send(rank, dest, data)
		return 0
	}
}

func (psim PSim) All2all_reduce(rank int, data interface {}, op func(a, b interface {}) interface {}) interface {} {
	result := psim.All2one_reduce(rank, 0, data, op)
	result = psim.One2all_broadcast(rank, 0, result)
	return result
}

func (psim PSim) Barrier(rank int) {
	psim.All2all_broadcast(rank, 0)
}
