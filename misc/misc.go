// Package misc contains miscellaneous decimal routes.
package misc

import (
	"github.com/ericlagergren/decimal"
)

// CmpTotal compares x and y in a manner similar to the Cmp method but allows
// ordering of all abstract representations. In particular, this means NaN
// values have a defined ordering. From lowest to highest the ordering is:
//
//  -Infinity
//  -127
//  -1.00
//  -1
//  -0.000
//  -0
//  0
//  1.2300
//  1.23
//  1E+9
//  Infinity
//  sNaN
//  NaN
//
func CmpTotal(x, y *decimal.Big) int {
	xs := nanOrd(x)
	ys := nanOrd(y)
	switch {
	case xs > ys:
		return +1
	case xs < ys:
		return -1
	case xs == ys:
		return 0
	default:
		return x.Cmp(y)
	}
}

func nanOrd(x *decimal.Big) int {
	if x.IsNaN(0) {
		if x.IsNaN(+1) { // qnan
			return +1
		}
		return 0 // snan
	}
	return -1 // non-nan
}
