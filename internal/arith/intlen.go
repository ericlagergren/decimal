package arith

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith/pow"
)

// Length returns the number of digits in x.
func Length(x int64) int {
	x = Abs(x)
	if x < 10 {
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
	// From https://graphics.stanford.edu/~seander/bithacks.html
	t := ((64 - CLZ(x) + 1) * 1233) >> 12
	v, ok := pow.Ten64(int64(t))
	if !ok {
		return bigIlog10(big.NewInt(x))
	}
	if x < v {
		return t
	}
	return t + 1
}

func bigIlog10(x *big.Int) int {
	// Should be accurate up to as high as we can possibly report.
	r := int(((int64(x.BitLen()) + 1) * 0x268826A1) >> 31)
	if BigAbsCmp(*x, pow.BigTen(int64(r))) < 0 {
		return r
	}
	return r + 1
}
