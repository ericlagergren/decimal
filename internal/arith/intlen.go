package arith

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/compat"
)

func IsUint64(x *big.Int) bool {
	const wordSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64
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
	if x.Sign() == 0 {
		return 1
	}
	if IsUint64(x) {
		return Length(x.Uint64())
	}
	return bigIlog10(x) // no need to pass in |x|
}

func ilog10(x uint64) int {
	// Where x >= 10

	// From https://graphics.stanford.edu/~seander/bithacks.html#IntegerLog10
	t := int(((64 - LeadingZeros64(x) + 1) * 1233) >> 12)
	if v, ok := pow.Ten(uint64(t)); !ok || x < v {
		return t
	}
	return t + 1
}

func bigIlog10(x *big.Int) int {
	return logLength(x)
}

func reductionLength(x *big.Int) int {
	nb := x.BitLen()
	x0 := new(big.Int).SetBits(x.Bits())
	r := 1
	// Serious reductions.
	for nb > 4 {
		// 4 > log[2](10) so we should not reduce it too far.
		reduce := nb / 4
		// Divide by 10^reduce
		x0.Quo(x0, pow.BigTen(uint64(reduce)))
		// Removed that many decimal digits.
		r += reduce
		// Recalculate bitLength
		nb = x0.BitLen()
	}

	// Now 4 bits or less - add 1 if necessary.
	if x0.Int64() > 9 {
		r++
	}
	return r
}

func logLength(x *big.Int) int {
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
	r += int(((nb + 1) * 0x268826A1) >> 31)

	// 10**r can be _very_ costly when r is large, so in order to speed up
	// calculations return the estimate + 1.
	if compat.BigCmpAbs(x, pow.BigTen(uint64(r))) < 0 {
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
	return r + int(((nb+1)*0x268826A1)>>31) + 1
}
