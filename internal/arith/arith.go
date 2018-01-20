package arith

import (
	"math/big"
	"math/bits"
)

const Is32Bit = bits.UintSize == 32

func Add128(x, y uint64) (z1, z0 uint64) {
	z0 = x + y
	z1 = (x&y | (x|y)&^z0) >> 63
	return z1, z0
}

func Set128(z *big.Int, z1, z0 uint64) *big.Int {
	ww := makeWord(z.Bits(), 128/bits.UintSize)
	if Is32Bit {
		ww[3] = big.Word(z1 >> 32)
		ww[2] = big.Word(z1)
		ww[1] = big.Word(z0 >> 32)
		ww[0] = big.Word(z0)
	} else {
		ww[1] = big.Word(z1)
		ww[0] = big.Word(z0)
	}
	return z.SetBits(ww)
}

func makeWord(z []big.Word, n int) []big.Word {
	if n <= cap(z) {
		return z[:n]
	}
	const e = 4
	return make([]big.Word, n, n+e)
}

func Words(x uint64) []big.Word {
	if Is32Bit {
		return []big.Word{big.Word(x), big.Word(x >> 32)}
	}
	return []big.Word{big.Word(x)}
}
