package arith

func Abs(x int64) uint64 {
	m := x >> 63
	return uint64((x ^ m) - m)
}

func Cmp(x, y uint64) int {
	if x != y {
		if x > y {
			return +1
		}
		return -1
	}
	return 0
}

// AbsCmp128 compares |x| and |y|*shift in 128 bits.
func AbsCmp128(x, y, shift uint64) int {
	y1, y0 := Mul128(y, shift)
	if y1 != 0 {
		return +1
	}
	return Cmp(x, y0)
}
