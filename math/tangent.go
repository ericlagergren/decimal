package math

import (
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/misc"
)

func prepTan(z, x *decimal.Big, ctx decimal.Context) (*decimal.Big, bool) {
	x0 := alias(z, x)
	if pi2 := pi2(x0, ctx); x.CmpAbs(pi2) >= 0 {
		// for cos to work correctly the input must be in (-2pi, 2pi)
		if x.Signbit() {
			ctx.Add(x0, x, pi2)
		} else {
			ctx.Sub(x0, x, pi2)
		}

		var tmp decimal.Big
		ctx.QuoInt(x0, x0, Pi(&tmp, ctx))
		if x.Signbit() {
			ctx.Sub(x0, x0, one)
		} else {
			ctx.Add(x0, x0, one)
		}

		v, ok := x0.Int64()
		if !ok {
			return nil, false
		}
		uv := arith.Abs(v)

		// Adjust so we have ceil(v/10) + ctx.Precision, but check for overflows.
		// 1+((v-1)/10) will be widly incorrect for v == 0, but x/y = 0 iff
		// x = 0 and y != 0. In this case, -2pi <= x >= 2pi, so we're fine.
		prec, ok := checked.Add(1+((uv-1)/10), uint64(ctx.Precision))
		if !ok || prec > maxInt {
			return nil, false
		}
		pctx := decimal.Context{Precision: int(prec)}

		pctx.Mul(x0, Pi(&tmp, pctx), x0)

		ctx.Precision++
		// so toRemove = m*Pi so |x-toRemove| < Pi/2
		ctx.Sub(x0, x, x0)
	} else {
		x0.Copy(x)
	}
	return x0, true
}

// Tan returns the tangent, in radians, of x.
//
// Range:
//     Input: -pi/2 <= x <= pi/2
//     Output: all real numbers
//
// Special cases:
//     Tan(NaN) = NaN
//     Tan(±Inf) = NaN
func Tan(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	if x.IsInf(0) {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	ctx := decimal.Context{Precision: precision(z) + defaultExtraPrecision}
	x0, ok := prepTan(z, x, ctx)
	if !ok {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	ctx.Precision++

	// tan(x) = sign(x)*sqrt(1/cos(x)^2-1)

	// tangent has an asymptote at pi/2 and we'll need more precision as we get
	// closer the reason we need it is as we approach pi/2 is due to the squaring
	// portion, it will cause the small value of Cosine to be come extremely
	// small. we COULD fix it by simply doubling the precision, however, when
	// the precision gets larger it will be a significant impact to performance,
	// instead we'll only add the extra precision when we need it by using the
	// difference to see how much extra precision we need we'll speed things up
	// by only using a quick compare to see if we need to do a deeper inspection.
	tctx := ctx
	var tmp decimal.Big
	if x.CmpAbs(onePtFour) >= 0 {
		if x.Signbit() {
			ctx.Add(&tmp, x, pi2(&tmp, ctx))
		} else {
			ctx.Sub(&tmp, x, pi2(&tmp, ctx))
		}
		tctx.Precision += ctx.Precision + tmp.Scale() - tmp.Precision()
	}
	tmp.Context = tctx

	Cos(&tmp, x0)
	ctx.Mul(&tmp, &tmp, &tmp)
	ctx.Quo(&tmp, one, &tmp)
	ctx.Sub(&tmp, &tmp, one)
	Sqrt(&tmp, &tmp)
	if x0.Signbit() {
		misc.CopyNeg(&tmp, &tmp)
	}
	ctx.Precision -= defaultExtraPrecision + 1
	return ctx.Set(z, &tmp)
}
