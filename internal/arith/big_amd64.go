// +build amd64

package arith

import (
	"math/big"
)

func Word(x uint64) big.Word    { return big.Word(x) }
func Words(x uint64) []big.Word { return []big.Word{Word(x)} }

type uint128 [2]big.Word

// Add sets z to x + y and returns z.
func Add(z, x *big.Int, y uint64) *big.Int {
	zw := z.Bits()
	xw := x.Bits()
	yw := Word(y)

	neg := x.Sign() < 0
	switch {
	case len(xw) == 0:
		neg = false
		zw = setw(zw, yw)
	case y == 0:
		zw = set(zw, xw)
	case !neg:
		zw = add(zw, xw, yw)
	case len(xw) > 1, xw[0] >= yw:
		zw = sub(zw, xw, yw)
	default: // len(xw) == 1 && y < xw[0]
		neg = !neg
		zw = sub(zw, Words(y), xw[0])
	}

	z.SetBits(zw)
	if neg {
		z.Neg(z)
	}
	return z
}

// Sub sets z to x - y and returns z.
func Sub(z, x *big.Int, y uint64) *big.Int {
	zw := z.Bits()
	xw := x.Bits()
	yw := Word(y)

	neg := x.Sign() < 0
	switch {
	case len(xw) == 0:
		neg = true
		zw = setw(zw, yw)
	case y == 0:
		zw = set(zw, xw)
	case neg:
		zw = add(zw, xw, yw)
	case len(xw) > 1, xw[0] >= yw:
		zw = sub(zw, xw, yw)
	default: // len(xw) == 1 && y < xw[0]
		neg = !neg
		zw = sub(zw, Words(y), xw[0])
	}

	z.SetBits(zw)
	if neg {
		z.Neg(z)
	}
	return z
}

// Add128 sets z to x + y and returns z.
func Add128(z *big.Int, x, y uint64) *big.Int {
	var ww uint128
	ww[1], ww[0] = addWW(Word(x), Word(y))
	return z.SetBits(ww[:])
}

// Sub128 sets z to x - y and returns z.
func Sub128(z *big.Int, x, y uint64) *big.Int {
	ww := uint128{Word(x), Word(y)}
	var neg bool
	if ww[0] >= ww[1] {
		ww[1], ww[0] = subWW(ww[0], ww[1])
	} else {
		neg = true
		ww[1], ww[0] = subWW(ww[1], ww[0])
	}
	z.SetBits(ww[:])
	if neg {
		z.Neg(z)
	}
	return z
}

// Mul128 sets z to x * y and returns z.
func Mul128(z *big.Int, x, y uint64) *big.Int {
	if x == 0 || y == 0 {
		return z.SetUint64(0)
	}
	var ww uint128
	ww[1], ww[0] = mulWW(Word(x), Word(y))
	return z.SetBits(ww[:])
}

// MulInt64 sets z to x * y and returns z.
func MulInt64(z, x *big.Int, y uint64) *big.Int {
	if y == 0 || x.Sign() == 0 {
		return z.SetUint64(0)
	}
	z.SetBits(mulAddWW(z.Bits(), x.Bits(), Word(y)))
	if x.Sign() < 0 { // no len check since x != 0 && y != 0
		z.Neg(z)
	}
	return z
}

// The following is (mostly) copied from math/big/arith.go, licensed under the
// BSD 3-clause license: https://github.com/golang/go/blob/master/LICENSE

const (
	_W  = 64       // word size in bits
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

// NOTE(eric): add r if needed
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

// mulWW returns both halves of the 128-bit multiplication, x * y.
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

func set(z, x []big.Word) []big.Word {
	z = makeWord(z, len(x))
	copy(z, x)
	return z
}

func setw(z []big.Word, x big.Word) []big.Word {
	z = makeWord(z, 1)
	z[0] = x
	return z
}

// add sets z to x + y and returns z.
func add(z, x []big.Word, y big.Word) []big.Word {
	m := len(x)
	const n = 1

	// m > 0 && y > 0

	z = makeWord(z, m+1)
	var c big.Word
	// addVV(z[0:m], x, y) but WW since len(y) == 1
	c, z[0] = addWW(x[0], y)
	if m > n {
		c = addVW(z[n:m], x[n:], c)
	}
	z[m] = c
	return norm(z)
}

// sub sets z to x - y and returns z.
func sub(z, x []big.Word, y big.Word) []big.Word {
	m := len(x)
	const n = 1

	switch {
	case m < n:
		panic("underflow")
	case m == 0:
		return setw(z, y)
	case y == 0:
		return set(z, x)
	}
	// m > 0

	z = makeWord(z, m)
	// subVV(z[0:m], x, y) but WW since len(y) == 1
	var c big.Word
	c, z[0] = subWW(x[0], y)
	if m > n {
		c = subVW(z[n:], x[n:], c)
	}
	if c != 0 {
		panic("underflow")
	}
	return norm(z)
}

// addVW sets z to x + y and returns the carry.
func addVW(z, x []big.Word, y big.Word) (c big.Word) {
	c = y
	for i, xi := range x[:len(z)] {
		zi := xi + c
		z[i] = zi
		c = xi &^ zi >> (_W - 1)
	}
	return c
}

// subVW sets z to x - y and returns the carry.
func subVW(z, x []big.Word, y big.Word) (c big.Word) {
	c = y
	for i, xi := range x[:len(z)] {
		zi := xi - c
		z[i] = zi
		c = zi &^ xi >> (_W - 1)
	}
	return c
}

// addWW returns both halves of the 128-bit addition, x + y.
func addWW(x, y big.Word) (z1, z0 big.Word) {
	z0 = x + y
	z1 = (x&y | (x|y)&^z0) >> (_W - 1)
	return
}

// subWW returns both halves of the 128-bit subtraction, x - y.
func subWW(x, y big.Word) (z1, z0 big.Word) {
	z0 = x - y
	z1 = (y&^x | (y|^x)&z0) >> (_W - 1)
	return
}
