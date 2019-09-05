package math

import (
	stdMath "math"

	"github.com/ericlagergren/decimal/v4"
	"github.com/ericlagergren/decimal/v4/misc"
)

var sqrt3_3 = decimal.New(577350, 6) // sqrt(3) / 3

func prepAtan(z, x *decimal.Big, ctx decimal.Context) (
	*decimal.Big,
	*decimal.Big,
	*decimal.Big,
	int,
	int,
) {
	ctx.Precision += defaultExtraPrecision

	x0 := alias(z, x)

	// since smaller values converge faster
	// we'll use argument reduction
	// if |x| > 1 then set x = 1/x
	segment := 0
	if x.CmpAbs(one) > 0 {
		segment = 2
		ctx.Quo(x0, one, x)
	} else if x.CmpAbs(sqrt3_3) > 0 {
		// if |x| > sqrt(3)/3 (approximated to 0.577350)
		segment = 1

		// then set x = (sqrt(3)*x-1)/(sqrt(3)+x)
		sqrt3 := sqrt3(new(decimal.Big), ctx)
		ctx.Mul(x0, sqrt3, x)    // sqrt(3) * x
		ctx.Sub(x0, x0, one)     // sqrt(3)*x - 1
		ctx.Add(sqrt3, sqrt3, x) // sqrt(3) + x
		ctx.Quo(x0, x0, sqrt3)   // (sqrt(3)*x - 1) / (sqrt(3) + x)
	} else {
		x0.Copy(x)
	}

	// next we'll use argument halving technic
	// atan(y) = 2 atan(y/(1+sqrt(1+y^2)))
	// we'll repeat this up to a point
	// we have competing operations at some
	// point the sqrt causes a problem
	// note (http:// fredrikj.net/blog/2014/12/faster-high-ctx-Atangents/)
	// suggests that r= 8 times is a good value for
	// precision with 1000s of digits to millions
	// however it was easy to determine there is a
	// better sliding window to use instead
	// which is what we use as it turns out
	// when the ctx is large, a larger r value
	// compared to 8 is every effective
	xf, _ := x0.Float64()
	xf = stdMath.Abs(xf)
	// the formula simple but works a bit better then a fixed value (8)
	r := stdMath.Max(0, stdMath.Ceil(0.31554321636851*stdMath.Pow(float64(ctx.Precision), 0.654095561044508)))
	var p float64

	// maxPrec is the largest precision value we can use bit shifts instead of
	// math.Pow which is more expensive.
	const maxPrec = 3286
	if ctx.Precision <= maxPrec {
		p = 1 / float64(uint64(1)<<uint64(r))
	} else {
		p = stdMath.Pow(2, -r)
	}
	halved := int(stdMath.Ceil(stdMath.Log(xf/p) / stdMath.Ln2))

	// if the value is already less than 1/(2^r) then halfed
	// will be negative and we don't need to apply
	// the double angle formula because it would hurt performance
	// so we'll set halfed to zero
	if halved < 0 {
		halved = 0
	}

	sq := decimal.WithContext(ctx)
	for i := 0; i < halved; i++ {
		ctx.FMA(sq, x0, x0, one)
		Sqrt(sq, sq)
		ctx.Add(sq, sq, one)
		ctx.Quo(x0, x0, sq)
	}

	var x2 decimal.Big
	ctx.Mul(&x2, x0, x0)
	var x2p1 decimal.Big
	ctx.Add(&x2p1, &x2, one)
	return x0, &x2, &x2p1, segment, halved
}

func getAtanP(ctx decimal.Context, x2 *decimal.Big) SplitFunc {
	var p decimal.Big
	return func(n uint64) *decimal.Big {
		// P(n) = 2n for all n > 0
		if n == 0 {
			return one
		}
		if n < stdMath.MaxUint64/2 {
			p.SetUint64(n * 2)
		} else {
			ctx.Mul(&p, p.SetUint64(n), two)
		}
		return ctx.Mul(&p, &p, x2)
	}
}

func getAtanQ(ctx decimal.Context, x2p1 *decimal.Big) SplitFunc {
	var q decimal.Big
	return func(n uint64) *decimal.Big {
		// B(n) = (2n+1) for all n >= 0

		// atanMax is the largest number we can use to compute (2n + 1) without
		// overflow.
		const atanMax = (stdMath.MaxUint64 - 1) / 2
		if n < atanMax {
			q.SetUint64((2 * n) + 1)
		} else {
			ctx.FMA(&q, q.SetUint64(n), two, one)
		}
		return ctx.Mul(&q, &q, x2p1)
	}
}

