package psim

import (
	"math"
)

func BUS(i, j int) bool {
	return true
}

func SWITCH(i,j int) bool {
	return true
}

func MESH1(p int) func(i, j int) bool {
	return func(i, j int) bool {
		return Pow((i - j), 2) == 1.0
	}
}

func TORUS1(p int) func(i, j int) bool {
	return func(i, j int) {
		return (i - j + p) % p == 1 || (j - i + p) % p == 1
	}
}

func MESH2(p int) func(i, j int) bool {
	q := Floor(Sqrt(p) + 0.1)
	return func(i, j int) {
		return Pow((i%q-j%q), 2) ^ Pow((i/q-j/q), 2)
	}
}

func TORUS2(p int) func(i, j int) bool {
	q := Floor(Sqrt(p) + 0.1)
	return func(i, j int) {
		return ((i % q - j % q + q) % q) ^ ((i / q - j / q + q) % q) || ((j % q - i % q + q) % q) ^ ((j / q - i / q + q) % q)
	}
}
func TREE(i, j int) bool {
	return i == Ceil((j - 1) / 2) || j == Ceil((i - 1) / 2)
}
