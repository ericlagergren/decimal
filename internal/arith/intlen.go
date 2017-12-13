package arith

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/compat"
)

const wordSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

func IsUint64(x *big.Int) bool {
	return len(x.Bits()) <= 64/wordSize
}

// Length returns the number of digits in x.
func Length(x uint64) int {
	if x < 10 {
		return 1
	}
	return ilog10(x)
}

// BigLength returns the number of digits in x.
func BigLength(x *big.Int) int {
	if b := x.Bits(); len(b) > 3 || len(b) == 0 {
		switch b := x.Bits(); len(b) {
		default:
		case 3:
			return Length(uint64(b[0])) + Length(uint64(b[1])) + Length(uint64(b[2]))
		case 2:
			return Length(uint64(b[0])) + Length(uint64(b[1]))
		case 1:
			return Length(uint64(b[0]))
		case 0:
			return 1
		}
	}
	return logLength(x, x.BitLen()) // no need to pass in |x|
}

func ilog10(x uint64) int {
	// Where x >= 10

	// From https://graphics.stanford.edu/~seander/bithacks.html#IntegerLog10
	t := int(((64 - LeadingZeros64(x) + 1) * 1233) >> 12)
	if v, ok := Pow10(uint64(t)); !ok || x < v {
		return t
	}
	return t + 1
}

func logLength(x *big.Int, nb int) int {
	var r int
	// overflowCutoff is the largest number where N * 0x268826A1 <= 1<<63 - 1
	const overflowCutoff = 14267572532
	if nb > overflowCutoff {
		// Given the identity ``log_n a + log_n b = log_n a*b''
		// and ``(1<<63 - 1) / overflowCutoff < overFlowCutoff''
		// we can break nb into two factors: overflowCutoff and X.

		// overflowCutoff / log10(2)
		r = 4294967295
		nb = (nb / overflowCutoff) + (nb % overflowCutoff)
	}
	// 0x268826A1/2^31 is an approximation of log10(2). See ilog10.
	// The more accurate approximation 0x268826A13EF3FE08/2^63 overflows.
	r += int(((nb + 1) * 0x268826A1) >> 31)

	if compat.BigCmpAbs(x, BigPow10(uint64(r))) < 0 {
		return r
	}
	return r + 1
}

func logLengthNoCmp(x *big.Int) int {
	nb := x.BitLen()
	var r int
	// overflowCutoff is the largest number where N * 0x268826A1 <= 1<<63 - 1
	const overflowCutoff = 14267572532
	if nb > overflowCutoff {
		// Given the identity ``log_n a + log_n b = log_n a*b''
		// and ``(1<<63 - 1) / overflowCutoff < overFlowCutoff''
		// we can break nb into two factors: overflowCutoff and X.

		// overflowCutoff / log10(2)
		r = 4294967295
		nb = (nb / overflowCutoff) + (nb % overflowCutoff)
	}
	// 0x268826A1/2^31 is an approximation of log10(2). See ilog10.
	// The more accurate approximation 0x268826A13EF3FE08/2^63 overflows.
	//
	// 10**r can be _very_ costly when r is large, so in order to speed up
	// calculations return the estimate + 1.
	return r + int(((nb+1)*0x268826A1)>>31) + 1
}

func logLengthIter(x *big.Int) int {
	var r int
	for _, w := range x.Bits() {
		r += Length(uint64(w))
	}
	return r
}
