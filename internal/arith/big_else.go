// +build !amd64

package arith

import "math/big"

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
