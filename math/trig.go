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

type tang struct {
	t Term // Initialize A with -z*z, B with -1.
}

func (t *tang) Next() Term {
	// tan(z) can be represented as
	//
	//         ∞
	//     z / K   (-(z^2) / (2m - 1)), z ∈ ℂ
	//         m=1
	//
	t.t.B.Add(t.t.B, two)
	return t.t
}

// Tan sets z to the tangent of the radian argument x.
func Tan(z, x *decimal.Big) *decimal.Big {
	nx := new(decimal.Big).Neg(x)
	g := tang{t: Term{A: new(decimal.Big).Mul(nx, x), B: decimal.New(-1, 0)}}
	return z.Quo(x, Lentz(&g, z.Context().Precision()))
}

// Cot sets z to the cotangent of the radian argument x.
func Cot(z, x *decimal.Big) *decimal.Big {
	s := Sin(new(decimal.Big), x)
	return z.Quo(Cos(z, x), s)
}

// Sec sets z to the secand of the radian argument x.
func Sec(z, x *decimal.Big) *decimal.Big {
	return z.Quo(one, Cos(z, x))
}

// Csc sets z to the cosecant of the radian argument x.
func Csc(z, x *decimal.Big) *decimal.Big {
	return z.Quo(one, Sin(z, x))
}
