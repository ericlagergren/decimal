package math

import (
	"fmt"
	"math"

	"github.com/ericlagergren/decimal"
)

// Term is a specific term in a continued fraction. A and B correspond with
// the a and b variables of the typical representation of a continued fraction.
// An example can be seen in the book, "Numerical Recipes in C: The Art of
// Scientific Computing" (ISBN 0-521-43105-5) in figure 5.2.1 found on page
// 169.
type Term struct {
	A, B *decimal.Big
}

func (t Term) String() string {
	return fmt.Sprintf("[%s / %s]", t.A, t.B)
}

// Generator represents a continued fraction.
type Generator interface {
	// Next returns the next term in the fraction. For efficiency's sake, the
	// caller must not modify any of the Term's fields.
	Next() Term
}

// Lentzer, if implemented, will allow Generators to provide their own backing
// storage for the Lentz function. f will be the value returned from Lentz.
type Lentzer interface {
	Lentz() (f, Δ, C, D, eps *decimal.Big)
}

// lentzer implements the Lentzer interface.
type lentzer struct{}

func (l lentzer) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	return new(decimal.Big), // f
		new(decimal.Big), // Δ
		new(decimal.Big), // C
		new(decimal.Big), // D
		decimal.New(1, 15) // 1e-15
}

var defaultLentzer = lentzer{}

var tiny = decimal.New(10, 30)

// Lentz computes the continued fraction provided by the Generator using the
// modified Lentz algorithm to prec precision. The continued fraction should be
// represented as such:
//
//                          a1
//     f(x) = b0 + --------------------
//                            a2
//                 b1 + ---------------
//                               a3
//                      b2 + ----------
//                                 a4
//                           b3 + -----
//                                  ...
//
// If terms need to be subtracted, aN should be negative. To compute a
// continued fraction without b0, the Generator should be offset and begin with
// a2, b1 and the return value from Lentz should be divided by a1.
//
// Lentz will panic after 1<<63 - 2 terms.
func Lentz(g Generator, prec int32) *decimal.Big {
	// Lentz differs from other functions whose signatures typically mirror
	//
	//     func F(z, x *T) *T
	//
	// because it checks to see if the Generator implements the Lentzer
	// interface, and if so uses f as the backing storage.

	// We use the modified Lentz algorithm from
	// "Numerical Recipes in C: The Art of Scientific Computing" (ISBN
	// 0-521-43105-5), pg 171.
	//
	// Set f0 = b0; if b0 = 0 set f0 = tiny.
	// Set C0 = f0.
	// Set D0 = 0.
	// For j = 1, 2,...
	// 		Set D_j = b_j+a_j*D{_j−1}.
	// 		If D_j = 0, set D_j = tiny.
	// 		Set C_j = b_j+a_j/C{_j−1}.
	// 		If C_j = 0, set C_j = tiny.
	// 		Set D_j = 1/D_j.
	// 		Set ∆_j = C_j*D_j.
	// 		Set f_j = f{_j-1}∆j.
	// 		If |∆_j - 1| < eps then exit.
	//
	// Instead of comparing Δ to eps, we compare f_j to the f{_j-1} to see if
	// the two terms have converged.
	t := g.Next()

	// See if our Generator provides us with backing storage.
	lz, ok := g.(Lentzer)
	if !ok {
		lz = defaultLentzer
	}

	f, Δ, C, D, eps := lz.Lentz()
	f.Set(t.B)
	if f.Sign() == 0 {
		f.Set(tiny)
	}
	C.Set(f)
	D.SetMantScale(0, 0)

	// TODO: is there a better cutoff?
	for i := 0; i < math.MaxInt64; i++ {
		t = g.Next()

		// Set D_j = b_j + a_j*D{_j-1}
		// Reuse D for the multiplication.
		D.Add(t.B, D.Mul(t.A, D))

		// If D_j = 0, set D_j = tiny
		if D.Sign() == 0 {
			D.Set(tiny)
		}

		// Set C_j = b_j + a_j/C{_j-1}
		// Reuse C for the division.
		C.Add(t.B, C.Quo(t.A, C))

		// If C_j = 0, set C_j = tiny
		if C.Sign() == 0 {
			C.Set(tiny)
		}

		// Set D_j = 1/D_j
		D.Quo(one, D)

		// Set Δ_j = C_j*D_j
		Δ.Mul(C, D)

		// Set f_j = f{_j-1}*Δ_j
		f.Mul(f, Δ)

		fmt.Println(f)

		// If |Δ_j - 1| < eps then exit
		if Δ.Sub(Δ, one).Abs(Δ).Cmp(eps) < 0 {
			fmt.Println("iters", i, f)
			return f.Round(prec)
		}

		//dump(f, Δ, D, C, eps)
	}
	panic("Lentz: too many iterations")
}

func dump(f, Δ, D, C, eps *decimal.Big) {
	fmt.Printf("f: %s, Δ: %s, D: %s, C: %s, eps: %s\n",
		f, Δ, D, C, eps)
}
