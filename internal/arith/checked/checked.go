// Package checked implements basic checked arithmetic.
package checked

import (
	"math"
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/c"
)

func UAdd(x, y uint) (sum uint, ok bool) {
	sum = x + y
	return sum, sum >= x
}

func UAdd64(x, y uint64) (sum uint64, ok bool) {
	return x + y, arith.LeadingZeros64(x) < 32 && arith.LeadingZeros64(y) < 32
}

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

// Mul32 returns x * y and a bool indicating whether the multiplication was
// successful.
func Mul32(x, y int32) (prod int32, ok bool) {
	p, ok := Mul(int64(x), int64(y))
	return int32(p), ok && int64(int32(p)) == p
}

// Sub returns x - y and a bool indicating whether the subtraction was successful.
func Sub(x, y int64) (diff int64, ok bool) {
	return Add(x, -y)
}

func USub(x, y uint) (diff uint, ok bool) {
	return x - y, x >= y
}

// Sub32 returns x - y and a bool indicating whether the subtraction was
// successful.
func Sub32(x, y int32) (diff int32, ok bool) {
	return Add32(x, -y)
}

// SumSub returns x + y - z and a bool indicating whether the operations were
// successful.
func SumSub(x, y, z int32) (res int32, ok bool) {
	// Use int64s since only the result needs to fit in an int32, not all the
	// intermediate steps.
	return Int32(int64(x) + int64(y) - int64(z))
}

// SubSum returns x - y + z and a bool indicating whether the operations were
// successful.
func SubSum(x, y, z int32) (res int32, ok bool) {
	// Use int64s since only the result needs to fit in an int32, not all the
	// intermediate steps.
	return Int32(int64(x) - int64(y) + int64(z))
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
