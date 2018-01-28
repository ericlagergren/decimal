package math

import (
	stdMath "math"
	"math/bits"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/misc"
)

const maxInt = 1<<(bits.UintSize-1) - 1 // Also: uint64(int(^uint(0) << 1))

func prepCosine(z, x *decimal.Big, ctx decimal.Context) (*decimal.Big, int, bool) {
	x0 := alias(z, x)

	var tmp decimal.Big
	var pi decimal.Big

	// for better results, we need to make sure the value we're working with a
	// value is closer to zero.
	ctx.Mul(&pi, Pi(&pi, ctx), two) // 2 * Pi
	if x.CmpAbs(&pi) >= 0 {
		// for cos to work correctly the input must be in (-2Pi, 2Pi).
		ctx.Quo(&tmp, x, &pi)
		v, ok := tmp.Int64()
		if !ok {
			return nil, 0, false
		}
		uv := arith.Abs(v)

		// Adjust so we have ceil(v/10) + ctx.Precision, but check for overflows.
		// 1+((v-1)/10) will be widly incorrect for v == 0, but x/y = 0 iff
		// x = 0 and y != 0. In this case, -2pi <= x >= 2pi, so we're fine.
		prec, ok := checked.Add(1+((uv-1)/10), uint64(ctx.Precision))
		if !ok || prec > maxInt {
			return nil, 0, false
		}
		pctx := decimal.Context{Precision: int(prec)}

		if uv <= stdMath.MaxInt64/2 {
			tmp.SetMantScale(v, 0)
		}

		pctx.Mul(&tmp, &pi, &tmp)

		// so toRemove = 2*Pi*v so x - toRemove < 2*Pi
		ctx.Sub(x0, x, &tmp)
	} else {
		x0.Copy(x)
	}

	// add 1 to the precision for the up eventual squaring.
	ctx.Precision++

	// now preform the half angle.
	// we'll repeat this up to log(precision)/log(2) and keep track
	// since we'll be dividing x we need to give a bit more precision
	// we'll be repeated applying the double angle formula
	// we could use a higher angle formula but wouldn't buy us anything.
	// Each time we half we'll have to increase the values precision by 1
	// and since we'll dividing at most 11 time that means at most 11 digits
	// but we'll figure out the minimum time we'll apply the double angle
	// formula

	// we'll we reduce until it's x <= p where p= 1/2^8 (approx 0.0039)
	// we figure out the number of time to divide by solving for r
	// in x/p = 2^r  so r = log(x/p)/log(2)

	xf, _ := x0.Float64()
	// We only need to do the calculation if xf >= 0.0004. Anything below that
	// and we're <= 0.
	var halved int
	if xf = stdMath.Abs(xf); xf >= 0.0004 {
		// Originally: ceil(log(xf/0.0048828125) / ln2)
		halved = int(stdMath.Ceil(1.4427*stdMath.Log(xf) + 11))
		// The general case is halved > 0, since we only get 0 if xf is very
		// close to 0.0004.
		if halved > 0 {
			// Increase the precision based on the number of divides. Overflow is
			// unlikely, but possible.
			ctx.Precision += halved
			if ctx.Precision <= halved {
				return nil, 0, false
			}
			// The maximum value for halved will be 14, given
			//     ceil(1.4427*log(2*pi)+11) = 14
			ctx.Quo(x0, x0, tmp.SetUint64(1<<uint64(halved)))
		}
	}

	ctx.Mul(x0, x0, x0)
	misc.CopyNeg(x0, x0)
	return x0, halved, true
}

func getCosineP(negXSq *decimal.Big) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		if n == 0 {
			return one
		}
		return negXSq
	}
}

func getCosineQ(ctx decimal.Context) func(n uint64) *decimal.Big {
	var q, tmp decimal.Big
	return func(n uint64) *decimal.Big {
		// (0) = 1, q(n) = 2n(2n-1) for n > 0
		if n == 0 {
			return one
		}

		// most of the time n will be a small number so
		// use the fastest method to calculate 2n(2n-1)
		const cosine4NMaxN = 2147483648
		if n < cosine4NMaxN {
			return q.SetUint64((2 * n) * (2*n - 1)) // ((n*n) << 2) - (n << 1)
		}
		q.SetUint64(n)
		ctx.Mul(&tmp, &q, two)
		ctx.Mul(&q, &tmp, &tmp)
		return ctx.Sub(&q, &q, &tmp)
	}
}

// Cos returns the cosine, in radians, of x.
//
// Range:
//     Input: all real numbers
//     Output: -1 <= Cos(x) <= 1
//
// Special cases:
//		Cos(NaN)  = NaN
//		Cos(±Inf) = NaN
func Cos(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	if x.IsInf(0) {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	ctx := decimal.Context{Precision: precision(z) + defaultExtraPrecision}

	negXSq, halved, ok := prepCosine(z, x, ctx)
	if !ok {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	ctx.Precision += halved
	z.Copy(BinarySplitDynamic(ctx,
		func(_ uint64) *decimal.Big { return one },
		getCosineP(negXSq),
		func(_ uint64) *decimal.Big { return one },
		getCosineQ(ctx),
	))

	// now undo the half angle bit
	for i := 0; i < halved; i++ {
		ctx.Mul(z, z, z)
		ctx.Mul(z, z, two)
		ctx.Sub(z, z, one)
	}
	ctx.Precision -= defaultExtraPrecision + halved
	return ctx.Round(z)
}
