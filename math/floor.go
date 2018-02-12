package math

import (
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/misc"
)

// Floor sets z to the greatest integer value less than or equal to x and returns
// z.
func Floor(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}
	ctx := z.Context
	if z.Signbit() {
		ctx.RoundingMode = decimal.AwayFromZero
	} else {
		ctx.RoundingMode = decimal.ToZero
	}
	return ctx.RoundToInt(z.Copy(x))
}

// Ceil sets z to the least integer value greater than or equal to x and returns
// z.
func Ceil(z, x *decimal.Big) *decimal.Big {
	// ceil(x) = -floor(-x)
	return z.Neg(Floor(z, misc.CopyNeg(z, x)))
}
