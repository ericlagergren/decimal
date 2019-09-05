package math_test

import (
	"fmt"

	"github.com/ericlagergren/decimal/v4"
	"github.com/ericlagergren/decimal/v4/math"
)

type tanGenerator struct {
	ctx decimal.Context
	k   uint64
	a   *decimal.Big
	b   *decimal.Big
}

func (t tanGenerator) Next() bool {
	return true
}

func (t *tanGenerator) Term() math.Term {
	t.k += 2
	return math.Term{A: t.a, B: t.b.SetUint64(t.k)}
}

func (t *tanGenerator) Context() decimal.Context { return t.ctx }

// Tan sets z to the tangent of the radian argument x.
func Tan(z, x *decimal.Big) *decimal.Big {
	// Handle special cases like 0, Inf, and NaN.

	// In the continued fraction
	//
	//                z    z^2   z^2   z^2
	//     tan(z) = ----- ----- ----- -----
	//                1  -  3  -  5  -  7  - ···
	//
	// the terms are subtracted, so we need to negate "A"

	ctx := decimal.Context{Precision: z.Context.Precision + 1}
	x0 := ctx.Mul(new(decimal.Big), x, x)
	x0.Neg(x0)

	g := &tanGenerator{
		ctx: ctx,
		k:   1<<64 - 1,
		a:   x0,
		b:   new(decimal.Big),
	}

	// Since our fraction doesn't have a leading (b0) we need to divide our
	// result by a1.
	tan := ctx.Quo(z, x, math.Lentz(z, g))
	ctx.Precision--
	return ctx.Set(z, tan)
}

func ExampleLentz_tan() {
	z := decimal.WithPrecision(17)
	x := decimal.New(42, 0)

	fmt.Printf("tan(42) = %s\n", Tan(z, x).Round(16))
	// Output: tan(42) = 2.291387992437486
}
