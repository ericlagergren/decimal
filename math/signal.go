package math

import (
	"fmt"

	"github.com/ericlagergren/decimal"
)

// TODO(eric): export ../decimal.go's version of this?
func signal(x *decimal.Big, c decimal.Condition, err error) *decimal.Big {
	switch ctx := &x.Context; ctx.OperatingMode {
	case decimal.Go:
		// Go mode always panics on NaNs.
		if _, ok := err.(decimal.ErrNaN); ok {
			panic(err)
		}
	case decimal.GDA:
		ctx.Conditions = c
		if c&ctx.Traps != 0 {
			ctx.Err = err
		}
	default:
		ctx.Conditions = c | decimal.InvalidContext
		ctx.Err = fmt.Errorf("invalid OperatingMode: %d", ctx.OperatingMode)
		// TODO(eric): Add a SetNaN method?
		x.SetString("qNaN")
	}
	return x
}
