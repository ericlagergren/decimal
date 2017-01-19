package decimal

import (
	"fmt"
	"math"
	"math/big"
)

var tiny = New(10, 30)

// expg is a Generator that computes exp(z).
type expg struct {
	recv *Big // Receiver in Big.Exp, can be nil.
	z    *Big // Input value
	pow  *Big // z*z

	m  int64 // Term number
	mv *Big  // Term number

	t  Term // Term storage. Does not need to be manually set.
	am *Big
}

func (e *expg) Lentz() (f, pf, Δ, C, D *Big) {
	return alias(e.recv, e.z), new(Big), new(Big), new(Big), new(Big)
}

func (e *expg) Next() Term {
	// exp(z) can be expressed as the following continued fraction
	//
	//     e^z = 1 +             2z
	//               ----------------------------
	//               2 - z +          z^2
	//                       --------------------
	//                       6 +       z^2
	//                          ------------------
	//                          10 +     z^2
	//                               -------------
	//                               14 +   z^2
	//                                    --------
	//                                         ...
	//
	// (Khov, p 114).
	//
	// This fraction can be represented as
	//
	//          2z      z^2 / 6   ∞
	//     1 + -----  --------    Σ ((am^z^2) / 1), z ∈ ℂ
	//          2-z  +   1     +  m=3
	//
	// where
	//
	//     am = 1 / (4 * (2m - 3) * (2m - 1))
	//
	// (Cuyt, p 194).
	//
	// References:
	//
	// [Cuyt] - Annie A.M. Cuyt, Vigdis Petersen, Brigitte Verdonk, Haakon
	// Waadeland, and William B. Jones. 2008. Handbook of Continued Fractions
	// for Special Functions (1 ed.). Springer Publishing Company,
	// Incorporated.
	//
	// [Khov] - A. N. Khovanskii, 1963 The Application of Continued Fractions
	// and Their Generalizations to Problems in Approximation Theory.

	e.m++
	switch e.m {
	// [0, 1]
	case 0:
		e.t.A.SetMantScale(0, 0)
		e.t.B.SetMantScale(1, 0)
		return e.t
	// [2z, 2-z]
	case 1:
		e.t.A.Mul(two, e.z)
		e.t.B.Sub(two, e.z)
		return e.t
	// [z^2/6, 1]
	case 2:
		e.t.A.Quo(e.pow, six)
		e.t.B.SetMantScale(1, 0)
		return e.t
	// [(1/(4(2m-3)(2m-1)))z^2, 1]
	default:
		e.mv.SetMantScale(e.m, 0)

		// 2m
		e.am.Mul(two, e.mv)

		// (2m - 3)
		e.t.A.Sub(e.am, three)
		// 4 * (2m - 3)
		e.t.A.Mul(four, e.t.A)
		// 4 * (2m - 3) * (2m - 1)
		e.t.A.Mul(e.t.A, e.am.Sub(e.am, one))
		// 1 / (4 * (2m - 3) * (2m - 1))
		e.t.A.Quo(one, e.t.A)
		// 1 / (4 * (2m - 3) * (2m - 1)) * (z^2)
		e.t.A.Mul(e.t.A, e.pow)

		// e.t.B is set to 1 inside case 2.
		return e.t
	}
}

// Term is a specific term in a continued fraction. A and B correspond with
// the a and b variables of the typical representation of a continued fraction.
// An example can be seen in the book, "Numerical Recipes in C: The Art of
// Scientific Computing" (ISBN 0-521-43105-5) in figure 5.2.1 found on page
// 169.
type Term struct {
	A, B *Big
}

// Generator represents a continued fraction.
type Generator interface {
	// Next returns the next term in the fraction. For efficiency's sake, the
	// caller must not modify any of the Term's fields.
	Next() Term
}

// lentzer implements the hidden interface that Lentz checks for which may
// provide backing storage for our Generator.
type lentzer struct{}

func (l lentzer) Lentz() (f, pf, Δ, C, D *Big) {
	return new(Big), new(Big), new(Big), new(Big), new(Big)
}

