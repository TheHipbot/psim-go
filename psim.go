/*
Package psimgo is a go implementation of the library PSim for python.
PSim is used to simulate parallel processing on a single machine.


 */
package psimgo

import (
	"math"
)

// Go booleans cannot be treated as integers using
// the ^ xor operation. Created a simple function
// to be used in following code
func xor(i, j, v int) bool {
	return (i == v || j == v) && i != j
}

// Topology Communication functions
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

type PSim struct {
	p int
	topology func(i, j int) bool
}

func (psim PSim) send(j int) {

}