// Atan returns the arctangent, in radians, of x.
//
// Range:
//     Input: all real numbers
//     Output: -pi/2 <= Atan(x) <= pi/2
//
// Special cases:
//		Atan(NaN)  = NaN
//		Atan(±Inf) = ±x * pi/2
func Atan(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	ctx := decimal.Context{Precision: precision(z) + defaultExtraPrecision}

	if x.IsInf(0) {
		pi2(z, ctx)
		if x.IsInf(-1) {
			misc.SetSignbit(z, true)
		}
		return z
	}

	//when x <-1 we use -atan(-x) instead
	if x.Cmp(negone) < 0 {
		Atan(z, new(decimal.Big).Neg(x))
		return z.Neg(z)
	}

	y, ySq, ySqPlus1, segment, halfed := prepAtan(z, x, ctx) // z == y, maybe.
	result := BinarySplitDynamic(ctx,
		func(_ uint64) *decimal.Big { return y },
		getAtanP(ctx, ySq),
		func(_ uint64) *decimal.Big { return one },
		getAtanQ(ctx, ySqPlus1),
	)

	// undo the double angle part
	ySq.Context = ctx
	tmp := Pow(ySq, two, z.SetMantScale(int64(halfed), 0)) // clobber ySq
	ctx.Mul(z, result, tmp)

	// to handle the argument reduction step
	// check which segment the value was from
	// seg 0: 0 < value <= sqrt(3)/3  // then result = result
	// seg 1: sqrt(3)/3 < value <= 1  // then result = pi/6 + result
	// set 2: 1 < value               // then result = pi/2 - result
	switch segment {
	case 1:
		piOver6 := pi(tmp, ctx) // clobber _2p
		ctx.Quo(piOver6, piOver6, six)
		ctx.Add(z, piOver6, z)
	case 2:
		ctx.Sub(z, pi2(tmp, ctx), z) // clobber _2p
	}
	ctx.Precision -= defaultExtraPrecision
	return ctx.Round(z)
}

// Atan2 calculates arctan of y/x and uses the signs of y and x to determine
// the valid quadrant
//
// Range:
//     y input: all real numbers
//     x input: all real numbers
//     Output: -pi < Atan2(y, x) <= pi
//
// Special cases:
//     Atan2(NaN, NaN)      = NaN
//     Atan2(y, NaN)        = NaN
//     Atan2(NaN, x)        = NaN
//     Atan2(±0, x >=0)     = ±0
//     Atan2(±0, x <= -0)   = ±pi
//     Atan2(y > 0, 0)      = +pi/2
//     Atan2(y < 0, 0)      = -pi/2
//     Atan2(±Inf, +Inf)    = ±pi/4
//     Atan2(±Inf, -Inf)    = ±3pi/4
//     Atan2(y, +Inf)       = 0
//     Atan2(y > 0, -Inf)   = +pi
//     Atan2(y < 0, -Inf)   = -pi
//     Atan2(±Inf, x)       = ±pi/2
//     Atan2(y, x > 0)      = Atan(y/x)
//     Atan2(y >= 0, x < 0) = Atan(y/x) + pi
//     Atan2(y < 0, x < 0)  = Atan(y/x) - pi
func Atan2(z, y, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(y, x) {
		return z
	}

	// Return context and work context. For this function it's easier to have
	// two separate contexts than it is to constantly subtract our extra precision.
	rctx := decimal.Context{Precision: precision(z)}
	wctx := decimal.Context{Precision: rctx.Precision + defaultExtraPrecision}

	neg := y.Signbit()
	xs := x.Sign()

	// Special cases.
	if y.Sign() == 0 {
		if xs >= 0 && !x.Signbit() {
			// Atan2(y == 0, x >= +0)  = ±0
			return z.CopySign(z.SetUint64(0), y)
		}
		// Atan2(y == 0, x <= -0) = ±pi
		return rctx.Round(misc.SetSignbit(pi(z, wctx), neg))
	}
	if xs == 0 {
		// Atan2(y, x == 0) = ±pi/2
		return rctx.Round(misc.SetSignbit(pi2(z, wctx), neg))
	}
	if x.IsInf(0) {
		if x.IsInf(+1) {
			if y.IsInf(0) {
				// Atan2(±Inf, +Inf) = ±pi/4
				wctx.Quo(z, pi(z, wctx), four)
				return rctx.Round(misc.SetSignbit(z, neg))
			}
			// Atan2(y, +Inf) = ±0
			return rctx.Round(misc.SetSignbit(z.SetUint64(0), neg))
		}
		if y.IsInf(0) {
			// Atan2(±Inf, -Inf) = ±3 * pi/4
			wctx.Quo(z, pi(z, wctx), four)
			wctx.Mul(z, z, three)
			return rctx.Round(misc.SetSignbit(z, neg))
		}
		// Atan2(y, -Inf) = ±pi
		return rctx.Round(misc.SetSignbit(pi(z, wctx), neg))
	}
	if y.IsInf(0) {
		// Atan2(±Inf, x) = ±pi/2
		return rctx.Round(misc.SetSignbit(pi2(z, wctx), neg))
	}

	Atan(z, wctx.Quo(z, y, x))
	if xs < 0 {
		pi := pi(new(decimal.Big), wctx)
		if z.Sign() <= 0 {
			wctx.Add(z, z, pi)
		} else {
			wctx.Sub(z, z, pi)
		}
	}
	return rctx.Round(z)
}
