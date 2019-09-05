package math

import "github.com/ericlagergren/decimal/v4"

// Sin returns the sine, in radians, of x.
//
// Range:
//     Input: all real numbers
//     Output: -1 <= Sin(x) <= 1
//
// Special cases:
//     Sin(NaN) = NaN
//     Sin(Inf) = NaN
func Sin(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}
	if x.IsInf(0) {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	// Sin(x) = Cos(pi/2 - x)
	ctx := decimal.Context{Precision: precision(z) + defaultExtraPrecision}
	Cos(z, ctx.Sub(z, pi2(alias(z, x), ctx), x))
	ctx.Precision -= defaultExtraPrecision
	return ctx.Round(z)
}
