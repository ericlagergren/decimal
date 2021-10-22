package decimal

import (
	"fmt"
)

type tanGenerator struct {
	ctx Context
	k   uint64
	a   *Big
	b   *Big
}

func (t tanGenerator) Next() bool {
	return true
}

func (t *tanGenerator) Term() Term {
	t.k += 2
	return Term{A: t.a, B: t.b.SetUint64(t.k)}
}

func (t *tanGenerator) Context() Context { return t.ctx }

// Tan sets z to the tangent of the radian argument x.
func Tan(ctx Context, z, x *Big) *Big {
	// Handle special cases like 0, Inf, and NaN.

	// In the continued fraction
	//
	//                z    z^2   z^2   z^2
	//     tan(z) = ----- ----- ----- -----
	//                1  -  3  -  5  -  7  - ···
	//
	// the terms are subtracted, so we need to negate "A"

	work := ctx
	if work.Precision == 0 {
		work.Precision = DefaultPrecision
	}
	work.Precision++

	x0 := work.Mul(new(Big), x, x)
	x0.Neg(x0)

	g := &tanGenerator{
		ctx: ctx,
		k:   1<<64 - 1,
		a:   x0,
		b:   new(Big),
	}

	// Since our fraction doesn't have a leading (b0), we need to
	// divide our result by a1.
	tan := work.Quo(z, x, work.Lentz(z, g))

	return ctx.Set(z, tan)
}

func ExampleContext_Lentz_tan() {
	var z Big
	ctx := Context{Precision: 17}
	x := New(42, 0)

	fmt.Printf("tan(42) = %s\n", Tan(ctx, &z, x).Round(16))
	// Output: tan(42) = 2.291387992437486
}
