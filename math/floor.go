package math

import (
	"github.com/ericlagergren/decimal/v4"
	"github.com/ericlagergren/decimal/v4/misc"
)

// Floor sets z to the greatest integer value less than or equal to x and returns
// z.
func Floor(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}
	ctx := z.Context
	ctx.RoundingMode = decimal.ToNegativeInf
	return ctx.RoundToInt(z.Copy(x))
}

// Ceil sets z to the least integer value greater than or equal to x and returns
// z.
func Ceil(z, x *decimal.Big) *decimal.Big {
	// ceil(x) = -floor(-x)
	return z.Neg(Floor(z, misc.CopyNeg(z, x)))
}
