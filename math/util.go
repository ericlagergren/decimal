package math

import (
	"github.com/ericlagergren/decimal"
)

var (
	negfour   = decimal.New(-4, 0)
	zero      = decimal.New(0, 0)
	one       = decimal.New(1, 0)
	two       = decimal.New(2, 0)
	three     = decimal.New(3, 0)
	six       = decimal.New(6, 0)
	eight     = decimal.New(8, 0)
	eleven    = decimal.New(11, 0)
	sixteen   = decimal.New(16, 0)
	eighteen  = decimal.New(18, 0)
	thirtyTwo = decimal.New(32, 0)
	eightyOne = decimal.New(81, 0)
)

// alias returns a if a != b, otherwise it returns a newly-allocated Big. It
// should be used if a *might* be able to be used for storage, but only if it
// doesn't alias b. The returned Big will have a's Context.
func alias(a, b *decimal.Big) *decimal.Big {
	if a != b {
		return a
	}
	return decimal.WithContext(a.Context)
}

func precision(z *decimal.Big) (p int) {
	p = z.Context.Precision
	if p > 0 && p <= decimal.UnlimitedPrecision {
		return p
	}
	if p == 0 {
		z.Context.Precision = decimal.DefaultPrecision
	} else {
		z.Context.Conditions |= decimal.InvalidContext
	}
	return decimal.DefaultPrecision
}

func maxscl(x *decimal.Big) int {
	if x.Context.MaxScale != 0 {
		return x.Context.MaxScale
	}
	return decimal.MaxScale
}

func minscl(x *decimal.Big) int {
	if x.Context.MinScale != 0 {
		return x.Context.MinScale
	}
	return decimal.MinScale
}

func etiny(z *decimal.Big) int    { return minscl(z) - (precision(z) - 1) }
func adjusted(x *decimal.Big) int { return (-x.Scale() + x.Precision()) - 1 }

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
