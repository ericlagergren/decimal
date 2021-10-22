// Package math implements various useful mathematical functions
// and constants.
//
// Deprecated: the functionality has been moved into the decimal
//             package itself.
package math

import "github.com/ericlagergren/decimal"

// Acos returns the arccosine, in radians, of x.
//
// Range:
//     Input: -1 <= x <= 1
//     Output: 0 <= Acos(x) <= pi
//
// Special cases:
//     Acos(NaN)  = NaN
//     Acos(±Inf) = NaN
//     Acos(x)    = NaN if x < -1 or x > 1
//     Acos(-1)   = pi
//     Acos(1)    = 0
func Acos(z, x *decimal.Big) *decimal.Big {
	return z.Context.Acos(z, x)
}

// Asin returns the arcsine, in radians, of x.
//
// Range:
//     Input: -1 <= x <= 1
//     Output: -pi/2 <= Asin(x) <= pi/2
//
// Special cases:
//		Asin(NaN)  = NaN
//		Asin(±Inf) = NaN
//		Asin(x)    = NaN if x < -1 or x > 1
//		Asin(±1)   = ±pi/2
func Asin(z, x *decimal.Big) *decimal.Big {
	return z.Context.Asin(z, x)
}

// Atan returns the arctangent, in radians, of x.
//
// Range:
//     Input: all real numbers
//     Output: -pi/2 <= Atan(x) <= pi/2
//
// Special cases:
//		Atan(NaN)  = NaN
//		Atan(±Inf) = ±x * pi/2
func Atan(z, x *decimal.Big) *decimal.Big {
	return z.Context.Atan(z, x)
}

// Atan2 calculates arctan of y/x and uses the signs of y and x to determine
// the valid quadrant
//
// Range:
//     y input: all real numbers
//     x input: all real numbers
//     Output: -pi < Atan2(y, x) <= pi
//
// Special cases:
//     Atan2(NaN, NaN)      = NaN
//     Atan2(y, NaN)        = NaN
//     Atan2(NaN, x)        = NaN
//     Atan2(±0, x >=0)     = ±0
//     Atan2(±0, x <= -0)   = ±pi
//     Atan2(y > 0, 0)      = +pi/2
//     Atan2(y < 0, 0)      = -pi/2
//     Atan2(±Inf, +Inf)    = ±pi/4
//     Atan2(±Inf, -Inf)    = ±3pi/4
//     Atan2(y, +Inf)       = 0
//     Atan2(y > 0, -Inf)   = +pi
//     Atan2(y < 0, -Inf)   = -pi
//     Atan2(±Inf, x)       = ±pi/2
//     Atan2(y, x > 0)      = Atan(y/x)
//     Atan2(y >= 0, x < 0) = Atan(y/x) + pi
//     Atan2(y < 0, x < 0)  = Atan(y/x) - pi
func Atan2(z, y, x *decimal.Big) *decimal.Big {
	return z.Context.Atan2(z, y, x)
}

// BinarySplit sets z to the result of the binary splitting formula and returns
// z. The formula is defined as:
//
//         ∞    a(n)p(0) ... p(n)
//     S = Σ   -------------------
//         n=0  b(n)q(0) ... q(n)
//
// It should only be used when the number of terms is known ahead of time. If
// start is not in [start, stop) or stop is not in (start, stop], BinarySplit
// will panic.
func BinarySplit(z *decimal.Big, ctx decimal.Context, start, stop uint64, A, P, B, Q SplitFunc) *decimal.Big {
	return decimal.BinarySplit(z, ctx, start, stop, A, P, B, Q)
}

// BinarySplitDynamic sets z to the result of the binary splitting formula. It
// should be used when the number of terms is not known ahead of time. For more
// information, See BinarySplit.
func BinarySplitDynamic(ctx decimal.Context, A, P, B, Q SplitFunc) *decimal.Big {
	return decimal.BinarySplitDynamic(ctx, A, P, B, Q)
}

// Ceil sets z to the least integer value greater than or equal
// to x and returns z.
func Ceil(z, x *decimal.Big) *decimal.Big {
	return z.Context.Ceil(z, x)
}

