// +build amd64

package arith

import (
	"math/big"
)

func Word(x int64) big.Word    { return big.Word(Abs(x)) }
func Words(x int64) []big.Word { return []big.Word{Word(x)} }

type uint128 [2]big.Word

// Add128 sets z to x + y and returns z.
func Add128(z *big.Int, x, y int64) *big.Int {
	ww := uint128{Word(x), Word(y)}
	neg := x < 0
	if neg == (y < 0) {
		ww[1], ww[0] = addWW(ww[0], ww[1])
	} else {
		if ww[0] >= ww[1] {
			ww[1], ww[0] = subWW(ww[0], ww[1])
		} else {
			neg = !neg
			ww[1], ww[0] = subWW(ww[1], ww[0])
		}
	}
	z.SetBits(ww[:])
	if neg {
		z.Neg(z)
	}
	return z
}

// Sub128 sets z to x - y and returns z.
func Sub128(z *big.Int, x, y int64) *big.Int {
	ww := uint128{Word(x), Word(y)}
	neg := x < 0
	if (x < 0) != (y < 0) {
		ww[1], ww[0] = addWW(ww[0], ww[1])
	} else {
		if ww[0] >= ww[1] {
			ww[1], ww[0] = subWW(ww[0], ww[1])
		} else {
			neg = !neg
			ww[1], ww[0] = subWW(ww[1], ww[0])
		}
	}
	z.SetBits(ww[:])
	if neg {
		z.Neg(z)
	}
	return z
}

// Mul128 sets z to x * y and returns z.
func Mul128(z *big.Int, x, y int64) *big.Int {
	var ww uint128
	ww[1], ww[0] = mulWW(Word(x), Word(y))
	z.SetBits(ww[:])
	if (x < 0) != (y < 0) {
		z.Neg(z)
	}
	return z
}

// MulInt64 sets z to x * y and returns z.
func MulInt64(z, x *big.Int, y int64) *big.Int {
	if y == 0 || x.Sign() == 0 {
		return z.SetUint64(0)
	}
	z.SetBits(mulAddWW(z.Bits(), x.Bits(), Word(y)))
	if (x.Sign() < 0) != (y < 0) { // no len check since x != 0 && y != 0
		z.Neg(z)
	}
	return z
}

// The following is (mostly) copied from math/big/arith.go, licensed under the
// BSD 3-clause license: https://github.com/golang/go/blob/master/LICENSE

const (
	_W  = 64       // word size in bits
	_S  = _W / 8   // word size in bytes
	_B  = 1 << _W  // digit base
	_M  = _B - 1   // digit mask
	_W2 = _W / 2   // half word size in bits
	_B2 = 1 << _W2 // half digit base
	_M2 = _B2 - 1  // half digit mask
)

func norm(z []big.Word) []big.Word {
	i := len(z)
	for i > 0 && z[i-1] == 0 {
		i--
	}
	return z[0:i]
}

func makeWord(z []big.Word, n int) []big.Word {
	if n <= cap(z) {
		return z[:n]
	}
	const e = 4
	return make([]big.Word, n, n+e)
}

func mulAddWW(z, x []big.Word, y big.Word) []big.Word {
	m := len(x)
	z = makeWord(z, m+1)
	z[m] = mulAddVWW(z[0:m], x, y)
	return norm(z)
}

// TODO(eric): add r if needed
func mulAddVWW(z, x []big.Word, y big.Word) (c big.Word) {
	for i := range z {
		c, z[i] = mulAddWWW(x[i], y, c)
	}
	return c
}

func mulAddWWW(x, y, c big.Word) (z1, z0 big.Word) {
	z1, zz0 := mulWW(x, y)
	if z0 = zz0 + c; z0 < zz0 {
		z1++
	}
	return z1, z0
}

func mulWW(x, y big.Word) (z1, z0 big.Word) {
	x0 := x & _M2
	x1 := x >> _W2
	y0 := y & _M2
	y1 := y >> _W2
	w0 := x0 * y0
	t := x1*y0 + w0>>_W2
	w1 := t & _M2
	w2 := t >> _W2
	w1 += x0 * y1
	z1 = x1*y1 + w2 + w1>>_W2
	z0 = x * y
	return
}

func addWW(x, y big.Word) (z1, z0 big.Word) {
	z0 = x + y
	z1 = (x&y | (x|y)&^z0) >> (_W - 1)
	return
}

func subWW(x, y big.Word) (z1, z0 big.Word) {
	z0 = x - y
	z1 = (y&^x | (y|^x)&z0) >> (_W - 1)
	return
}
