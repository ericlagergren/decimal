package arith

import (
	"math/big"
	"math/bits"

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

func ilog10(x uint64) int {
	// Where x >= 10

	// From https://graphics.stanford.edu/~seander/bithacks.html#IntegerLog10
	//t := int(((64 - bits.LeadingZeros64(x) + 1) * 1233) >> 12)
	t := int((bits.Len64(x) * 1233) >> 12)
	if v, ok := Pow10(uint64(t)); !ok || x < v {
		return t
	}
	return t + 1
}

// BigLength returns the number of digits in x.
func BigLength(x *big.Int) int {
	if x.Sign() == 0 {
		return 1
	}
	return logLength(x, x.BitLen())
}

func logLength(x *big.Int, nb int) (r int) {
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

func logLengthNoCmp(x *big.Int, nb int) (r int) {
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

	// Per exploringbinary.com/a-pattern-in-powers-of-ten-and-their-binary-equivalents/
	// the trailing digits of a power of 10 in binary will match the power of 10.
	//
	//    [100] -> 1100[100]
	//    [1000]-> 111110[1000].
	//
	// Normally, we would compare x to 10**r and, if x is smaller, increment its
	// length by 1 because the number of decimal digits in a number can span
	// multiple bit lengths. However, this can be costly if r is large, so
	// testing to see if x>>len(r) is set will give us better accuracy than
	// simply adding 1, which might result in incorrect lengths.
	if x.Bit(Length(uint64(r))) == 1 {
		return r
	}
	return r + 1
}

func logLengthIter(x *big.Int) (r int) {
	for _, w := range x.Bits() {
		if w != 0 {
			r += Length(uint64(w >> 1))
		}
	}
	return r
}
