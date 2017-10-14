package math

import (
	"fmt"

	"github.com/ericlagergren/decimal"
)

// Term is a specific term in a continued fraction. A and B correspond with the
// a and b variables of the typical representation of a continued fraction. An
// example can be seen in the book, "Numerical Recipes in C: The Art of
// Scientific Computing" (ISBN 0-521-43105-5) in figure 5.2.1 on page 169.
type Term struct {
	A, B *decimal.Big
}

func (t Term) String() string {
	return fmt.Sprintf("[%s / %s]", t.A, t.B)
}

// Generator represents a continued fraction.
type Generator interface {
	// Next returns true if there are future terms. Every call to Term, even the
	// first one, must be preceded by a call to Next. In general, Generators
	// should always return true unless an exceptional condition occurs.
	Next() bool

	// Term returns the next term in the fraction. The caller must not modify
	// any of the Term's fields.
	Term() Term
}

// Lentzer, if implemented, allows Generators to provide their own backing
// storage for the Lentz function.
type Lentzer interface {
	// Lentz provides the backing storage for a Generator.
	//
	// C and D should have large enough precision to provide a correct result.
	// (See note for the  Lentz function.)
	//
	// eps should be a sufficiently small decimal, likely 1e-15 or smaller.
	//
	// For more information, refer to "Numerical Recipes in C: The Art of
	// Scientific Computing" (ISBN 0-521-43105-5), pg 171.
	Lentz() (f, Δ, C, D, eps *decimal.Big)
}

// lentzer implements the Lentzer interface.
type lentzer struct{ prec int32 }

func (l lentzer) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	f = new(decimal.Big)
	Δ = new(decimal.Big)
	C = new(decimal.Big)
	D = new(decimal.Big)
	eps = decimal.New(1, l.prec-1)

	C.Context.SetPrecision(l.prec)
	D.Context.SetPrecision(l.prec)
	return f, Δ, C, D, eps
}

var (
	defaultLentzer = lentzer{}
	tiny           = decimal.New(10, 60)
)

// Lentz sets z to the result of the continued fraction provided by the
// Generator and returns z. The continued fraction should be represented as such:
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
// Or, equivalently:
//
//                  a1   a2   a3
//     f(x) = b0 + ---- ---- ----
//                  b1 + b2 + b3 + ···
//
// If terms need to be subtracted, aN should be negative. To compute a continued
// fraction without b0, divide the result by a1.
//
// Note: the accuracy of the result may be affected by the precision of
// intermedite results. If larger precision is desired it may be necessary for
// the Generator to implement the Lentzer interface and set a higher precision
// for C and D.
func Lentz(z *decimal.Big, g Generator) *decimal.Big {
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

	// See if our Generator provides us with backing storage.
	lz, ok := g.(Lentzer)
	if !ok {
		lz = lentzer{prec: z.Context.Precision()}
	}

	f, Δ, C, D, eps := lz.Lentz()

	if !g.Next() {
		return f
	}
	t := g.Term()

	f.Set(t.B)
	if f.Sign() == 0 {
		f.Set(tiny)
	}
	C.Set(f)
	D.SetMantScale(0, 0)

	//fmt.Println(" = f:", f)
	for i := 0; g.Next(); i++ {
		t = g.Term()

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

		// if i%1000 == 0 {
		// 	fmt.Println(i)
		// }

		//f0 := new(decimal.Big).Copy(f).Round(prec)
		//Δ0 := new(decimal.Big).Copy(Δ)
		//fmt.Println(" = f:", f0)
		//fmt.Println("Δ1 :", Δ0)
		//fmt.Printf("Δ2 : %f\n", Δ0.Sub(Δ0, one))
		//fmt.Printf("eps: %f\n", eps)
		//fmt.Println()

		// If |Δ_j - 1| < eps then exit
		if Δ.Sub(Δ, one).Abs(Δ).Cmp(eps) < 0 {
			break
		}

		//dump(f, Δ, D, C, eps)
	}

	return z.Set(f)
}

func dump(f, Δ, D, C, eps *decimal.Big) {
	fmt.Printf(`
f  : %s
Δ  : %s
D  : %s
C  : %s
eps: %s
`, f, Δ, D, C, eps)
}
