// +build !amd64

package arith

import "math/big"

// Mul128 returns the 128-bit multiplication of x and y.
func Mul128(x, y uint64) (hi, lo uint64) {
	const mask = 0xFFFFFFFF

	x1 := x & mask
	y1 := y & mask
	t := x1 * y1
	w3 := t & mask
	k := t >> 32

	x >>= 32
	t = (x * y1) + k
	k = (t & mask)
	w1 := t >> 32

	y >>= 32
	t = (x1 * y) + k

	hi = (x * y) + w1 + (t >> 32)
	lo = (t << 32) + w3
	return hi, lo
}

func alias(z, x *big.Int) *big.Int {
	if z != x {
		// We have to check the first element of their internal slices since
		// Big doesn't store a pointer to a big.Int.
		zb, xb := z.Bits(), x.Bits()
		if cap(zb) > 0 && cap(xb) > 0 && &zb[0:cap(zb)][cap(zb)-1] != &xb[0:cap(xb)][cap(xb)-1] {
			return z
		}
	}
	return new(big.Int)
}

// Add sets z to x + y and returns z.
func Add(z, x *big.Int, y uint64) *big.Int {
	return z.Add(x, alias(z, x).SetUint64(y))
}

// Sub sets z to x - y and returns z.
func Sub(z, x *big.Int, y uint64) *big.Int {
	return z.Sub(x, alias(z, x).SetUint64(y))
}

// MulUint64 sets z to x * y and returns z.
func MulUint64(z, x *big.Int, y uint64) *big.Int {
	if y == 0 || x.Sign() == 0 {
		return z.SetUint64(0)
	}
	return z.Mul(x, alias(z, x).SetUint64(y))
}
