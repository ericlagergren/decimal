// Package c provides basic, internal constants.
package c

import (
	"math"
	"math/big"
)

const (
	BadScale = math.MaxInt32
	Inflated = math.MaxInt64
)

var (
	TenInt = big.NewInt(10)

	MaxInt64 = big.NewInt(math.MaxInt64)
	MinInt64 = big.NewInt(math.MinInt64)

	MaxInt32 = big.NewInt(math.MaxInt32)
	MinInt32 = big.NewInt(math.MinInt32)
)
