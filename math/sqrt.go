package math

import (
	"math"
	"math/big"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith/checked"
)

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	// ±Inf*±Inf + q*q
	// p*p + ±Inf*±Inf
	if p.IsInf(0) || q.IsInf(0) {
		return z.SetInf(true)
	}

	p0 := alias(z, p).Set(p)
	q0 := new(decimal.Big).Set(q)

	// Two routes: either we can compute hypot(p, q) the safe way or we can
	// simply compute sqrt(p*p, q*q). Try the latter first.

	_, okp := checked.Add32(p.Scale(), p.Scale())
	_, okq := checked.Add32(q0.Scale(), q0.Scale())
	if okp && okq {
		p0.Mul(p0, p0)
		q0.Mul(q0, q0)
		// Store p*p+q*q in p0
		p0.Add(p0, q0)
		return Sqrt(z, p0)
	}

	if p0.Sign() <= 0 {
		p0.Neg(p0)
	}
	if q0.Sign() <= 0 {
		q0.Neg(q0)
	}
	if p0.Cmp(q0) < 0 {
		p0, q0 = q0, p0
	}
	if p0.Sign() == 0 {
		return z.SetMantScale(0, 0)
	}
	zp := z.Context().Precision()
	q0.SetPrecision(zp)
	q0.Quo(q0, p0)
	q0.Mul(q0, q0)
	q0.Add(q0, one)
	return z.Mul(p0, Sqrt(q0, q0)).Round(zp)
}

// Sqrt sets z to the square root of x and returns z. Sqrt will panic on
// negative values since Big cannot represent imaginary numbers.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	if sgn := x.Sign(); sgn < 1 {
		// sqrt(0) == 0
		// sqrt({ N < 0 }) = undefined, but still give z a sane value.
		z.SetMantScale(0, 0)
		if sgn < 0 {
			panic(decimal.ErrNaN{"square root of negative number"})
		}
		return z
	}

	// sqrt(+Inf) == +Inf
	if x.IsInf(0) {
		return z.SetInf(true)
	}

	// Tests are on macOS with an 2.9 GHz Intel Core i5, Go 1.7.3

	// First fast path. Check if x is a perfect square. If it is, we can avoid
	// having to inflate x and can just use the hardware SQRT. Note that we can
	// only catch perfect squares that aren't big.Ints.
	//
	// Tests show this path takes ~50 ns/op.
	if sq, ok := perfectSquare(x); ok {
		return z.SetMantScale(sq, 0)
	}

	zp := z.Context().Precision()

	// Temporary inflation. Should be enough to accurately determine the sqrt
	// with zp precision.
	zpadj := int(zp) << 1

	tmp := alias(z, x).Copy(x)

	if !shiftRadixRight(tmp, zpadj) {
		return z.SetInf(tmp.Signbit())
	}

	// Second fast path. We had to inflate x, so whether it's a perfect square
	// is irrelevant to us now. Since sqrt(x*x) >= x for all 64-bit x (see the
	// perfectSquare routine for proof), check to see if x is <= 64 bits and
	// use the hardware SQRT.
	//
	// Tests show this path takes ~170 ns/op.
	if !tmp.IsBig() && zpadj <= 19 {
		sqrt := int64(math.Sqrt(float64(tmp.Int64())))
		scl := (zpadj + 1) / 2
		if zpadj < 0 || zpadj%2 != 0 {
			scl--
		}
		return z.SetMantScale(sqrt, int32(scl)).Round(zp)
	}

	// General case. Use Newton's method with big.Int.
	//
	// Tests show this path takes ~4,000 ns/op.

	n := tmp.Int()
	ix := new(big.Int).Rsh(n, uint((n.BitLen()+1)>>1))

	var a, p big.Int
	for {
		p.Set(ix)
		ix.Add(ix, a.Quo(n, ix)).Rsh(ix, 1)
		if ix.Cmp(&p) == 0 {
			return z.SetBigMantScale(ix, zp).Round(zp)
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
	case 0, 1, 4, 9:
		// "Show that floating point sqrt(x*x) >= x for all long x."
		// https://math.stackexchange.com/a/238885/153292
		tst := int64(math.Sqrt(float64(xc)))
		return tst, tst*tst == xc
	default:
		return 0, false
	}
}

func shiftRadixRight(x *decimal.Big, n int) bool {
	ns, ok := checked.Sub32(x.Scale(), int32(n))
	if !ok {
		return false
	}
	x.SetScale(ns)
	return true
}
