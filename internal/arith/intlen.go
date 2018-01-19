package arith

import (
	"math/big"
	"math/bits"

	"github.com/ericlagergren/decimal/internal/compat"
)

func IsUint64(x *big.Int) bool {
	return len(x.Bits()) <= 64/bits.UintSize
}

// Length returns the number of digits in x.
func Length(x uint64) (r int) {
	if x >= 10 {
		// From https://graphics.stanford.edu/~seander/bithacks.html#IntegerLog10
		r = int((bits.Len64(x) * 1233) >> 12)
		if v, ok := Pow10(uint64(r)); !ok || x < v {
			return r
		}
	}
	return r + 1
}

// BigLength returns the number of digits in x.
func BigLength(x *big.Int) int {
	if x.Sign() == 0 {
		return 1
	}

	var (
		m  uint64
		nb = uint64(x.BitLen())
	)

	// overflowCutoff is the largest number where N * 0x268826A1 <= 1<<63 - 1
	const overflowCutoff = 14267572532
	if nb > overflowCutoff {
		// Given the identity ``log_n a + log_n b = log_n a*b''
		// and ``(1<<63 - 1) / overflowCutoff < overFlowCutoff''
		// we can break nb into two factors: overflowCutoff and X.

		// overflowCutoff / log10(2)
		m = 4294967295
		nb = (nb / overflowCutoff) + (nb % overflowCutoff)
	}

	// 0x268826A1/2^31 is an approximation of log10(2). See ilog10.
	// The more accurate approximation 0x268826A13EF3FE08/2^63 overflows.
	m += ((nb + 1) * 0x268826A1) >> 31

	if compat.BigCmpAbs(x, BigPow10(m)) < 0 {
		return int(m)
	}
	return int(m + 1)
}
