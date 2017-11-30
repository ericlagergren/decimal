package math

import "github.com/ericlagergren/decimal"

var (
	negfour   = decimal.New(-4, 0)
	one       = decimal.New(1, 0)
	two       = decimal.New(2, 0)
	six       = decimal.New(6, 0)
	eight     = decimal.New(8, 0)
	sixteen   = decimal.New(16, 0)
	thirtyTwo = decimal.New(32, 0)
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
	if p > 0 {
		return p
	}
	if p == 0 {
		z.Context.Precision = decimal.DefaultPrecision
	} else {
		z.Context.Conditions |= decimal.InvalidContext
	}
	return decimal.DefaultPrecision
}

func etiny(z *decimal.Big) int { return decimal.MinScale - (precision(z) - 1) }
