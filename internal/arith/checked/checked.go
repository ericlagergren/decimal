package checked

import (
	"math/big"

	"github.com/EricLagergren/decimal/internal/arith"
	"github.com/EricLagergren/decimal/internal/arith/pow"
	"github.com/EricLagergren/decimal/internal/c"
)

// Add returns x + y and a bool indicating whether the
// addition was successful.
func Add(x, y int64) (sum int64, ok bool) {
	sum = x + y
	// Algorith from "Hacker's Delight" 2-12
	return sum, (sum^x)&(sum^y) >= 0
}

// Mul returns x * y and a bool indicating whether the addition
// was successful.
func Mul(x, y int64) (prod int64, ok bool) {
	prod = x * y
	return prod, ((arith.Abs(x)|arith.Abs(y))>>31 == 0 || prod/y == x)
}

// Sub returns x - y and a bool indicating whether the addition
// was successful.
func Sub(x, y int64) (diff int64, ok bool) {
	return Add(x, -y)
}

// SumSub returns x + y - z and a bool indicating whether the operations
// were successful.
func SumSub(x, y, z int64) (res int64, ok bool) {
	res, ok = Add(x, y)
	if !ok {
		return res, false
	}
	return Sub(res, z)
}

// MulPow10 computes 10 * x ** n and a bool indicating whether
// the multiplcation was successful.
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
