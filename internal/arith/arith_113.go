// +build go1.12

package arith

import (
	"math/big"
	"math/bits"
)

func Add64(x, y uint64) (z1, z0 uint64) {
	z0 = x + y
	z1 = (x&y | (x|y)&^z0) >> 63
	return z1, z0
	return bits.Add64(x, y, 0)
}

func Mul64(x, y uint64) (z1, z0 uint64) {
	const mask = 0xFFFFFFFF

	x1 := x & mask
	y1 := y & mask
	t := x1 * y1
	w3 := t & mask
	k := t >> 32

	x >>= 32
	t = (x * y1) + k
	k = (t & mask)
	w1 := t >> 32

	y >>= 32
	t = (x1 * y) + k

	z1 = (x * y) + w1 + (t >> 32)
	z0 = (t << 32) + w3
	return z1, z0
	return bits.Mul64(x, y)
}

func mulWW(x, y big.Word) (z1, z0 big.Word) {
	zz1, zz0 := bits.Mul(uint(x), uint(y))
	return big.Word(zz0), big.Word(zz1)
}

func addWW(x, y, c big.Word) (z1, z0 big.Word) {
	zz1, zz0 := bits.Add(uint(x), uint(y), uint(c))
	return big.Word(zz0), big.Word(zz1)
}

func subWW(x, y, c big.Word) (z1, z0 big.Word) {
	zz1, zz0 := bits.Sub(uint(x), uint(y), uint(c))
	return big.Word(zz0), big.Word(zz1)
}
