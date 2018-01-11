package math_test

import (
	"fmt"
	gmath "math"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
)

var one = new(decimal.Big).SetUint64(1)

type phiGenerator struct{ ctx decimal.Context }

func (p phiGenerator) Context() decimal.Context { return p.ctx }

func (p phiGenerator) Next() bool {
	return true
}

func (p phiGenerator) Term() math.Term {
	return math.Term{A: one, B: one}
}

func (p phiGenerator) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	// Add a little extra precision to C and D so we get an "exact" result after
	// rounding.
	f = decimal.WithPrecision(p.ctx.Precision)
	Δ = decimal.WithPrecision(p.ctx.Precision)
	C = decimal.WithPrecision(p.ctx.Precision)
	D = decimal.WithPrecision(p.ctx.Precision)
	eps = decimal.New(1, p.ctx.Precision)
	return f, Δ, C, D, eps
}

// Phi sets z to the golden ratio, φ, and returns z.
func Phi(z *decimal.Big) *decimal.Big {
	ctx := z.Context
	ctx.Precision++
	math.Wallis(z, phiGenerator{ctx: ctx})
	ctx.Precision--
	return ctx.Round(z)
}

// This example demonstrates using Lentz by calculating the golden ratio, φ.
func ExampleLentz_phi() {
	z := decimal.WithPrecision(16)
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
