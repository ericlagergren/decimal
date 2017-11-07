package arith

import (
	"math/big"
)

// Abs returns |x|.
func Abs(x int64) int64 {
	mask := -int64(uint64(x) >> 63)
	return (x + mask) ^ mask
}

// AbsCmp compares |x| and |y|.
func AbsCmp(x, y int64) int {
	x = Abs(x)
	y = Abs(y)
	if x != y {
		if x > y {
			return +1
		}
		return -1
	}
	return 0
}

// AbsCmp128 compares |x| and |y|*shift in 128 bits.
func AbsCmp128(x, y int64, shift uint64) int {
	// x is unchanged so its high bits are always 0.
	const xh = 0
	yh, yl := mulWW(Word(y), big.Word(shift))
	if xh != yh {
		if xh > yh {
			return +1
		}
		return -1
	}
	xl := Word(x)
	if xl != yl {
		if xl > yl {
			return +1
		}
		return -1
	}
	return 0
}

func CmpBits(x, y []big.Word) (r int) {
	// Copied from math/big.nat.go
	m := len(x)
	n := len(y)
	if m != n || m == 0 {
		switch {
		case m < n:
			r = -1
		case m > n:
			r = 1
		}
		return
	}

	i := m - 1
	for i > 0 && x[i] == y[i] {
		i--
	}

	switch {
	case x[i] < y[i]:
		r = -1
	case x[i] > y[i]:
		r = 1
	}
	return
}
