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

// Mul returns x * y and a bool indicating whether the addition was successful.
func Mul(x, y int64) (prod int64, ok bool) {
	prod = x * y
	return prod, ((arith.Abs(x)|arith.Abs(y))>>31 == 0 || prod/y == x)
}

// Mul returns x * y and a bool indicating whether the addition as successful.
func Mul32(x, y int32) (prod int32, ok bool) {
	p, ok := Mul(int64(x), int64(y))
	return int32(p), ok && int64(int32(p)) == p
}

// Sub returns x - y and a bool indicating whether the addition was successful.
func Sub(x, y int64) (diff int64, ok bool) {
	return Add(x, -y)
}

// Sub32 returns x - y and a bool indicating whether the addition was
// successful.
func Sub32(x, y int32) (diff int32, ok bool) {
	return Add32(x, -y)
}

// SumSub returns x + y - z and a bool indicating whether the operations were
// successful.
func SumSub(x, y, z int32) (res int32, ok bool) {
	// Use int64s since only the result needs to fit in an int32, not all the
	// intermediate steps.
	v, ok := Add(int64(x), int64(y))
	if !ok {
		return 0, false
	}
	return Int32(v - int64(z))
}

// SubSum returns x - y + z and a bool indicating whether the operations were
// successful.
func SubSum(x, y, z int32) (res int32, ok bool) {
	// Use int64s since only the result needs to fit in an int32, not all the
	// intermediate steps.
	v, ok := Sub(int64(x), int64(y))
	if !ok {
		return 0, false
	}
	return Int32(v + int64(z))
}

// MulPow10 computes 10 * x ** n and a bool indicating whether the
// multiplcation was successful.
func MulPow10(x int64, n int32) (p int64, ok bool) {
	if x == 0 || n <= 0 || x == c.Inflated {
		return x, true
	}
	if n < pow.Tab64Len && n < pow.ThreshLen {
		if x == 1 {
			return pow.Ten64(int64(n))
		}
		if arith.Abs(int64(n)) < pow.Thresh(n) {
			p, ok := pow.Ten64(int64(n))
			if !ok {
				return 0, false
			}
			return Mul(x, p)
		}
	}
	return 0, false
}

// MulBigPow10 computes 10 * x ** n.
// It reuses x.
func MulBigPow10(x *big.Int, n int32) *big.Int {
	if x.Sign() == 0 || n <= 0 {
		return x
	}
	b := pow.BigTen(int64(n))
	return x.Mul(x, &b)
}

// Int32 returns true if x can fit in an int32.
func Int32(x int64) (int32, bool) {
	return int32(x), x <= math.MaxInt32 && x >= math.MinInt32
}
