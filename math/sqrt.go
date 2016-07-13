package math

import (
	"math"
	"math/big"

	"github.com/EricLagergren/decimal"
	"github.com/EricLagergren/decimal/internal/arith"
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

// Sqrt sets z to the square root of x and returns z.
// The precision of Sqrt is determined by z's Context.
// Sqrt will panic on negative values since decimal.Big cannot
// represent imaginary numbers.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	if x.SignBit() {
		panic("math.Sqrt: cannot take square root of negative number")
	}

	switch {
	case x.IsNaN():
		panic(ErrNaN{"square root of NaN"})
	case x.IsInf():
		return z.SetInf()
	case x.Sign() == 0:
		return z.SetMantScale(0, 0)
	}

	// First fast path---check if x is a perfect square. If it is, we can avoid
	// having to inflate x and can possibly use can use the hardware SQRT.
	// Note that we can only catch perfect squares that aren't big.Ints.
	if sq, ok := perfectSquare(x); ok {
		return z.SetMantScale(sq, 0).SetContext(x.Context())
	}

	zp := z.Context().Precision()

	// Temporary inflation. Should be enough to accurately determine the sqrt
	// with at least zp digits after the radix.
	zpadj := int(zp) << 1

	var tmp *decimal.Big
	if z != x {
		zctx := z.Context()
		tmp = z.Set(x)
		tmp.SetContext(zctx)
	} else {
		tmp = new(decimal.Big).Set(x)
	}
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
	if h != 2 && h != 3 && h != 5 && h != 6 && h != 7 && h != 8 {
		// "Show that floating point sqrt(x*x) >= x for all long x."
		// https://math.stackexchange.com/a/238885/153292
		tst := int64(math.Sqrt(float64(xc)))
		return tst, tst*tst == xc
	}
	return 0, false
}
