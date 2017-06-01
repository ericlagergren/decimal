package arith

import "math/big"

// Abs returns |x|.
func Abs(x int64) int64 {
	mask := -int64(uint64(x) >> 63)
	return (x + mask) ^ mask
}

// BigAbs returns |x|.
func BigAbs(x *big.Int) *big.Int {
	m := make([]big.Word, len(x.Bits()))
	copy(m, x.Bits())
	return new(big.Int).SetBits(m)
}

// bigAbsAlias returns a big.Int set to |x| whose inner slices
// alias each other. Do not use unless the return value will not be
// modified.
func bigAbsAlias(x *big.Int) *big.Int {
	return new(big.Int).SetBits(x.Bits())
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

func BigAbsCmp(x, y big.Int) int {
	var x0, y0 big.Int
	// SetBits sets to |v|, thus giving an absolute comparison.
	x0.SetBits(x.Bits())
	y0.SetBits(y.Bits())
	return x0.Cmp(&y0)
}
