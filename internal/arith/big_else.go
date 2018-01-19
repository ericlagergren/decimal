// +build !amd64

package arith

import "math/big"

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

// Add128 sets z to x + y and returns z.
func Add128(z *big.Int, x, y uint64) *big.Int {
	return z.Add(z.SetUint64(x), new(big.Int).SetUint64(y))
}

// Sub128 sets z to x - y and returns z.
func Sub128(z *big.Int, x, y uint64) *big.Int {
	return z.Sub(z.SetUint64(x), new(big.Int).SetUint64(y))
}

// Mul128 sets z to x * y and returns z.
func Mul128(z *big.Int, x, y uint64) *big.Int {
	if x == 0 || y == 0 {
		return z.SetUint64(0)
	}
	return z.Mul(z.SetUint64(x), new(big.Int).SetUint64(y))
}

// MulUint64 sets z to x * y and returns z.
func MulUint64(z, x *big.Int, y uint64) *big.Int {
	if y == 0 || x.Sign() == 0 {
		return z.SetUint64(0)
	}
	return z.Mul(x, alias(z, x).SetUint64(y))
}
