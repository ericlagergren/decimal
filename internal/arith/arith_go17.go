// +build go1.7,!go1.9

package arith

func CLZ(x int64) (n int) {
	if x == 0 {
		return 64
	}
	return _W - BitLen(x)
}

func BitLen(x int64) (n int) {
	for ; x >= 0x8000; x >>= 16 {
		n += 16
	}
	if x >= 0x80 {
		x >>= 8
		n += 8
	}
	if x >= 0x8 {
		x >>= 4
		n += 4
	}
	if x >= 0x2 {
		x >>= 2
		n += 2
	}
	if x >= 0x1 {
		n++
	}
	return n
}
