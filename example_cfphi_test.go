package decimal

import (
	"fmt"
	"math"
)

type phiGenerator struct {
	ctx Context
}

func (p phiGenerator) Context() Context {
	return p.ctx
}

func (phiGenerator) Next() bool {
	return true
}

func (phiGenerator) Term() Term {
	return Term{A: New(1, 0), B: New(1, 0)}
}

func (p phiGenerator) Lentz() (f, Δ, C, D, eps *Big) {
	// Add a little extra precision to C and D so we get an "exact" result after
	// rounding.
	f = WithContext(p.ctx)
	Δ = WithContext(p.ctx)
	C = WithContext(p.ctx)
	D = WithContext(p.ctx)
	scale := p.ctx.Precision
	if scale == 0 {
		scale = DefaultPrecision
	}
	eps = New(1, scale)
	return f, Δ, C, D, eps
}

// Phi sets z to the golden ratio, φ, and returns z.
func Phi(ctx Context, z *Big) *Big {
	ctx.Precision++
	ctx.Wallis(z, phiGenerator{ctx: ctx})
	ctx.Precision--
	return ctx.Round(z)
}

// This example demonstrates using Lentz by calculating the
// golden ratio, φ.
func ExampleContext_Lentz_phi() {
	ctx := Context64
	var z Big
	Phi(ctx, &z)
	p := (1 + math.Sqrt(5)) / 2

	fmt.Printf(`
Go     : %g
Decimal: %s`, p, &z)
	// Output:
	//
	// Go     : 1.618033988749895
	// Decimal: 1.618033988749895
}
