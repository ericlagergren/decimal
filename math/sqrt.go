package math

import (
	"fmt"
	"math"
	"math/big"

	"github.com/EricLagergren/decimal"
	"github.com/EricLagergren/decimal/internal/arith/checked"
)

var _ = fmt.Println

func shiftRadixRight(x *decimal.Big, n int) {
	ns, ok := checked.Sub32(x.Scale(), int32(n))
	if !ok {
		panic(ok)
	}
	x.SetScale(ns)
}

// Sqrt sets z to the square root of x and returns z.
// The precision of Sqrt is determined by z's Context.
// Sqrt will panic on negative values since decimal.Big cannot
// represent imaginary numbers.
func Sqrt(z *decimal.Big, x *decimal.Big) *decimal.Big {
	if x.SignBit() {
		panic("math.Sqrt: cannot take square root of negative number")
	}

	// Check if x is a perfect square. If it is, we can avoid having to
	// inflate x and can possibly use can use the hardware SQRT.
	// Note that we can only catch perfect squares that aren't big.Ints.
	if sq, ok := perfectSquare(x); ok {
		return z.SetMantScale(sq, 0).SetContext(x.Context())
	}
	// x isn't a perfect square or x is a big.Int

	tmp := new(decimal.Big).Set(x)
	shiftRadixRight(tmp, int(z.Context().Precision()<<1))
	n := tmp.Int()
	ix := new(big.Int).Rsh(n, uint((n.BitLen()+1)>>1))

	var a, p big.Int
	for {
		p.Set(ix)
		ix.Add(ix, a.Quo(n, ix)).Rsh(ix, 1)
		if ix.Cmp(&p) == 0 {
			return z.SetBigMantScale(ix, z.Context().Precision())
		}
	}
}

// perfectSquare algorithm slightly partially borrowed from
// https://stackoverflow.com/a/295678/2967113
func perfectSquare(x *decimal.Big) (square int64, ok bool) {
	if x.IsBig() || !x.IsInt() {
		return 0, false
	}
	xc := x.Int64()
	h := xc & 0xF
	if h > 9 {
		return 0, false
	}
	if h != 2 && h != 3 && h != 5 && h != 6 && h != 7 && h != 8 {
		// "Show that floating point sqrt(x*x) >= x for all long x."
		// https://math.stackexchange.com/a/238885/153292
		tst := int64(math.Sqrt(float64(xc)))
		return tst, tst*tst == xc
	}
	return 0, false
}
