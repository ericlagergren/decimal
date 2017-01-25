package math

import "github.com/ericlagergren/decimal"

// newSincos creates a sincos generator calculating either sin(x) or cos(x),
// using z as its backing storage.
func newSincos(z, x *decimal.Big, cos bool) *sincos {
	g := sincos{
		recv: alias(z, x),
		pow:  new(decimal.Big).Mul(x, x),
		m:    -1,
		t: Term{
			A: new(decimal.Big).SetPrecision(55),
			B: new(decimal.Big),
		},
	}
	if cos {
		g.c = -2
	} else {
		g.c = +2
	}
	return &g
}

type sincos struct {
	recv *decimal.Big // receiver
	pow  *decimal.Big // z*z
	m    int64        // m
	t    Term         // Current term
	c    int64        // c

	// Lazily initialized inside Next only if m >= maxM
	big bool         // true if the following are initialized
	mv  *decimal.Big // Decimal form of m
	cv  *decimal.Big // Decimal form of c
}

func (s *sincos) Lentz() (f, pf, Δ, C, D *decimal.Big) {
	return s.recv, // f
		new(decimal.Big), // pf
		new(decimal.Big), // Δ
		new(decimal.Big).SetPrecision(16), // C
		new(decimal.Big).SetPrecision(500) // D
}

func (s *sincos) Next() Term {
	s.m++

	// sin(z) can be represented as:
	//
	//              z
	//    --------------------
	//        ∞    /   a_m   \
	//    1 + K   | --------- |, z ∈ ℂ
	//        m=1  \ 1 - a_m /
	//
	// where
	//
	//    a_m = 2m(2m + c)
	//
	// where c is either the constant 1 for sin(z) or -1 for cos(z).
	//
	// cos(z) can be represented as the above function, except instead of
	// dividing by z, the continued fraction is divided by 1.
	//
	// In both cases, a_m can be simplified to
	//
	//    a_m = m(4m + 2c).
	//
	// Our data structure computes 2c in the beginning instead of each
	// invocation of Next.

	// First term: [0, 1]
	if s.m == 0 {
		return Term{A: zero, B: one}
	}

	// Subsequent terms: [z^2 / (m(4m + 2c)), 1 - (z^2 (m(4m + 2c)))]

	// maxM is the largest a_m we're able to compute with integer
	// multiplication. We can actually add 2 when computing cosine, but that
	// doesn't allow us to use a constant and wouldn't make much of a
	// difference anyway.
	const maxM = 1518500249

	// Calculate a_m. If we can, use integer multiplication.
	if s.m <= maxM {
		// m(4m + 2c) can be computed with integers if s.m < math.MaxInt64.
		s.t.A.SetMantScale(s.m*((4*s.m)+s.c /* is actually 2c */), 0)
	} else {
		if !s.big {
			s.mv = decimal.New(s.m, 0)
			s.cv = decimal.New(s.c /* is actually 2c */, 0)
			s.big = true
		} else {
			s.mv.SetMantScale(s.m, 0)
		}

		// 4m
		s.t.A.Mul(four, s.mv)
		// 4m + 2c
		s.t.A.Add(s.t.A, s.cv /* is actually 2c */)
		// m(4m + 2c)
		s.t.A.Mul(s.mv, s.t.A)
	}

	// z^2 / m(4m + 2c)
	s.t.A.Quo(s.pow, s.t.A)

	// 1 - (z^2 / m(4m + 2c))
	s.t.B.Sub(one, s.t.A)
	return s.t
}

// Sin sets z to the sine of the radian argument x.
func Sin(z, x *decimal.Big) *decimal.Big {
	K := Lentz(newSincos(z, x, false), z.Context().Precision())
	return z.Quo(x, K)
}

// Cos sets z to the cosine of the radian argument x.
func Cos(z, x *decimal.Big) *decimal.Big {
	K := Lentz(newSincos(z, x, true), z.Context().Precision())
	return z.Quo(one, K)
}

type tancotg struct {
	Term              // Initialize A with -z*z, B with -1.
	two  *decimal.Big // ±2
}

func (t *tancotg) Next() Term {
	// tan(z) can be represented as
	//
	//         ∞
	//     z / K   (-(z^2) / (2m - 1)), z ∈ ℂ
	//         m=1
	//
	// cot(z) can be represented as
	//
	//            ∞
	//            K   (-(z^2) / 1+2m), z/π ∉ ℤ
	//      1     m=1
	//     --- + -------------------
	//      z             z
	//

	t.B.Add(t.B, t.two)
	return t.Term
}

// Tan sets z to the tangent of the radian argument x.
func Tan(z, x *decimal.Big) *decimal.Big {
	nx := new(decimal.Big).Neg(x)
	g := tancotg{Term{A: new(decimal.Big).Mul(nx, x), B: decimal.New(-1, 0)}}
	return z.Quo(x, Lentz(&g, z.Context().Precision()))
}

// Cot sets z to the cotangent of the radian argument x.
func Cot(z, x *decimal.Big) *decimal.Big {
	g := tancotg{}
	K := Lentz(&g, z.Context().Precision())
	// A little more work than normal since
	//
	//    cot(z) = 1/z + K/z
	//
	// where K is the converged fraction.

	// 1 / z
	z.Quo(one, x)

	// K / z
	Kv := new(decimal.Big).Quo(K, x)

	// 1/2 + K/z
	return z.Add(z, Kv)
}

// Sec sets z to the secand of the radian argument x.
func Sec(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Csc sets z to the cosecant of the radian argument x.
func Csc(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Exs sets z to the exsecant of the radian argument x.
func Exs(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Exc sets z to the exosecant of the radian argument x.
func Exc(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Ver sets z to the versed sine of the radian argument x.
func Ver(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Vcs sets z to the versed cosine of the radian argument x.
func Vcs(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Cvs sets z to the coversed sine sine of the radian argument x.
func Cvs(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Cvc sets z to the coversed cosine sine of the radian argument x.
func Cvc(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Hvs sets z to the haversed sine sine of the radian argument x.
func Hvs(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Hvc sets z to the haversed cosine of the radian argument x.
func Hvc(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Hcv sets z to the hacoversed sine of the radian argument x.
func Hcv(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Hcc sets z to the hacoversed cosine of the radian argument x.
func Hcc(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}
