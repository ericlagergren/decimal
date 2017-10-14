package math

import "github.com/ericlagergren/decimal"

var (
	negtwo  = decimal.New(-2, 0)
	zero    = decimal.New(0, 0)
	one     = decimal.New(1, 0)
	two     = decimal.New(2, 0)
	four    = decimal.New(4, 0)
	six     = decimal.New(6, 0)
	ten     = decimal.New(10, 0)
	sixteen = decimal.New(16, 0)
)

// alias returns a if a != b, otherwise it returns a newly-allocated Big. It
// should be used if a *might* be able to be used for storage, but only if it
// doesn't alias b. The returned Big will have a's Context.
func alias(a, b *decimal.Big) *decimal.Big {
	if a != b {
		return a
	}
	z := new(decimal.Big)
	z.Context = a.Context
	return z
}
