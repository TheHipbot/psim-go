/*
Package psimgo is a go implementation of the library PSim for python.
PSim is used to simulate parallel processing on a single machine.


 */
package psimgo

import (
	"math"
)

// ====== Helper Functions ======

// Go booleans cannot be treated as integers using
// the ^ xor operation. Created a simple function
// to be used in following code
func xor(i, j, v int) bool {
	return (i == v || j == v) && i != j
}

func divmod(i, j int) (d, r int) {
	result := int(math.Floor(i/j))
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
	p int
	topology func(i, j int) bool
	initialized bool
	pipes [][]chan interface {}
}

func (psim PSim) run(f func(rank int, comm PSim)) {
	// default p num procs
	if psim.p == 0 {
		psim.p = 1
	}

	// default topology
	if psim.topology == nil {
		psim.toplogy = SWITCH
	}

	// initialize channels
	for i := range psim.p {
		for j := range psim.p {
			psim.pipes[i][j] = make(chan interface {})
		}
	}

	// start go threads
	for r := range psim.p {
		go f(r, psim)
	}

	// psim object has been initialized
	psim.initialized = true
}

func (psim *PSim) send(i, j int, data interface {}) {
	// if i or j less than 0 or greater then nprocs error
	// if i -> j communication unavailable, error
	if i < 0 || i > psim.p || j < 0 || j > psim.p || !psim.topology(i, j) {
		// TODO error
	}

	// send data
	psim.pipes[i][j] <-data
}

func (psim *PSim) recv(i, j int) interface {} {
	// if i or j less than 0 or greater then nprocs error
	// if i -> j communication unavailable, error
	if i < 0 || i > psim.p || j < 0 || j > psim.p || !psim.topology(i, j) {
		// TODO error
	}

	// recv data
	return <-psim.pipes[j][i]
}

func (psim *PSim) one2all_broadcast(rank, source int, data interface {}) interface {} {
	if rank == source {
		for i := 0; i < psim.p; i++ {
			if i != source {
				psim.send(source, i, data)
			}
		}
		return data
	} else {
		return psim.recv(rank, source)
	}
}

func (psim *PSim) all2all_broadcast(rank, source int, data interface {}) interface {} {
	return 0
}

func (psim *PSim) one2all_scatter(rank, source int, data []interface {}) []interface {} {
	if rank == source {
		h, reminder := divmod(len(data), psim.p)

		if reminder > 0 {
			h += 1
		}

		for i := 1; i < psim.p; i++ {
			psim.send(rank, i, data[i*h:i*h+h])
		}
		return data[0:h]
	} else {
		return psim.recv(rank, source)
	}
}



