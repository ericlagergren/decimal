// Package misc contains miscellaneous decimal routes.
package misc

import (
	"math/big"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/c"
)

var (
	pos = decimal.New(+1, 0)
	neg = decimal.New(-1, 0)
)

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

func etiny(z *decimal.Big) int { return minscl(z) - (precision(z) - 1) }
func etop(z *decimal.Big) int  { return maxscl(z) - (precision(z) - 1) }

const (
	// Radix is the base in which decimal arithmetic is effected.
	Radix = 10

	// IsCanonical is true since Big decimals are always normalized.
	IsCanonical = true
)

// Canonical sets z to the canonical form of z.
//
// Since Big values are always canonical, it's identical to Copy.
func Canonical(z, x *decimal.Big) *decimal.Big { return z.Copy(x) }

// TODO(eric): do these...
//
// And sets z to the digit-wise logical ``and'' of x and y and returns z.
// func And(z, x, y *Big) *Big
//
// Invert sets z to the digit-wise logical ``inversion'' of x and returns z.
// func Invert(z, x *Big) *Big
//
// Or sets z to the digit-wise logical ``or'' of x and y and returns z.
// func Or(z, x, y *Big) *Big
//
// Xor sets z to the digit-wise logical ``exclusive or'' of x and y and returns
// z.
// func Xor(z, x, y *Big) *Big

// CmpTotal compares x and y in a manner similar to the Big.Cmp, but allows
// ordering of all abstract representations.
//
// In particular, this means NaN values have a defined ordering. From lowest to
// highest the ordering is:
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
	xs := ord(x, false)
	ys := ord(y, false)
	if xs != ys {
		if xs > ys {
			return +1
		}
		return -1
	}
	if xs != 0 {
		return 0
	}
	return x.Cmp(y)
}

// CmpTotalAbs is like CmpTotal but instead compares the absolute values of x
// and y.
func CmpTotalAbs(x, y *decimal.Big) int {
	xs := ord(x, true)
	ys := ord(y, true)
	if xs != ys {
		if xs > ys {
			return +1
		}
		return -1
	}
	if xs != 0 {
		return 0
	}
	return x.CmpAbs(y)
}

// CopyAbs is like Abs, but no flags are changed and the result is not rounded.
func CopyAbs(z, x *decimal.Big) *decimal.Big {
	return z.CopySign(x, pos)
}

// CopyNeg is like Neg, but no flags are changed and the result is not rounded.
func CopyNeg(z, x *decimal.Big) *decimal.Big {
	if x.Signbit() {
		return z.CopySign(x, pos)
	}
	return z.CopySign(x, neg)
}

// Mantissa returns the mantissa of x. If the mantissa cannot fit into a uint64
// or x is not finite, the bool will be false. This may be used to convert a
// decimal representing a monetary to its most basic unit (e.g., $123.45 to 12345
// cents.)
func Mantissa(x *decimal.Big) (uint64, bool) {
	mp, _ := decimal.Raw(x)
	return *mp, x.IsFinite() && *mp != c.Inflated
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

// MinAbs returns the lesser of the absolute value of the provided values. The
// result is undefined if no values are provided.
func MinAbs(x ...*decimal.Big) *decimal.Big {
	m := x[0]
	for _, v := range x[1:] {
		if v.CmpAbs(m) < 0 {
			m = v
		}
	}
	return m
}

// maxfor sets z to 999...n with the provided sign.
func maxfor(z *big.Int, n, sign int) {
	arith.Sub(z, arith.BigPow10(uint64(n)), 1)
	if sign < 0 {
		z.Neg(z)
	}
}

// NextMinus sets z to the smallest representable number that's smaller than x
// and returns z. If x is negative infinity the result will be negative infinity.
// If the result is zero its sign will be negative and its scale will be MinScale.
func NextMinus(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	if x.IsInf(0) {
		if x.IsInf(-1) {
			return z.SetInf(true)
		}
		_, m := decimal.Raw(z)
		maxfor(m, precision(z), +1)
		return z.SetBigMantScale(m, -etop(z))
	}

	ctx := z.Context
	ctx.RoundingMode = decimal.ToNegativeInf
	ctx.Set(z, x)
	ctx.Sub(z, x, new(decimal.Big).SetMantScale(1, -etiny(z)+1))
	z.Context.Conditions &= ctx.Conditions
	return z
}

// NextPlus sets z to the largest representable number that's larger than x and
// returns z. If x is positive infinity the result will be positive infinity. If
// the result is zero it will be positive and its scale will be MaxScale.
func NextPlus(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	if x.IsInf(0) {
		if x.IsInf(+1) {
			return z.SetInf(false)
		}
		_, m := decimal.Raw(z)
		maxfor(m, precision(z), -1)
		return z.SetBigMantScale(m, -etop(z))
	}

	ctx := z.Context
	ctx.RoundingMode = decimal.ToPositiveInf
	ctx.Set(z, x)
	ctx.Add(z, x, new(decimal.Big).SetMantScale(1, -etiny(z)+1))
	z.Context.Conditions &= ctx.Conditions
	return z
}

func ord(x *decimal.Big, abs bool) (r int) {
	// -2 == -qnan
	// -1 == -snan
	//  0 == not nan
	// +1 == snan
	// +2 == qnan
	if x.IsNaN(0) {
		if x.IsNaN(+1) { // qnan
			r = +2
		} else {
			r = +1
		}
		if !abs && x.Signbit() {
			r = -r
		}
	}
	return r
}

func precision(z *decimal.Big) int {
	p := z.Context.Precision
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

// SameQuantum returns true if x and y have the same exponent (scale).
func SameQuantum(x, y *decimal.Big) bool { return x.Scale() == y.Scale() }

// SetSignbit sets z to -z if sign is true, otherwise to +z.
func SetSignbit(z *decimal.Big, sign bool) *decimal.Big {
	if sign {
		return z.CopySign(z, neg)
	}
	return z.CopySign(z, pos)
}
