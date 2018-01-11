package math

import (
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
)

// Exp sets z to e ** x and returns z.
func Exp(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	if x.IsInf(0) {
		if x.IsInf(+1) {
			// e ** +Inf = +Inf
			return z.SetInf(false)
		}
		// e ** -Inf = 0
		return z.SetUint64(0)
	}

	if x.Sign() == 0 {
		// e ** 0 = 1
		return z.SetUint64(1)
	}

	k := x.Precision() + -x.Scale()
	if k < 0 {
		k = 0
	}
	const expMax = 19
	if k > expMax {
		z.Context.Conditions |= decimal.Inexact | decimal.Rounded
		if x.Signbit() {
			z.Context.Conditions |= decimal.Subnormal | decimal.Underflow | decimal.Clamped
			return z.SetMantScale(0, -etiny(z))
		}
		z.Context.Conditions |= decimal.Overflow
		return z.SetInf(false)
	}

	ctx := decimal.Context{Precision: precision(z) + 3}
	tmp := alias(z, x) // scratch space

	// |x| <= 9 * 10 ** -(prec + 1)
	lim := tmp.SetMantScale(9, ctx.Precision+1)
	if x.CmpAbs(lim) <= 0 {
		z.Context.Conditions |= decimal.Rounded | decimal.Inexact
		return ctx.Round(z.SetMantScale(1, 0).Quantize(ctx.Precision - 1 - 3))
	}
	ctx.Precision += k + 2

	if x.IsInt() {
		if v, ok := x.Uint64(); ok && v == 1 {
			// e ** 1 = e
			return E(z)
		}
	}

	// Argument reduction:
	//    exp(x) = e**r ** 10**k where x = r * 10**k

	r := z.Copy(x).SetScale(x.Scale() + k)
	g := expg{
		prec: ctx.Precision,
		z:    r,
		pow:  ctx.Mul(new(decimal.Big), r, r),
		t:    Term{A: decimal.WithContext(ctx), B: decimal.WithContext(ctx)},
	}

	// TODO(eric): This library provides better performance than other libraries
	// at ~300 digits of precision (compared to libmpdec). Perhaps we should
	// consider using an alternate algorithm for low precision ranges. libmpdec
	// uses Horner's method.

	m := z
	if k != 0 {
		m = decimal.WithContext(ctx)
	}

	Wallis(m, &g)

	if k != 0 {
		k, _ := arith.Pow10(uint64(k))
		fastPowUint(ctx, z, m, k)
	}

	ctx.Precision -= k + 3 + 2
	return ctx.Round(z)
}

// expg is a Generator that computes exp(z).
type expg struct {
	prec int          // Precision
	z    *decimal.Big // Input value
	pow  *decimal.Big // z*z
	m    uint64       // Term number
	t    Term         // Term storage. Does not need to be manually set.
}

func (e *expg) Context() decimal.Context {
	return decimal.Context{Precision: e.prec}
}

func (e *expg) Next() bool { return true }

func (e *expg) Wallis() (a, a1, b, b1, p, eps *decimal.Big) {
	a = decimal.WithPrecision(e.prec)
	a1 = decimal.WithPrecision(e.prec)
	b = decimal.WithPrecision(e.prec)
	b1 = decimal.WithPrecision(e.prec)
	p = decimal.WithPrecision(e.prec)
	eps = decimal.New(1, e.prec)
	return a, a1, b, b1, p, eps
}

func (e *expg) Term() Term {
	// exp(z) can be expressed as the following continued fraction
	//
	//     e^z = 1 +             2z
	//               ------------------------------
	//               2 - z +          z^2
	//                       ----------------------
	//                       6 +        z^2
	//                           ------------------
	//                           10 +     z^2
	//                                -------------
	//                                14 +   z^2
	//                                     --------
	//                                          ...
	//
	// (Khov, p 114)
	//
	// which can be represented as
	//
	//          2z    z^2 / 6    ∞
	//     1 + ----- ---------   K ((a_m^z^2) / 1), z ∈ ℂ
	//          2-z +    1     + m=3
	//
	// where
	//
	//     a_m = 1 / (4 * (2m - 3) * (2m - 1))
	//
	// which can be simplified to
	//
	//     a_m = 1 / (16 * (m-1)^2 - 4)
	//
	// (Cuyt, p 194).
	//
	// References:
	//
	// [Cuyt] - Cuyt, A.; Petersen, V.; Brigette, V.; Waadeland, H.; Jones, W.B.
	// (2008). Handbook of Continued Fractions for Special Functions. Springer
	// Netherlands. https://doi.org/10.1007/978-1-4020-6949-9
	//
	// [Khov] - Merkes, E. P. (1964). The Application of Continued Fractions and
	// Their Generalizations to Problems in Approximation Theory
	// (A. B. Khovanskii). SIAM Review, 6(2), 188–189.
	// https://doi.org/10.1137/1006052

	switch e.m {
	// [0, 1]
	case 0:
		e.t.A.SetUint64(0)
		e.t.B.SetUint64(1)
	// [2z, 2-z]
	case 1:
		e.t.A.Mul(two, e.z)
		e.t.B.Sub(two, e.z)
	// [z^2/6, 1]
	case 2:
		e.t.A.Quo(e.pow, six)
		e.t.B.SetUint64(1)
	// [(1/(16((m-1)^2)-4))(z^2), 1]
	default:
		// maxM is the largest m value we can use to compute 4(2m - 3)(2m - 1)
		// using unsigned integers.
		const maxM = 1518500252

		// 4(2m - 3)(2m - 1) ≡ 16(m - 1)^2 - 4
		if e.m <= maxM {
			e.t.A.SetUint64(16*((e.m-1)*(e.m-1)) - 4)
		} else {
			e.t.A.SetUint64(e.m - 1)

			// (m-1)^2
			e.t.A.Mul(e.t.A, e.t.A)

			// 16 * (m-1)^2 - 4 = 16 * (m-1)^2 + (-4)
			e.t.A.FMA(sixteen, e.t.A, negfour)
		}

		// 1 / (16 * (m-1)^2 - 4)
		e.t.A.Quo(one, e.t.A)

		// 1 / (16 * (m-1)^2 - 4) * (z^2)
		e.t.A.Mul(e.t.A, e.pow)

		// e.t.B is set to 1 inside case 2.
	}

	e.m++
	return e.t
}

var _ Walliser = (*expg)(nil)
