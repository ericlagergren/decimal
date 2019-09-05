package math

import (
	"github.com/ericlagergren/decimal/v4"
	"github.com/ericlagergren/decimal/v4/misc"
)

// Asin returns the arcsine, in radians, of x.
//
// Range:
//     Input: -1 <= x <= 1
//     Output: -pi/2 <= Asin(x) <= pi/2
//
// Special cases:
//		Asin(NaN)  = NaN
//		Asin(±Inf) = NaN
//		Asin(x)    = NaN if x < -1 or x > 1
//		Asin(±1)   = ±pi/2
func Asin(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	cmp1 := x.CmpAbs(one)
	if x.IsInf(0) || cmp1 > 0 {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	ctx := decimal.Context{Precision: precision(z)}

	if cmp1 == 0 {
		pi2(z, ctx)
		if x.Signbit() {
			misc.SetSignbit(z, true)
		}
		return z
	}

	ctx.Precision += defaultExtraPrecision

	// Asin(x) = 2 * atan(x / (1 + sqrt(1 - x*x)))

	x2 := ctx.Mul(alias(z, x), x, x)
	ctx.Quo(x2, x, ctx.Add(x2, Sqrt(x2, ctx.Sub(x2, one, x2)), one))
	z.Copy(Atan(decimal.WithContext(ctx), x2))
	ctx.Mul(z, z, two)
	ctx.Precision -= defaultExtraPrecision
	return ctx.Round(z)
}