// Cos returns the cosine, in radians, of x.
//
// Range:
//     Input: all real numbers
//     Output: -1 <= Cos(x) <= 1
//
// Special cases:
//		Cos(NaN)  = NaN
//		Cos(±Inf) = NaN
func Cos(z, x *decimal.Big) *decimal.Big {
	return z.Context.Cos(z, x)
}

// E sets z to the mathematical constant e and returns z.
func E(z *decimal.Big) *decimal.Big {
	return z.Context.E(z)
}

// Exp sets z to e**x and returns z.
func Exp(z, x *decimal.Big) *decimal.Big {
	return z.Context.Exp(z, x)
}

// Floor sets z to the greatest integer value less than or equal
// to x and returns z.
func Floor(z, x *decimal.Big) *decimal.Big {
	return z.Context.Floor(z, x)
}

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	return z.Context.Hypot(z, p, q)
}

// Lentz sets z to the result of the continued fraction provided
// by the Generator and returns z.
//
// The continued fraction should be represented as such:
//
//                          a1
//     f(x) = b0 + --------------------
//                            a2
//                 b1 + ---------------
//                               a3
//                      b2 + ----------
//                                 a4
//                           b3 + -----
//                                  ...
//
// Or, equivalently:
//
//                  a1   a2   a3
//     f(x) = b0 + ---- ---- ----
//                  b1 + b2 + b3 + ···
//
// If terms need to be subtracted, the a_N terms should be
// negative. To compute a continued fraction without b_0, divide
// the result by a_1.
//
// If the first call to the Generator's Next method returns
// false, the result of Lentz is undefined.
//
// Note: the accuracy of the result may be affected by the
// precision of intermediate results. If larger precision is
// desired, it may be necessary for the Generator to implement
// the Lentzer interface and set a higher precision for f, Δ, C,
// and D.
func Lentz(z *decimal.Big, g Generator) *decimal.Big {
	return z.Context.Lentz(z, g)
}

// Log sets z to the natural logarithm of x and returns z.
func Log(z, x *decimal.Big) *decimal.Big {
	return z.Context.Log(z, x)
}

// Log10 sets z to the common logarithm of x and returns z.
func Log10(z, x *decimal.Big) *decimal.Big {
	return z.Context.Log10(z, x)
}

// Pi sets z to the mathematical constant pi and returns z.
func Pi(z *decimal.Big) *decimal.Big {
	return z.Context.Pi(z)
}

// Pi sets z to the mathematical constant pi and returns z.
func Pow(z, x, y *decimal.Big) *decimal.Big {
	return z.Context.Pow(z, x, y)
}

// Sin returns the sine, in radians, of x.
//
// Range:
//     Input: all real numbers
//     Output: -1 <= Sin(x) <= 1
//
// Special cases:
//     Sin(NaN) = NaN
//     Sin(Inf) = NaN
func Sin(z, x *decimal.Big) *decimal.Big {
	return z.Context.Sin(z, x)
}

// Sqrt sets z to the square root of x and returns z.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	return z.Context.Sqrt(z, x)
}

// Tan returns the tangent, in radians, of x.
//
// Range:
//     Input: -pi/2 <= x <= pi/2
//     Output: all real numbers
//
// Special cases:
//     Tan(NaN) = NaN
//     Tan(±Inf) = NaN
func Tan(z, x *decimal.Big) *decimal.Big {
	return z.Context.Tan(z, x)
}

// Wallis sets z to the result of the continued fraction provided
// by the Generator and returns z.
//
// The fraction is evaluated in a top-down manner, using the
// recurrence algorithm discovered by John Wallis. For more
// information on continued fraction representations, see the
// Lentz function.
func Wallis(z *decimal.Big, g Generator) *decimal.Big {
	return z.Context.Wallis(z, g)
}

type Contexter = decimal.Contexter
type Generator = decimal.Generator
type Lentzer = decimal.Lentzer
type SplitFunc = decimal.SplitFunc
type Term = decimal.Term
type Walliser = decimal.Walliser
