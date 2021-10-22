// Package misc contains miscellaneous decimal routes.
//
// Deprecated: the functionality has been moved into the decimal
//             package itself.
package misc

import "github.com/ericlagergren/decimal"

const (
	// Radix is the base in which decimal arithmetic is effected.
	Radix = decimal.Radix

	// IsCanonical is true since Big decimals are always normalized.
	IsCanonical = decimal.IsCanonical
)

// Canonical sets z to the canonical form of z.
//
// Since Big values are always canonical, it's identical to Copy.
func Canonical(z, x *decimal.Big) *decimal.Big {
	return z.Canonical(x)
}

// CmpTotal compares x and y in a manner similar to the Big.Cmp,
// but allows ordering of all abstract representations.
//
// In particular, this means NaN values have a defined ordering.
// From lowest to highest the ordering is:
//
//    -NaN
//    -sNaN
//    -Infinity
//    -127
//    -1.00
//    -1
//    -0.000
//    -0
//    +0
//    +1.2300
//    +1.23
//    +1E+9
//    +Infinity
//    +sNaN
//    +NaN
//
func CmpTotal(x, y *decimal.Big) int {
	return x.CmpTotal(y)
}

// CmpTotalAbs is like CmpTotal but instead compares the absolute
// values of x and y.
func CmpTotalAbs(x, y *decimal.Big) int {
	return x.CmpTotalAbs(y)
}

// CopyAbs is like Abs, but no flags are changed and the result
// is not rounded.
func CopyAbs(z, x *decimal.Big) *decimal.Big {
	return z.CopyAbs(x)
}

func CopyNeg(z, x *decimal.Big) *decimal.Big {
	return z.CopyNeg(x)
}

// Mantissa returns the mantissa of x and reports whether the
// mantissa fits into a uint64 and x is finite.
//
// This may be used to convert a decimal representing a monetary
// value to its most basic unit (e.g., $123.45 to 12345 cents).
func Mantissa(x *decimal.Big) (uint64, bool) {
	return x.Mantissa()
}

// Max returns the greater of the provided values.
//
// The result is undefined if no values are are provided.
func Max(x ...*decimal.Big) *decimal.Big {
	m := x[0]
	for _, v := range x[1:] {
		if v.Cmp(m) > 0 {
			m = v
		}
	}
	return m
}

// MaxAbs returns the greater of the absolute value of the provided values.
//
// The result is undefined if no values are provided.
func MaxAbs(x ...*decimal.Big) *decimal.Big {
	m := x[0]
	for _, v := range x[1:] {
		if v.CmpAbs(m) > 0 {
			m = v
		}
	}
	return m
}

// Min returns the lesser of the provided values.
//
// The result is undefined if no values are are provided.
func Min(x ...*decimal.Big) *decimal.Big {
	m := x[0]
	for _, v := range x[1:] {
		if v.Cmp(m) < 0 {
			m = v
		}
	}
	return m
}

// MinAbs returns the lesser of the absolute value of the
// provided values.
//
// The result is undefined if no values are provided.
func MinAbs(x ...*decimal.Big) *decimal.Big {
	m := x[0]
	for _, v := range x[1:] {
		if v.CmpAbs(m) < 0 {
			m = v
		}
	}
	return m
}

// NextMinus sets z to the smallest representable number that's
// smaller than x and returns z.
//
// If x is negative infinity the result will be negative
// infinity. If the result is zero its sign will be negative and
// its scale will be MinScale.
func NextMinus(z, x *decimal.Big) *decimal.Big {
	return z.Context.NextMinus(z, x)
}

// NextPlus sets z to the largest representable number that's
// larger than x and returns z.
//
// If x is positive infinity the result will be positive
// infinity. If the result is zero it will be positive and its
// scale will be MaxScale.
func NextPlus(z, x *decimal.Big) *decimal.Big {
	return z.Context.NextPlus(z, x)
}

// SameQuantum reports whether x and y have the same exponent
// (scale).
func SameQuantum(x, y *decimal.Big) bool {
	return x.SameQuantum(y)
}

// SetSignbit sets z to -z if sign is true, otherwise to +z.
func SetSignbit(z *decimal.Big, sign bool) *decimal.Big {
	return z.SetSignbit(sign)
}
