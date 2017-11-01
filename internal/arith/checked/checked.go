// Package checked implements basic checked arithmetic.
package checked

import (
	"math"
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/c"
)

// Add returns x + y and a bool indicating whether the addition was successful.
func Add(x, y int64) (sum int64, ok bool) {
	sum = x + y
	// Algorithm from "Hacker's Delight" 2-12
	return sum, (sum^x)&(sum^y) >= 0
}

// Add32 returns x + y and a bool indicating whether the addition was
// successful.
func Add32(x, y int32) (sum int32, ok bool) {
	sum = x + y
	// Algorithm from "Hacker's Delight" 2-12
	return sum, (sum^x)&(sum^y) >= 0
}

// Mul returns x * y and a bool indicating whether the multiplication was
// successful.
func Mul(x, y int64) (prod int64, ok bool) {
	prod = x * y
	return prod, ((arith.Abs(x)|arith.Abs(y))>>31 == 0 || prod/y == x)
}

// Sub returns x - y and a bool indicating whether the subtraction was successful.
func Sub(x, y int64) (diff int64, ok bool) {
	return Add(x, -y)
}

// Sub32 returns x - y and a bool indicating whether the subtraction was
// successful.
func Sub32(x, y int32) (diff int32, ok bool) {
	return Add32(x, -y)
}

// MulPow10 computes 10 * x**n and a bool indicating whether the multiplcation
// was successful.
func MulPow10(x int64, n uint64) (p int64, ok bool) {
	if x == 0 {
		return x, true
	}
	if n >= pow.TabLen-1 || x == c.Inflated {
		return 0, false
	}
	up, ok := pow.Ten(n)
	if !ok {
		return 0, false
	}
	if x == 1 {
		return int64(up), true
	}
	return Mul(x, int64(up))
}

// MulBigPow10 computes 10 * x**n. It reuses x.
func MulBigPow10(x *big.Int, n uint64) *big.Int {
	if x.Sign() == 0 {
		return x
	}
	return x.Mul(x, pow.BigTen(uint64(n)))
}

// Int32 returns true if x can fit in an int32.
func Int32(x int64) (int32, bool) {
	return int32(x), x <= math.MaxInt32 && x >= math.MinInt32
}
