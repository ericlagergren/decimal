// Package c provides internal constants.
package c

import (
	"math"
	"math/big"
)

const (
	is64bit = 1 << (^uintptr(0) >> 63) / 2 // 0 or 1
	is32bit = is64bit &^ 1                 // 0 or 1

	maxScale32    = 425000000
	maxScaleInf32 = 1000000001

	maxScale64    = 999999999999999999
	maxScaleInf64 = 2000000000000000001

	MaxScale    = maxScale32*is32bit + maxScale64*is64bit
	MaxScaleInf = maxScaleInf32*is32bit + maxScaleInf64*is64bit
)

const Inflated uint64 = math.MaxUint64

var (
	OneInt        = big.NewInt(1)
	TwoInt        = big.NewInt(2)
	FiveInt       = big.NewInt(5)
	TenInt        = big.NewInt(10)
	OneMillionInt = big.NewInt(1000000)

	TenFloat = big.NewFloat(10)

	MaxInt64 = big.NewInt(math.MaxInt64)
	MinInt64 = big.NewInt(math.MinInt64)

	MaxInt32 = big.NewInt(math.MaxInt32)
	MinInt32 = big.NewInt(math.MinInt32)
)
