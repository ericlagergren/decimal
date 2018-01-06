// Package checked implements basic checked arithmetic.
package checked

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
)

// Add returns x + y and a bool indicating whether the addition was successful.
func Add(x, y uint64) (sum uint64, ok bool) {
	sum = x + y
	return sum, sum > x
}

// Mul returns x * y and a bool indicating whether the multiplication was
// successful.
func Mul(x, y uint64) (prod uint64, ok bool) {
	// Multiplication routine is from https://stackoverflow.com/a/26320664/2967113
	const (
		halfbits = 64 / 2
		halfmax  = 1<<halfbits - 1
	)

	xhi := x >> halfbits
	xlo := x & halfmax
	yhi := y >> halfbits
	ylo := y & halfmax

	low := xlo * ylo
	if xhi == 0 && yhi == 0 {
		return low, true
	}

	m0 := xlo * yhi
	m1 := xhi * ylo
	prod = low + (m0+m1)<<halfbits
	ovf := (xhi != 0 && yhi != 0) || prod < low || m0>>halfbits != 0 || m1>>halfbits != 0
	return prod, !ovf
}

// Sub returns x - y and a bool indicating whether the subtraction was successful.
func Sub(x, y uint64) (diff uint64, ok bool) {
	diff = x - y
	return diff, x >= y
}

// MulPow10 computes x * 10**n and a bool indicating whether the multiplcation
// was successful.
func MulPow10(x uint64, n uint64) (p uint64, ok bool) {
	if x == 0 {
		return 0, true
	}
	if p, ok = arith.Pow10(n); !ok {
		return 0, false
	}
	return Mul(x, p)
}

// MulBigPow10 sets z to x * 10**n and returns z.
func MulBigPow10(z, x *big.Int, n uint64) *big.Int {
	if x.Sign() == 0 {
		return z.SetUint64(0)
	}
	return z.Mul(x, arith.BigPow10(n))
}
