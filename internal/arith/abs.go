package arith

import "math/big"

// Abs32 returns |x|.
func Abs32(x int32) int32 {
	mask := -int32(uint32(x) >> 31)
	return (x + mask) ^ mask
}

// Abs returns |x|.
func Abs(x int64) int64 {
	mask := -int64(uint64(x) >> 63)
	return (x + mask) ^ mask
}

// AbsCmp compares |x| and |y|
func AbsCmp(x, y int64) int {
	x = Abs(x)
	y = Abs(y)
	if x > y {
		return +1
	}
	if x == y {
		return 0
	}
	return -1
}

// BigAbsCmp compares |x| and |y|.
func BigAbsCmp(x, y *big.Int) (r int) {
	return cmp(x.Bits(), y.Bits())
}

// TODO(eric): create a request for a CmpAbs method.
func cmp(x, y []big.Word) (r int) {
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
