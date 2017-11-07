// Package misc contains miscellaneous decimal routes.
package misc

import (
	"errors"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/internal/arith/pow"
)

var (
	pos = decimal.New(+1, 0)
	neg = decimal.New(-1, 0)
)

const (
	// Radix is the base in which decimal arithmetic is effected.
	Radix = 10

	// IsCanonical is true since we always normalize our Big decimals.
	IsCanonical = true
)

// Canonical sets z to the canonical form of z. Since Big values are always
// canonical, it's identical to Copy.
func Canonical(z, x *decimal.Big) *decimal.Big { return z.Copy(x) }

// TODO(eric): these...
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
// ordering of all abstract representations. In particular, this means NaN
// values have a defined ordering. From lowest to highest the ordering is:
//
//  -NaN
//  -sNaN
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

// CmpTotalAbs is like CmpTotal but instead compares |x| and |y|.
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

// CopyAbs is like Abs but no flags are changed (i.e., NaN values are accepted).
func CopyAbs(z, x *decimal.Big) *decimal.Big {
	if x.IsNaN(0) {
		z.CopySign(x, pos)
	} else {
		z.Abs(x)
	}
	return z
}

// CopyNegate is like Neg but no flags are changed (i.e., NaN values are
// negated.)
func CopyNegate(z, x *decimal.Big) *decimal.Big {
	if x.IsNaN(0) {
		z.CopySign(x, neg)
	} else {
		z.Neg(x)
	}
	return z
}



// Max returns the greater of the provided values. The result is undefined if no
// values are are provided.
func Max(x ...*decimal.Big) *decimal.Big {
	m := x[0]
	for _, v := range x[1:] {
		if v.Cmp(m) > 0 {
			m = v
		}
	}
	return m
}

// MaxAbs returns the greater of the absolute value of the provided values. The
// result is undefined if no values are provided.
func MaxAbs(x ...*decimal.Big) *decimal.Big {
	m := x[0]
	for _, v := range x[1:] {
		if v.CmpAbs(m) > 0 {
			m = v
		}
	}
	return m
}

// Min returns the lesser of the provided values. The result is undefined if no
// values are are provided.
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

func precision(x *decimal.Big) int {
	p := x.Context.Precision
	if p == 0 {
		return decimal.DefaultPrecision
	}
	if p < 0 && p != decimal.UnlimitedPrecision {
		p = -p
	}
	return p
}

// SameQuantum returns true if x and y have the same exponent (scale).
func SameQuantum(x, y *decimal.Big) bool { return x.Scale() == y.Scale() }

// Shift sets z to the digit-wise shifted mantissa of x. A positive shift value
// indicates a shift to the right; a negative shift value indicates a shift to
// the left. The shift value must of equal or lesser magnitude than z's
// precision; that is, it must be in the range [-precision, precision]. The
// result is undefined if x's precision is decimal.UnlimitedPrecision.
func Shift(z, x *decimal.Big, shift int) *decimal.Big {
	// TODO(eric): allow shifts with a negative scale?

	if x.Scale() != 0 {
		return z.Signal(
			decimal.InvalidOperation,
			errors.New("shift with a non-zero scale"))
	}

	if shift == 0 {
		return z.Set(x) // no shift
	}

	if !x.IsFinite() {
		if cond, err := decimal.CheckNaNs(x, nil, "shift"); err != nil {
			return z.Signal(cond, err) // nan
		}
		if x.IsInf(0) {
			return z.SetInf(x.IsInf(-1)) // inf
		}
		return z.SetMantScale(0, 0) // zero
	}

	zp := precision(z)
	if zp == decimal.UnlimitedPrecision {
		return z.SetMantScale(0, 0) // undefined
	}
	if arith.Abs(int64(shift)) >= int64(zp) {
		return z.SetMantScale(0, 0) // zero-filled shift is too large
	}

	// TODO(eric): add an implementation that uses x.compact and falls back to
	// x.unscaled instead of calling x.Int.

	_, unsc := decimal.Raw(z)
	xb := x.Int(unsc /* &z.unscaled */)
	xp := arith.BigLength(xb)
	if xp < zp {
		// Rescale so xb has the required length.
		checked.MulBigPow10(xb, uint64(zp-xp))
	}

	if shift < 0 {
		xb.Quo(xb, pow.BigTen(uint64(-shift))) // remove trailing N digits
	} else {
		if xp < zp {
			xb.Rem(xb, pow.BigTen(uint64(shift)))    // remove first N digits
			xb.Mul(xb, pow.BigTen(uint64(zp-shift))) // fill with zeros
		} else {
			xb.Rem(xb, pow.BigTen(uint64(zp-shift)))
			xb.Mul(xb, pow.BigTen(uint64(shift)))
		}
	}
	return z.SetBigMantScale(xb, 0)
}
