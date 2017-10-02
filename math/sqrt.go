package math

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ericlagergren/checked"
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
)

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	p0 := new(decimal.Big).Set(p)
	q0 := new(decimal.Big).Set(q)
	if p0.Sign() <= 0 {
		p0.Neg(p0)
	}
	if q0.Sign() <= 0 {
		q0.Neg(q0)
	}
	if p0.Sign() == 0 {
		return z.SetMantScale(0, 0)
	}
	p0.Mul(p0, p0)
	q0.Mul(q0, q0)
	return Sqrt(z, p0.Add(p0, q0))
}

// Sqrt sets z to the square root of x and returns z. Sqrt will panic on
// negative values since decimal.Big cannot represent imaginary numbers.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	switch {
	case x.Signbit():
		panic("math.Sqrt: cannot take square root of negative number")
	case x.IsNaN(true), x.IsNaN(false):
		panic(decimal.ErrNaN{"square root of NaN"})
	case x.IsInf(0):
		return z.SetInf(false)
	case x.Sign() == 0:
		return z.SetMantScale(0, 0)
	}

	// First fast path—check if x is a perfect square. If so, we can avoid
	// having to inflate x and can possibly use can use the hardware SQRT.
	// Note that we can only catch perfect squares that aren't big.Ints.
	if sq, ok := perfectSquare(x); ok {
		z.SetMantScale(sq, 0)
		z.Context = x.Context
		return z
	}

	zp := z.Context.Precision()

	// Temporary inflation. Should be enough to accurately determine the sqrt
	// with at least zp digits after the radix.
	zpadj := int(zp) << 1

	tmp := alias(z, x).Set(x)
	shiftRadixRight(tmp, zpadj)

	// Second fast path. Check to see if we can calculate the square root without
	// using big.Int
	if !x.IsBig() && zpadj <= 19 {
		n := tmp.Int64()
		ix := n >> uint((arith.BitLen(n)+1)>>1)
		var p int64
		for {
			p = ix
			ix += n / ix
			ix >>= 1
			if ix == p {
				return z.SetMantScale(ix, zp)
			}
		}
	}

	// x isn't a perfect square or x is a big.Int

	n := tmp.Int()
	ix := new(big.Int).Rsh(n, uint((n.BitLen()+1)>>1))

	var a, p big.Int
	for {
		p.Set(ix)
		ix.Add(ix, a.Quo(n, ix)).Rsh(ix, 1)
		if ix.Cmp(&p) == 0 {
			fmt.Println(ix, zp)
			return z.SetBigMantScale(ix, zp)
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
	switch h {
	case 2, 3, 5, 6, 7, 8:
		return 0, false
	default:
		// "Show that floating point sqrt(x*x) >= x for all long x."
		// https://math.stackexchange.com/a/238885/153292
		tst := int64(math.Sqrt(float64(xc)))
		return tst, tst*tst == xc
	}
}

func shiftRadixRight(x *decimal.Big, n int) {
	ns, ok := checked.Sub32(x.Scale(), int32(n))
	if !ok {
		panic(ok)
	}
	x.SetScale(ns)
}
