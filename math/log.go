package math

import (
	stdMath "math"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/c"
)

// Log10 sets z to the common logarithm of x and returns z.
func Log10(z, x *decimal.Big) *decimal.Big {
	if logSpecials(z, x) {
		return z
	}

	// If x is a power of 10 the result is the exponent and exact.
	var tpow bool
	if m, u := decimal.Raw(x); *m != c.Inflated {
		tpow = arith.PowOfTen(*m)
	} else {
		tpow = arith.PowOfTenBig(u)
	}
	if tpow {
		ctx := decimal.Context{Precision: precision(z)}
		return ctx.Set(z, z.SetMantScale(int64(adjusted(x)), 0))
	}
	return log(z, x, true)
}

// Log sets z to the natural logarithm of x and returns z.
func Log(z, x *decimal.Big) *decimal.Big {
	if logSpecials(z, x) {
		return z
	}
	if x.IsInt() {
		if v, ok := x.Uint64(); ok {
			switch v {
			case 1:
				// ln 1 = 0
				return z.SetMantScale(0, 0)
			case 10:
				// Specialized function.
				return ln10(z, precision(z))
			}
		}
	}
	return log(z, x, false)
}

// logSepcials checks for special values (Inf, NaN, 0) for logarithms.
func logSpecials(z, x *decimal.Big) bool {
	if z.CheckNaNs(x, nil) {
		return true
	}

	if sgn := x.Sign(); sgn <= 0 {
		if sgn == 0 {
			// ln 0 = -Inf
			z.SetInf(true)
		} else {
			// ln -x is undefined.
			z.Context.Conditions |= decimal.InvalidOperation
			z.SetNaN(false)
		}
		return true
	}

	if x.IsInf(+1) {
		// ln +Inf = +Inf
		z.SetInf(false)
		return true
	}
	return false
}

// log set z to log(x), or log10(x) if ten. It does not check for special values,
// nor implement any special casing.
func log(z, x *decimal.Big, ten bool) *decimal.Big {
	prec := precision(z)

	t := int64(adjusted(x))
	if t < 0 {
		t = -t - 1
	}
	t *= 2

	if arith.Length(arith.Abs(t))-1 > maxscl(z) {
		z.Context.Conditions |= decimal.Overflow | decimal.Inexact | decimal.Rounded
		return z.SetInf(t < 0)
	}

	// Argument reduction:
	// Given
	//    ln(a) = ln(b) + ln(c), where a = b * c
	// Given
	//    x = m * 10**n, where x ∈ ℝ
	// Reduce x (as y) so that
	//    1 <= y < 10
	// And create p so that
	//    x = y * 10**p
	// Compute
	//    log(y) + p*log(10)

	// TODO(eric): adj should be large enough. It's passed multiple iterations
	// of with a precision in [1, 5000) and a 128-bit decimal.
	adj := 6 + (3 * int(stdMath.Log(float64(x.Precision()))))
	if ten {
		adj += 3
	}
	ctx := decimal.Context{Precision: prec + adj}

	var p int64
	switch {
	// 1e+1000
	case x.Scale() <= 0:
		p = int64(x.Precision() - x.Scale() - 1)
	// 0.0001
	case x.Scale() >= x.Precision():
		p = -int64(x.Scale() - x.Precision() + 1)
	// 12.345
	default:
		p = int64(-x.Scale() + x.Precision() - 1)
	}

	// Rescale to 1 <= x <= 10
	y := decimal.WithContext(ctx).Copy(x).SetScale(x.Precision() - 1)
	// Continued fraction algorithm is for log(1+x)
	y.Sub(y, one)

	g := lgen{
		ctx: ctx,
		pow: ctx.Mul(new(decimal.Big), y, y),
		z2:  ctx.Add(new(decimal.Big), y, two),
		k:   -1,
		t:   Term{A: new(decimal.Big), B: new(decimal.Big)},
	}

	// TODO(eric): Similar to the comment inside Exp, this library only provides
	// better performance at ~750 digits of precision. Consider using Newton's
	// method or another algorithm for lower precision ranges.

	ctx.Quo(z, y.Mul(y, two), Lentz(z, &g))

	if p != 0 || ten {
		t := ln10(y, ctx.Precision) // recycle y

		// Avoid doing unnecessary work.
		switch p {
		default:
			p := g.z2.SetMantScale(p, 0) // recycle g.z2
			ctx.FMA(z, p, t, z)
		case 0:
			// OK
		case -1:
			ctx.Sub(z, z, t) // (-1 * t) + z = -t + z = z - t
		case 1:
			ctx.Add(z, t, z) // (+1 * t) + z = t + z
		}

		// We're calculating log10(x):
		//    log10(x) = log(x) / log(10)
		if ten {
			ctx.Quo(z, z, t)
		}
	}
	ctx.Precision -= adj
	return ctx.Round(z)
}

type lgen struct {
	ctx decimal.Context
	pow *decimal.Big // z*z
	z2  *decimal.Big // z+2
	k   int64
	t   Term
}

func (l *lgen) Context() decimal.Context { return l.ctx }

func (l *lgen) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	f = decimal.WithContext(l.ctx)
	Δ = decimal.WithContext(l.ctx)
	C = decimal.WithContext(l.ctx)
	D = decimal.WithContext(l.ctx)
	eps = decimal.New(1, l.ctx.Precision)
	return
}

func (a *lgen) Next() bool { return true }

func (a *lgen) Term() Term {
	// log(z) can be expressed as the following continued fraction:
	//
	//          2z      1^2 * z^2   2^2 * z^2   3^2 * z^2   4^2 * z^2
	//     ----------- ----------- ----------- ----------- -----------
	//      1 * (2+z) - 3 * (2+z) - 5 * (2+z) - 7 * (2+z) - 9 * (2+z) - ···
	//
	// (Cuyt, p 271).
	//
	// References:
	//
	// [Cuyt] - Cuyt, A.; Petersen, V.; Brigette, V.; Waadeland, H.; Jones, W.B.
	// (2008). Handbook of Continued Fractions for Special Functions. Springer
	// Netherlands. https://doi.org/10.1007/978-1-4020-6949-9

	a.k += 2
	if a.k != 1 {
		a.t.A.SetMantScale(-((a.k / 2) * (a.k / 2)), 0)
		a.ctx.Mul(a.t.A, a.t.A, a.pow)
	}
	a.t.B.SetMantScale(a.k, 0)
	a.ctx.Mul(a.t.B, a.t.B, a.z2)
	return a.t
}
