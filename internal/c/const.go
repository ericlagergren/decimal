// Package c provides internal constants.
package c

import (
	"math"
	"math/big"
)

const Inflated int64 = math.MinInt64 + 1

var (
	OneInt = big.NewInt(1)
	TwoInt = big.NewInt(2)
	TenInt = big.NewInt(10)

	TenFloat = big.NewFloat(10)

	MaxInt64 = big.NewInt(math.MaxInt64)
	MinInt64 = big.NewInt(math.MinInt64)

	MaxInt32 = big.NewInt(math.MaxInt32)
	MinInt32 = big.NewInt(math.MinInt32)
)
