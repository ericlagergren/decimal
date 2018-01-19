package arith

import (
	"math/big"
	"math/bits"
)

const Is32Bit = bits.UintSize == 32

func Words(x uint64) []big.Word {
	if Is32Bit {
		return []big.Word{big.Word(x), big.Word(x >> 32)}
	}
	return []big.Word{big.Word(x)}
}
