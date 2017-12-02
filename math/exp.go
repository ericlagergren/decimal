package math

import (
	"github.com/ericlagergren/decimal"
)

// Exp sets z to e ** x and returns z.
func Exp(z, x *decimal.Big) *decimal.Big {
	// TODO(eric): "pestle_: eric_lagergren, that is, exp(z+z0) = exp(z)*exp(z0)
	// 						 and exp(z) ~ 1+z for small enough z"

	if z.CheckNaNs(x, nil) {
		return z
	}

	if x.IsInf(0) {
		if x.IsInf(+1) {
			// e ** +Inf = +Inf
			return z.SetInf(false)
		}
		// e ** -Inf = 0
		return z.SetMantScale(0, 0)
	}

	if x.Sign() == 0 {
		// e ** 0 = 1
		return z.SetMantScale(1, 0)
	}

	// The algorithm behind this section is taken from libmpdec, which uses an
	// algorithm from Hull & Abraham, Variable Precision Exponential Function,
	// ACM Transactions on Mathematical Software, Vol. 12, No. 2, June 1986.
	// The comment explaining the algorithm in its original  format:
	/*
	 * We are calculating e^x = e^(r*10^t) = (e^r)^(10^t), where abs(r) < 1 and t >=
	 * 0.
	 *
	 * If t > 0, we have:
	 *
	 *   (1) 0.1 <= r < 1, so e^0.1 <= e^r. If t > MAX_T, overflow occurs:
	 *
	 *     MAX-EMAX+1 < log10(e^(0.1*10*t)) <= log10(e^(r*10^t)) <
	 * adjexp(e^(r*10^t))+1
	 *
	 *   (2) -1 < r <= -0.1, so e^r <= e^-0.1. If t > MAX_T, underflow occurs:
	 *
	 *     adjexp(e^(r*10^t)) <= log10(e^(r*10^t)) <= log10(e^(-0.1*10^t)) <
	 * MIN-ETINY
	 */
	t := x.Precision() - x.Scale()
	const expMax = 19
	if t > expMax {
		if x.Signbit() {
			z.Context.Conditions |= decimal.Subnormal |
				decimal.Underflow |
				decimal.Clamped |
				decimal.Inexact |
				decimal.Rounded
			return z.SetMantScale(0, -etiny(z))
		}
		z.Context.Conditions |= decimal.Overflow | decimal.Inexact | decimal.Rounded
		return z.SetInf(false)
	}

	prec := precision(z)

	// |x| <= 9 * 10 ** -(prec + 1)
	lim := alias(z, x).SetMantScale(9, prec-1)
	if x.CmpAbs(lim) <= 0 {
		z.Context.Conditions |= decimal.Rounded | decimal.Inexact
		return z.SetMantScale(1, 0).Quantize(prec - 1)
	}

	if x.IsInt() && !x.IsBig() {
		switch x.Uint64() {
		case 1:
			// e ** 1 = e
			return E(z)
		case 2:
			// e ** 2 = e * e
			e := E(z)
			return z.Mul(e, e)
		}
	}

	// TODO(eric): the +8 might not be enough extra precision for very large
	// precisions...
	prec += t + 8
	g := expg{
		prec: prec,
		z:    x,
		pow:  decimal.WithPrecision(decimal.UnlimitedPrecision).Mul(x, x),
		t: Term{
			A: decimal.WithPrecision(prec),
			B: decimal.WithPrecision(prec),
		},
	}
	Lentz(z, &g)
	return z
}

// expg is a Generator that computes exp(z).
type expg struct {
	prec int          // Precision
	z    *decimal.Big // Input value
	pow  *decimal.Big // z*z
	m    int64        // Term number
	t    Term         // Term storage. Does not need to be manually set.
}

func (e *expg) mode() decimal.RoundingMode { return decimal.ToNearestEven }

func (e *expg) Next() bool { return true }

func (e *expg) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	f = decimal.WithPrecision(e.prec)
	Δ = decimal.WithPrecision(e.prec)
	C = decimal.WithPrecision(e.prec)
	D = decimal.WithPrecision(e.prec)
	eps = decimal.New(1, e.prec)
	return
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
		e.t.A.SetMantScale(0, 0)
		e.t.B.SetMantScale(1, 0)
	// [2z, 2-z]
	case 1:
		e.t.A.Mul(two, e.z)
		e.t.B.Sub(two, e.z)
	// [z^2/6, 1]
	case 2:
		e.t.A.Quo(e.pow, six)
		e.t.B.SetMantScale(1, 0)
	// [(1/(16((m-1)^2)-4))(z^2), 1]
	default:
		// maxM is the largest m value we can use to compute 4(2m - 3)(2m - 1)
		// using integers.
		const maxM = 759250125

		// 4(2m - 3)(2m - 1) ≡ 16(m - 1)^2 - 4
		if e.m <= maxM {
			e.t.A.SetMantScale(16*((e.m-1)*(e.m-1))-4, 0)
		} else {
			e.t.A.SetMantScale(e.m-1, 0)

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

var _ Lentzer = (*expg)(nil)
