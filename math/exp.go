package math

import (
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/misc"
)

// expg is a Generator that computes exp(z).
type expg struct {
	prec int          // Precision
	z    *decimal.Big // Input value
	pow  *decimal.Big // z*z
	m    int64        // Term number
	t    Term         // Term storage. Does not need to be manually set.
}

var P int = 16

func (e *expg) Next() bool { return true }

func (e *expg) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	f = decimal.WithPrecision(500)
	Δ = decimal.WithPrecision(500)
	C = decimal.WithPrecision(500)
	D = decimal.WithPrecision(500)
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
	//          2z     z^2 / 6    ∞
	//     1 + -----  ---------   K ((a_m^z^2) / 1), z ∈ ℂ
	//          2-z  +    1     + m=3
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

// Exp sets z to e ** x and returns z.
func Exp(z, x *decimal.Big) *decimal.Big {
	// TODO(eric): "pestle_: eric_lagergren, that is, exp(z+z0) = exp(z)*exp(z0)
	// 						 and exp(z) ~ 1+z for small enough z"

	if z.CheckNaNs(x, nil) {
		return z
	}

	if x.IsInf(0) {
		// e ** +Inf = +Inf
		if x.IsInf(+1) {
			return z.SetInf(true)
		}
		// e ** -Inf = 0
		return z.SetMantScale(0, 0)
	}

	if x.Signbit() {
		// e ** -x = 1 / (e ** x)
		return z.Quo(one, Exp(z, misc.CopyAbs(z, x)))
	}

	if x.IsInt() {
		switch x.Uint64() {
		case 0:
			// e ** 0 = 1
			return z.SetMantScale(1, 0)
		case 1:
			// e ** 1 = e
			return E(z)
		// TODO(eric): should we handle case 3?
		case 2:
			// e ** 2 = e * e
			e := E(z)
			return z.Mul(e, e)
		}
	}

	t := Term{
		A: decimal.WithPrecision(500),
		B: decimal.WithPrecision(500),
	}
	g := expg{
		prec: precision(z),
		z:    x,
		pow:  decimal.WithPrecision(decimal.UnlimitedPrecision).Mul(x, x),
		t:    t,
	}
	return Lentz(z, &g)
}

var _ Lentzer = (*expg)(nil)
