package arith

// This section is borrowed and slightly modified from
// https://golang.org/src/math/...

const (
	_m    = 18446744073709551615 // ^uint64(0)
	_logS = _m>>8&1 + _m>>16&1 + _m>>32&1
	_S    = 1 << _logS

	_W = _S << 3 // word size in bits
	_B = 1 << _W // digit base
	_M = _B - 1  // digit mask

	WordSize = _W
)

const deBruijn64 = 0x03f79d71b4ca8b09

var deBruijn64Lookup = [64]byte{
	0, 1, 56, 2, 57, 49, 28, 3, 61, 58, 42, 50, 38, 29, 17, 4,
	62, 47, 59, 36, 45, 43, 51, 22, 53, 39, 33, 30, 24, 18, 12, 5,
	63, 55, 48, 27, 60, 41, 37, 16, 46, 35, 44, 21, 52, 32, 23, 11,
	54, 26, 40, 15, 34, 20, 31, 10, 25, 14, 19, 9, 13, 8, 7, 6,
}

func CTZ(x int64) (n int) {
	// Faster to do this in Go than assembler.
	return int(deBruijn64Lookup[uint64((x&-x)*(deBruijn64&_M))>>58])
}
