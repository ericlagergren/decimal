// Package checked implements basic checked arithmetic.
package checked

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
)

// Add returns x + y and a bool indicating whether the addition was successful.
func Add(x, y uint64) (sum uint64, ok bool) {
	sum = x + y
	return sum, sum >= x
}

// Sub returns x - y and a bool indicating whether the subtraction was successful.
func Sub(x, y uint64) (diff uint64, ok bool) {
	diff = x - y
	return diff, x >= y
}

func Mul(x, y uint64) (prod uint64, ok bool) {
	hi, lo := arith.Mul128(x, y)
	return lo, hi == 0
}

// MulPow10 computes x * 10**n and a bool indicating whether the multiplcation
// was successful.
func MulPow10(x uint64, n uint64) (uint64, bool) {
	p, ok := arith.Pow10(n)
	if !ok {
		return 0, x == 0
	}
	z1, z0 := arith.Mul128(x, p)
	return z0, z1 == 0
}

// MulBigPow10 sets z to x * 10**n and returns z.
func MulBigPow10(z, x *big.Int, n uint64) *big.Int {
	if x.Sign() == 0 {
		return z.SetUint64(0)
	}
	return z.Mul(x, arith.BigPow10(n))
}
