package math_test

import (
	"fmt"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
)

type tanGenerator struct {
	k int64
	a *decimal.Big
	b *decimal.Big
}

func (t tanGenerator) Next() bool {
	return true
}

func (t *tanGenerator) Term() math.Term {
	t.k += 2
	return math.Term{A: t.a, B: t.b.SetMantScale(t.k, 0)}
}

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
	x0 := new(decimal.Big).Mul(x, x)
	x0.Neg(x0)

	g := &tanGenerator{k: -1, a: x0, b: new(decimal.Big)}

	tan := math.Lentz(z, g)

	// Since our fraction doesn't have a leading (b0) we need to divide our
	// result by a1.
	tan.Quo(x, tan)

	return z.Set(tan)
}

func ExampleLentz_tan() {
	z := new(decimal.Big)
	x := decimal.New(42, 0)

	fmt.Printf("tan(42) = %s\n", Tan(z, x))
	// Output: tan(42) = 2.291387992437486
}
