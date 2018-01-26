package math

import "github.com/ericlagergren/decimal"

// Acos returns the arccosine, in radians, of x.
//
// Range:
//     Input: -1 <= x <= 1
//     Output: 0 <= Acos(x) <= pi
//
// Special cases:
//     Acos(NaN)  = NaN
//     Acos(±Inf) = NaN
//     Acos(x)    = NaN if x < -1 or x > 1
//     Acos(-1)   = pi
//     Acos(1)    = 0
func Acos(z, x *decimal.Big) *decimal.Big {
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
		if x.Signbit() {
			return Pi(z, ctx)
		}
		return z.SetUint64(0)
	}

	ctx.Precision += defaultExtraPrecision

	// Acos(x) = pi/2 - arcsin(x)

	// TODO(eric): when I devise an API for Pi, E, etc. that uses Context, switch
	// that that instead of allocating new decimals.
	ctx.Sub(z, pi2(z, ctx), Asin(decimal.WithContext(ctx), x))
	ctx.Precision -= defaultExtraPrecision
	return ctx.Round(z)
}