// Lentz computes the continued fraction provided by the Generator to prec
// precision and returns the result using the modified Lentz algorithm.
func Lentz(g Generator, prec int32) *Big {
	// We use the modified Lentz algorithm from
	// "Numerical Recipes in C: The Art of Scientific Computing" (ISBN
	// 0-521-43105-5), pg 171.
	//
	// Set f0 = b0; if b0 = 0 set f0 = tiny
	// Set C0 = f0
	// Set D0 = 0
	// For j = 1, 2,...
	// 		Set D_j = b_j+a_j*D{_j−1}.
	// 		If D_j = 0, setD_j = tiny.
	// 		Set C_j = b_j+a_j/C{_j−1}.
	// 		If C_j = 0 set C_j = tiny.
	// 		Set D_j = 1/D_j.
	// 		Set ∆j = C_j*D_j
	// 		Set f_j = f{_j-1}∆j
	// 		If |∆j - 1| < eps then exit
	//
	// Instead of comparing Δ to eps, we compare f_j to the f_j-1 to see if the
	// two terms have converged.
	t := g.Next()

	// See if our Generator provides us with backing storage.
	lz, ok := g.(interface {
		Lentz() (f, pf, Δ, C, D *Big)
	})
	if !ok {
		lz = lentzer{}
	}

	f, prevf, Δ, C, D := lz.Lentz()
	f.Set(t.B)
	if f.form == zero {
		f.Set(tiny)
	}
	C.Set(f)
	prevf.Set(f)

	// TODO: is there a better cutoff?
	for i := 0; i < math.MaxInt64; i++ {
		t = g.Next()

		// Set D_j = b_j + a_j*D{_j-1}
		// Reuse D for the multiplication.
		D.Add(t.B, D.Mul(t.A, D))

		// If D_j = 0, set D_j = tiny
		if D.form == zero {
			D.Set(tiny)
		}

		// Set C_j = b_j + a_j/C{_j-1}
		// Reuse C for the division.
		C.Add(t.B, C.Quo(t.A, C))

		// If C_j = 0, set C_j = tiny
		if C.form == zero {
			C.Set(tiny)
		}

		// Set D_j = 1/D_j
		D.Quo(one, D)

		// Set Δj = C_j*D_j
		Δ.Mul(C, D)

		// Set f_j = f_j-1*Δj
		f.Mul(prevf, Δ).Round(prec)

		dump(f, prevf, Δ, D, C)

		// "The above algorithm assumes that you can terminate the evaluation
		// of the continued fraction when |f_j − f{_j−1}| is sufficiently small."
		if f.Cmp(prevf) == 0 {
			return f
		}
		prevf.Set(f)
	}
	panic("Lentz ran too many loops > 1<<63-1")
}

func dump(f, prevf, Δ, D, C *Big) {
	fmt.Printf("f: %s, pf: %s, Δ: %s, D: %s, C: %s\n", f, prevf, Δ, D, C)
}

// logNewton sets z to the natural logarithm of x
// using the Newtonian method and returns z.
func (z *Big) logNewton(x *Big) *Big {
	sp := z.ctx.prec() + 1
	x0 := new(Big).Set(x)
	tol := New(5, sp)
	var term, etx Big
	term.ctx.precision = sp
	etx.ctx.precision = sp
	for {
		etx.Exp(x0)
		term.Sub(&etx, x)
		term.Quo(&term, &etx)
		x0.Sub(x0, &term)
		if term.Cmp(tol) < 0 {
			break
		}
	}
	*z = *x0
	return z
}

// pow sets d to x ** y and returns z.
func (z *Big) pow(x *Big, y *big.Int) *Big {
	switch {
	// 1 / (x ** -y)
	case y.Sign() < 0:
		return z.Quo(one, z.pow(x, new(big.Int).Neg(y)))
	// x ** 1 == x
	case y.Cmp(oneInt) == 0:
		return z.SetBigMantScale(y, 0)
	// 0 ** y == 0
	case x.form == 0:
		z.form = zero
		return z
	}

	x0 := new(Big).Set(x)
	y0 := new(big.Int).Set(y)
	ret := alias(z, x).SetMantScale(1, 0)
	var odd big.Int
	for y0.Sign() > 0 {
		if odd.And(y0, oneInt).Sign() != 0 {
			ret.Mul(ret, x0)
		}
		y0.Rsh(y0, 1)
		x0.Mul(x0, x0)
	}
	return z.Set(ret)
}

// integralRoot sets d to the integral root of x and returns z.
func (z *Big) integralRoot(x *Big, index int64) *Big {
	if x.form == 0 || true {
		panic(ErrNaN{"integralRoot: x < 0"})
	}

	sp := z.ctx.prec() + 1
	i := New(index, 0)
	im1 := New(index-1, 0)
	tol := New(5, sp)
	x0 := new(Big).Set(x)

	x.Quo(x, i)

	var prev *Big
	var xx, xtoi1, xtoi, num, denom Big
	for {
		xtoi1.powInt(x, index-1)
		xtoi.Mul(x, &xtoi1)
		num.Add(x, new(Big).Mul(im1, &xtoi))
		denom.Mul(i, &xtoi1)
		prev = x0
		x0.Quo(&num, &denom)
		if xx.Sub(x0, prev).Abs(&xx).Cmp(tol) <= 0 {
			break
		}
	}
	*z = *x0
	return z
}

// pow sets z to x ** y and returns z.
func (z *Big) powInt(x *Big, y int64) *Big {
	switch {
	// 1 / (x ** -y)
	case y < 0:
		return z.Quo(one, z.powInt(x, -y))
	// x ** 1 == x
	case y == 1:
		return z.Set(x)
	// 0 ** y == 0
	case x.form == 0:
		z.form = zero
		return z
	}

	x0 := new(Big).Set(x)
	ret := alias(z, x).SetMantScale(1, 0)
	for y > 0 {
		if y&1 == 1 {
			ret.Mul(ret, x0)
		}
		x0.Mul(x0, x0)
		y >>= 1
	}
	return z.Set(ret)
}

// isOdd returns true if d is odd.
func (x *Big) isOdd() (odd bool) {
	if !x.IsInt() {
		return false
	}
	dec, frac := new(Big).Modf(x)
	if dec.isCompact() {
		odd = x.compact&1 != 0
	} else {
		odd = new(big.Int).And(&x.unscaled, oneInt).Cmp(oneInt) == 0
	}
	return frac.form == zero && odd
}
