package math_test

import (
	"fmt"
	gmath "math"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
)

var one = decimal.New(1, 0)

type phiGenerator struct{ prec int }

func (p phiGenerator) Next() bool {
	return true
}

func (p phiGenerator) Term() math.Term {
	return math.Term{A: one, B: one}
}

func (p phiGenerator) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	// Add a little extra precision to C and D so we get an "exact" result after
	// rounding.
	f = decimal.WithPrecision(p.prec + 1)
	Δ = decimal.WithPrecision(p.prec + 1)
	C = decimal.WithPrecision(p.prec + 1)
	D = decimal.WithPrecision(p.prec + 1)
	eps = decimal.New(1, p.prec)
	return f, Δ, C, D, eps
}

// Phi sets z to the golden ratio, φ, and returns z.
func Phi(z *decimal.Big) *decimal.Big {
	p := z.Context.Precision
	if p == 0 {
		p = decimal.DefaultPrecision
	}
	g := phiGenerator{prec: p}
	return math.Lentz(z, g)
}

// This example demonstrates using Lentz by calculating the golden ratio, φ.
func ExampleLentz_phi() {
	z := new(decimal.Big)
	Phi(z)
	p := (1 + gmath.Sqrt(5)) / 2

	fmt.Printf(`
Go     : %g
Decimal: %s`, p, z)
	// Output:
	//
	// Go     : 1.618033988749895
	// Decimal: 1.618033988749895
}
