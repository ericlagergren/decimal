package math

import (
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
	tpow := false
	if m, u := decimal.Raw(x); m != c.Inflated {
		tpow = arith.PowOfTen(m)
	} else {
		tpow = arith.PowOfTenBig(u)
	}
	if tpow {
		return z.SetMantScale(int64(adjusted(x)), 0)
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
				prec := precision(z)
				// Specialized function.
				return round(ln10(z, prec+3), prec)
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

	t := int64(x.Scale()-x.Precision()) - 1
	if t < 0 {
		t = -t - 1
	}
	t *= 2

	if arith.Length(arith.Abs(t))-1 > decimal.MaxScale {
		z.Context.Conditions |= decimal.Overflow | decimal.Inexact | decimal.Rounded
		return z.SetInf(t < 0)
	}

	/* TODO(eric): figure out how to ensure this is corectly rounded.
	if f, exact := x.Float64(); exact && prec <= exactFloatPrec {
		var res float64
		if ten {
			res = math.Log10(f)
		} else {
			res = math.Log(f)
		}
		return round(z.SetFloat64(res), prec)
	}
	*/

	// Argument reduction:
	// Given
	//    ln(a) = ln(b) + ln(c)
	// Where
	//    a = b * c
	// Given
	//    x = m * 10**n
	// Reduce x (as y) so that
	//    1 <= y < 10
	// And create p so that
	//    x = y * 10**p
	// Compute
	//    log(y) + p*log(10)

	const adj = 4
	prec += adj
	y := decimal.WithPrecision(prec).Copy(x).SetScale(x.Precision() - 1)
	// Algorithm is for log(1+x)
	y.Sub(y, one)

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

	lgp := prec + 2
	g := lgen{
		prec: lgp,
		pow:  decimal.WithPrecision(lgp).Mul(y, y),
		z2:   decimal.WithPrecision(lgp).Add(y, two),
		k:    -1,
		t:    Term{A: decimal.WithPrecision(lgp), B: decimal.WithPrecision(lgp)},
	}

	tmp := decimal.WithPrecision(prec)
	y.Quo(y.Mul(y, two), Lentz(tmp, &g))

	if p != 0 || ten {
		t := ln10(tmp, prec) // recycle tmp

		// Avoid doing unnecessary work.
		switch p {
		default:
			p := g.pow.SetMantScale(p, 0) // recycle g.pow
			y.FMA(p, t, y)
		case 0:
			// OK
		case -1:
			y.Sub(y, t) // (-1 * t) + y = -t + y = y - t
		case 1:
			y.Add(t, y) // (+1 * t) + y = t + y
		}

		// We're calculating log10(x):
		//    log10(x) = log(x) / log(10)
		if ten {
			y.Quo(y, t)
		}
	}
	z.Context.Conditions |= tmp.Context.Conditions
	return z.Copy(round(y, prec-adj))
}

type lgen struct {
	prec int
	pow  *decimal.Big // z*z
	z2   *decimal.Big // z+2
	k    int64
	t    Term
}

func (l *lgen) mode() decimal.RoundingMode { return decimal.ToNearestEven }

func (l *lgen) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	f = decimal.WithPrecision(l.prec)
	Δ = decimal.WithPrecision(l.prec)
	C = decimal.WithPrecision(l.prec)
	D = decimal.WithPrecision(l.prec)
	eps = decimal.New(1, l.prec)
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
		a.t.A.Mul(a.t.A, a.pow)
	}
	a.t.B.SetMantScale(a.k, 0)
	a.t.B.Mul(a.t.B, a.z2)
	return a.t
}
