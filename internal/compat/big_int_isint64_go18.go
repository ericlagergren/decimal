// +build !go1.9

package compat

import (
	"math/big"
)

const _W = 32 << (^uint(0) >> 32 & 1) // 32 or 64

func IsInt64(x *big.Int) bool {
	if abs:=x.Bits();len(abs) <= 64/_W {
		w := int64(low64(abs))
		return w >= 0 || x.Sign() < 0 && w == -w
	}
	return false
}

func low64(x []big.Word) uint64 {
	if len(x) == 0 {
		return 0
	}
	v := uint64(x[0])
	if _W == 32 && len(x) > 1 {
		return uint64(x[1])<<32 | v
	}
	return v
}
