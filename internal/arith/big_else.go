// +build !amd64

package arith

import (
	"math/big"
)

func Words(x int64) []big.Word {
	ux := Abs(x)
	if w := Word(ux); uint64(w) == ux {
		return []big.Word{ux}
	}
	return []big.Word{ux, ux >> 32}
}

func Add(z, x *big.Int, y int64) *big.Int {
	return z.Add(x, big.NewInt(y))
}

func Sub(z, x *big.Int, y int64) *big.Int {
	return z.Sub(x, big.NewInt(y))
}

func Add128(z *big.Int, x, y int64) *big.Int {
	return z.Add(big.NewInt(x), big.NewInt(y))
}

func Sub128(z *big.Int, x, y int64) *big.Int {
	return z.Sub(big.NewInt(x), big.NewInt(y))
}

func Mul128(z *big.Int, x, y int64) *big.Int {
	return z.Mul(big.NewInt(x), big.NewInt(y))
}

func MulInt64(z, x *big.Int, y int64) *big.Int {
	return z.Mul(x, big.NewInt(y))
}
