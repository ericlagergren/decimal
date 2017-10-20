package arith

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith/pow"
)

// Length returns the number of digits in x.
func Length(x int64) int {
	if x = Abs(x); x < 10 {
		return 1
	}
	return ilog10(x)
}

// BigLength returns the number of digits in x.
func BigLength(x *big.Int) int {
	if x.Sign() == 0 {
		return 1
	}
	return bigIlog10(bigAbsAlias(x))
}

func ilog10(x int64) int {
	// Where x >= 10

	// From https://graphics.stanford.edu/~seander/bithacks.html#IntegerLog10
	t := int64(((64 - CLZ(x) + 1) * 1233) >> 12)
	if v, ok := pow.Ten64(t); !ok || x < v {
		return int(t)
	}
	return int(t) + 1
}

func bigIlog10(x *big.Int) int {
	// Where x > 0

	// 0x268826A1/2^31 is an approximation of log10(2). See ilog10.
	// The more accurate approximation 0x268826A13EF3FE08/2^63 overflows.
	r := ((int64(x.BitLen()) + 1) * 0x268826A1) >> 31
	if BigAbsCmp(x, pow.BigTen(r)) < 0 {
		return int(r)
	}
	return int(r) + 1
}
