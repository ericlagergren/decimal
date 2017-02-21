// Package decimal provides a high-performance, arbitrary precision,
// fixed-point decimal library.
//
// The following type is supported:
//
//     Big decimal numbers
//
// The zero value for a Big corresponds with 0. Its method naming is the same
// as math/big's, meaning:
//
//     func (z *T) SetV(v V) *T          // z = v
//     func (z *T) Unary(x *T) *T        // z = unary x
//     func (z *T) Binary(x, y *T) *T    // z = x binary y
//     func (x *T) Pred() P              // p = pred(x)
//
// In general, its conventions will mirror math/big's.
//
// In general, operations that use the receiver z as storage will not modify
// z's Context. Additionally, the arguments to Binary and Unary methods are
// allowed to alias, so the following is valid:
//
//     x := New(1, 0)
//     x.Add(x, x) // x == 2
//
// Compared to other decimal libraries, this package:
//
//     1. Does not have signals* or traps
//     2. Panics on NaN values† (e.g., +Inf + -Inf, x / 0)
//     3. Does not make a distinction between 0 and -0
//     4. Has only mutable decimals (for efficiency's sake)
//
//     *: see #2
//     †: and usually sets those values to 0
package decimal

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"runtime"
	"strconv"
	"strings"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/c"
)

// Note: For +/-inf/nan checks: https://play.golang.org/p/RtH3UCt5IH

// Big is a fixed-point, arbitrary-precision decimal number.
//
// A Big decimal is a number and a scale, the latter representing the number
// of digits following the radix if the scale is >= 0. Otherwise, it's the
// number * 10 ^ -scale.
type Big struct {
	// Big is laid out like this so it takes up as little memory as possible.
	// On 64-bit systems it takes up 42 bytes.
	// On 32-bit systems it takes up 32 bytes.
	//
	// compact is use if the value fits into an int64. The scale does not
	// affect whether this field is used; typically, if a decimal has <= 19
	// digits this field will be used.
	compact int64

	// scale is the number of digits following the radix. If scale is negative
	// the inflation is implied; neither the compact nor unscaled fields are
	// actually inflated.
	scale int32

	ctx      Context
	form     form
	unscaled big.Int
}

// form represents whether the Big decimal is normal, infinite, or NaN.
type form byte

const (
	zero   form = 0 // this constant must remain 0.
	finite form = 1

	// Reserve the top three bits for Inf state:
	//
	//    00000100 = Inf
	//    00000110 = +Inf
	//    00000101 = -Inf
	//
	// *Never* assign inf, only pinf and ninf.
	//
	// Note: if mode forms are added various form operations will need to be
	// revisited. Right now there are some operations running with the
	// assumption that the only state the form can be in is:
	//
	// 	- zero
	// 	- finite
	// 	- pinf
	// 	- ninf
	//
	inf  form = 1 << 5
	pinf form = inf | 1<<6
	ninf form = inf | 1<<7
)

//go:generate stringer -type=form

// An ErrNaN panic is raised by a Decimal operation that would lead to a NaN
// under IEEE-754 rules. An ErrNaN implements the error interface.
type ErrNaN struct {
	// TODO: Perhaps use math/big.ErrNaN if possible in the future?
	Msg string
}

func (e ErrNaN) Error() string {
	return e.Msg
}

// These methods are here to prevent typos.

func (x *Big) isInflated() bool {
	return x.compact == c.Inflated
}

func (x *Big) isCompact() bool {
	return x.compact != c.Inflated
}

func (x *Big) isEvenInt() bool {
	return x.IsInt() &&
		(x.isCompact() && x.compact&1 == 0) ||
		(x.isInflated() && x.unscaled.And(&x.unscaled, oneInt).Sign() == 0)
}

// New creates a new Big decimal with the given value and scale. For example:
//
//  	New(1234, 3) // 1.234
//  	New(42, 0)   // 42
//  	New(4321, 5) // 0.04321
//  	New(-1, 0)   // -1
//  	New(3, -10)  // 30,000,000,000
//
func New(value int64, scale int32) *Big {
	return new(Big).SetMantScale(value, scale)
}

// Abs sets z to the absolute value of x if x is finite and returns z.
func (z *Big) Abs(x *Big) *Big {
	if x.form != finite {
		return z
	}
	if x.isCompact() {
		z.compact = arith.Abs(x.compact)
	} else {
		z.unscaled.Abs(&x.unscaled)
	}
	z.scale = x.scale
	z.form = finite
	return z
}

// Add sets z to x + y and returns z.
func (z *Big) Add(x, y *Big) *Big {
	if x.form == finite && y.form == finite {
		z.form = finite
		if x.isCompact() {
			if y.isCompact() {
				return z.addCompact(x, y)
			}
			return z.addMixed(x, y)
		}
		if y.isCompact() {
			return z.addMixed(y, x)
		}
		return z.addBig(x, y)
	}

	if (x.form&y.form)&inf == inf && x.form&pinf != y.form&pinf {
		// +Inf + -Inf
		// -Inf + +Inf
		z.form = zero
		panic(ErrNaN{"addition of infinities with opposing signs"})
	}

	if x.form == zero && y.form == zero {
		// ±0 + ±0
		z.form = zero
		return z
	}

	if x.form&inf != 0 || y.form == zero {
		// ±Inf + y
		// x + ±0
		return z.Set(x)
	}

	// ±0 + y
	// x + ±Inf
	return z.Set(y)
}

// addCompact sets z to x + y and returns z.
func (z *Big) addCompact(x, y *Big) *Big {
	// Fast path: if the scales are the same we can just add
	// without adjusting either number.
	if x.scale == y.scale {
		z.scale = x.scale
		sum, ok := checked.Add(x.compact, y.compact)
		if ok {
			z.compact = sum
			if sum == 0 {
				z.form = zero
			}
		} else {
			z.unscaled.Add(big.NewInt(x.compact), big.NewInt(y.compact))
			z.compact = c.Inflated
			if z.unscaled.Sign() == 0 {
				z.form = zero
			}
		}
		return z
	}

	// Guess the scales. We need to inflate lo.
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	// Power of 10 we need to multiply our lo value by in order
	// to equalize the scales.
	inc := hi.scale - lo.scale
	z.scale = hi.scale

	scaledLo, ok := checked.MulPow10(lo.compact, inc)
	if ok {
		sum, ok := checked.Add(hi.compact, scaledLo)
		if ok {
			z.compact = sum
			return z
		}
	}
	scaled := checked.MulBigPow10(big.NewInt(lo.compact), inc)
	z.unscaled.Add(scaled, big.NewInt(hi.compact))
	z.compact = c.Inflated
	if z.unscaled.Sign() == 0 {
		z.form = zero
	}
	return z
}

// addMixed adds a compact Big with a non-compact Big.
// addMixed will panic if the first argument is not compact.
func (z *Big) addMixed(comp, non *Big) *Big {
	if comp.isInflated() {
		panic("decimal.Add (bug) comp.isInflated() == true")
	}
	if comp.scale == non.scale {
		z.unscaled.Add(big.NewInt(comp.compact), &non.unscaled)
		z.scale = comp.scale
		z.compact = c.Inflated
		if z.unscaled.Sign() == 0 {
			z.form = zero
		}
		return z
	}
	// Since we have to rescale we need to add two big.Ints together because
	// big.Int doesn't have an API for increasing its value by an integer.
	return z.addBig(&Big{
		unscaled: *big.NewInt(comp.compact),
		scale:    comp.scale,
	}, non)
}

func (z *Big) addBig(x, y *Big) *Big {
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	inc := hi.scale - lo.scale
	tmp := new(big.Int).Set(&lo.unscaled)
	scaled := checked.MulBigPow10(tmp, inc)
	z.unscaled.Add(&hi.unscaled, scaled)
	z.compact = c.Inflated
	z.scale = hi.scale
	if z.unscaled.Sign() == 0 {
		z.form = zero
	}
	return z
}

// log2(10)
const ln210 = 3.321928094887362347870319429489390175864831393024580612054

// BitLen returns the absolute value of x in bits. The result is undefined if
// x is an infinity.
func (x *Big) BitLen() int {
	if x.form != finite {
		return 0
	}

	// If using an artificially inflated number determine the
	// bitlen using the number of digits.
	//
	// http://www.exploringbinary.com/number-of-bits-in-a-decimal-integer/
	if x.scale < 0 {
		// Number of zeros in scale + digits in z.
		d := -int(x.scale) + x.Prec()
		return int(math.Ceil(float64(d-1) * ln210))
	}
	if x.isCompact() {
		return arith.BitLen(x.compact)
	}
	return x.unscaled.BitLen()
}

// Cmp compares d and x and returns:
//
//   -1 if z <  x
//    0 if z == x
//   +1 if z >  x
//
// It does not modify d or x.
func (z *Big) Cmp(x *Big) int {
	// Check for same pointers.
	if z == x {
		return 0
	}

	// Fast path: different signs. Catches non-finite forms like zero and
	// ±Inf.
	zs := z.Sign()
	xs := x.Sign()
	switch {
	case zs > xs:
		return +1
	case zs < xs:
		return -1
	case zs == 0 && xs == 0:
		return 0
	}

	// zs == xs

	// Same scales means we can compare straight across.
	if z.scale == x.scale {
		switch {
		case z.isCompact() && x.isCompact():
			if z.compact > x.compact {
				return +1
			}
			if z.compact < x.compact {
				return -1
			}
			return 0
		case z.isInflated() && x.isInflated():
			return z.unscaled.Cmp(&x.unscaled)
		default:
			// The inflated number is more than likely larger, but I'm not 100%
			// certain that inflated > compact is an invariant.
			zu, xu := &z.unscaled, &x.unscaled
			if z.isCompact() {
				zu = big.NewInt(z.compact)
			} else {
				xu = big.NewInt(x.compact)
			}
			return zu.Cmp(xu)
		}
	}

	// Signs are the same and the scales differ. Compare the lengths of their
	// integral parts; if they differ in length one number is larger.
	// E.g., 1234.01
	//        123.011
	zl := z.Prec() - int(z.scale)
	xl := x.Prec() - int(x.scale)
	if zl > xl {
		return +1
	}
	if zl < xl {
		return -1
	}

	// We have to inflate one of the numbrers. Designate z as hi and x as lo.
	var (
		// hi
		hi = z.scale
		zm = &z.unscaled
		zc = z.compact

		// lo
		lo = x.scale
		xm = &x.unscaled
		xc = x.compact
	)

	swap := hi < lo
	if swap {
		// z is now lo
		zc, xc = xc, zc
		zm, xm = xm, zm
		hi, lo = lo, hi
	}

	diff, ok := checked.Sub32(hi, lo)
	if !ok {
		panic("!ok")
	}

	// Inflate lo.
	if xc != c.Inflated {
		nx, ok := checked.MulPow10(xc, diff)
		if !ok {
			// Can't fit in an int64, use big.Int.
			xm = checked.MulBigPow10(big.NewInt(xc), diff)
			xc = c.Inflated
		} else {
			xc = nx
		}
	} else {
		xm = checked.MulBigPow10(xm, diff)
	}

	if swap {
		zc, xc = xc, zc
		zm, xm = xm, zm
	}

	if zc != c.Inflated {
		if xc != c.Inflated {
			return arith.AbsCmp(zc, xc)
		}
		return big.NewInt(zc).Cmp(xm)
	}
	if xc != c.Inflated {
		return z.unscaled.Cmp(big.NewInt(xc))
	}
	return z.unscaled.Cmp(xm)
}

// Context returns x's Context.
func (x *Big) Context() Context {
	return x.ctx
}

// Copy sets z to a copy of x and returns z.
func (z *Big) Copy(x *Big) *Big {
	if z != x {
		z.compact = x.compact
		z.ctx = x.ctx
		z.form = x.form
		z.scale = x.scale

		// Copy over unscaled if need be.
		if x.isInflated() {
			z.unscaled.Set(&x.unscaled)
		}
	}
	return z
}

// Format implements the fmt.Formatter interface.
// func (z *Big) Format(s fmt.State, r rune) {
// 	switch r {
// 	case 'e', 'g', 's', 'f':
// 		s.Write([]byte(z.String()))
// 	case 'E':
// 		s.Write([]byte(z.toString(true, upper)))
// 	default:
// 		fmt.Fprint(s, *z)
// 	}
// }

// IsBig returns true if x, with its fractional part truncated, cannot fit
// inside an int64. If x is an infinity the result is undefined.
func (x *Big) IsBig() bool {
	// x.form != finite == 0 or infinity
	if x.form != finite {
		return false
	}
	// x.scale <= -19 is too large for an int64.
	if x.scale <= -19 {
		return true
	}

	var v int64
	if x.isCompact() {
		if x.scale == 0 {
			return false
		}
		v = x.compact
	} else {
		if x.unscaled.Cmp(c.MinInt64) <= 0 || x.unscaled.Cmp(c.MaxInt64) > 0 {
			return true
		}
		// Repeat this line twice so we don't have to call x.unscaled.Int64.
		if x.scale == 0 {
			return false
		}
		v = x.unscaled.Int64()
	}
	_, ok := scalex(v, x.scale)
	return !ok
}

// Int returns x as a big.Int, truncating the fractional portion, if any. If
// x is an infinity the result is undefined.
func (x *Big) Int() *big.Int {
	if x.form != finite {
		return big.NewInt(0)
	}

	var b big.Int
	if x.isCompact() {
		b.SetInt64(x.compact)
	} else {
		b.Set(&x.unscaled)
	}
	if x.scale == 0 {
		return &b
	}
	if x.scale < 0 {
		return checked.MulBigPow10(&b, -x.scale)
	}
	p := pow.BigTen(int64(x.scale))
	return b.Div(&b, &p)
}

// Int64 returns x as an int64, truncating the fractional portion, if any. The
// result is undefined if x is an infinity or if x does not fit inside an
// int64.
func (x *Big) Int64() int64 {
	if x.form != finite {
		return 0
	}

	var b int64
	if x.isCompact() {
		b = x.compact
	} else {
		b = x.unscaled.Int64()
	}
	if x.scale == 0 {
		return b
	}
	b, ok := scalex(b, x.scale)
	if !ok {
		return 0
	}
	return b
}

// IsFinite returns true if x is finite.
func (x *Big) IsFinite() bool {
	return x.form == finite
}

// IsInf returns true if x is an infinity according to sign.
// If sign > 0, IsInf reports whether x is positive infinity.
// If sign < 0, IsInf reports whether x is negative infinity.
// If sign == 0, IsInf reports whether x is either infinity.
func (x *Big) IsInf(sign int) bool {
	return sign >= 0 && x.form&pinf == pinf || sign <= 0 && x.form == ninf
}

// IsInt reports whether x is an integer. Inf values are not integers.
func (x *Big) IsInt() bool {
	if x.form != finite {
		return x.form == zero
	}
	// The x.Cmp(one) check is necessary because x might be a decimal *and*
	// Prec <= 0 if x < 1.
	//
	// E.g., 0.1:  scale == 1, prec == 1
	//       0.01: scale == 2, prec == 1
	return x.scale <= 0 || (x.Prec() <= int(x.scale) && x.Cmp(one) > 0)
}

// Log sets z to the base-e logarithm of x and returns z.
/*func (z *Big) Log(x *Big) *Big {
	if x.ltez() {
		panic(ErrNaN{"base-e logarithm of x <= 0"})
	}
	if x.form &inf!=0 {
		z.form = inf
		return z
	}
	mag := int64(x.Prec() - int(x.scale) - 1)
	if mag < 3 {
		return z.logNewton(x)
	}
	root := z.integralRoot(x, mag)
	lnRoot := root.logNewton(root)
	return z.Mul(New(mag, 0), lnRoot)
}*/

// MarshalText implements encoding/TextMarshaler.
func (x *Big) MarshalText() ([]byte, error) {
	return x.format(true, lower), nil
}

// Mode returns the rounding mode of x.
func (x *Big) Mode() RoundingMode {
	return x.ctx.mode
}

// Mul sets z to x * y and returns z.
func (z *Big) Mul(x, y *Big) *Big {
	if x.form == finite && y.form == finite {
		z.form = finite
		if x.isCompact() {
			if y.isCompact() {
				return z.mulCompact(x, y)
			}
			return z.mulMixed(x, y)
		}
		if y.isCompact() {
			return z.mulMixed(y, x)
		}
		return z.mulBig(x, y)
	}

	if x.form == zero && y.form&inf != 0 || x.form&inf != 0 && y.form == 0 {
		// 0 * ±Inf
		// ±Inf * 0
		z.form = zero
		panic(ErrNaN{"multiplication of zero with infinity"})
	}

	if x.form&inf != 0 || y.form&inf != 0 {
		// ±Inf * y
		// x * ±Inf
		if x.Sign() != y.Sign() {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z
	}

	// 0 * y
	// x * 0
	z.form = zero
	return z
}

func (z *Big) mulCompact(x, y *Big) *Big {
	scale, ok := checked.Add32(x.scale, y.scale)
	if !ok {
		// x + -y ∈ {-1<<31, ..., 1<<31-1}
		if x.scale > 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}

	prod, ok := checked.Mul(x.compact, y.compact)
	if ok {
		z.compact = prod
	} else {
		z.unscaled.Mul(big.NewInt(x.compact), big.NewInt(y.compact))
		z.compact = c.Inflated
	}
	z.scale = scale
	z.form = finite
	return z
}

func (z *Big) mulMixed(comp, non *Big) *Big {
	if comp.isInflated() {
		panic("decimal.Mul (bug) comp.isInflated() == true")
	}
	if comp.scale == non.scale {
		scale, ok := checked.Add32(comp.scale, non.scale)
		if !ok {
			// x + -y ∈ {-1<<31, ..., 1<<31-1}
			if comp.scale > 0 {
				z.form = pinf
			} else {
				z.form = ninf
			}
			return z
		}
		z.unscaled.Mul(big.NewInt(comp.compact), &non.unscaled)
		z.compact = c.Inflated
		z.scale = scale
		z.form = finite
		return z
	}
	return z.mulBig(&Big{
		unscaled: *big.NewInt(comp.compact),
		scale:    comp.scale,
	}, non)
}

func (z *Big) mulBig(x, y *Big) *Big {
	scale, ok := checked.Add32(x.scale, y.scale)
	if !ok {
		// x + -y ∈ {-1<<31, ..., 1<<31-1}
		if x.scale > 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}
	z.unscaled.Mul(&x.unscaled, &y.unscaled)
	z.compact = c.Inflated
	z.scale = scale
	z.form = finite
	return z
}

// Neg sets z to -x and returns z.
func (z *Big) Neg(x *Big) *Big {
	if x.form&inf != 0 {
		// x.form is either 110 or 101
		//
		// 110 ⊕ 011 = 101
		// 101 ⊕ 011 = 110
		z.form ^= (pinf | ninf) ^ inf
		return z
	}
	if x.isCompact() {
		z.compact = -x.compact
	} else {
		z.unscaled.Neg(&x.unscaled)
		z.compact = c.Inflated
	}
	z.scale = x.scale
	z.form = x.form
	return z
}

// Prec returns the precision of x. That is, it returns the number of digits
// in the unscaled form of x. x == 0 has a precision of 1. The result is
// undefined if x is an infinity.
func (x *Big) Prec() int {
	if x.form&inf != 0 {
		return 0
	}
	if x.form == zero {
		return 1
	}
	if x.isCompact() {
		return arith.Length(x.compact)
	}
	return arith.BigLength(&x.unscaled)
}

// Quo sets z to x / y and returns z.
func (z *Big) Quo(x, y *Big) *Big {
	if x.form == finite && y.form == finite {
		z.form = finite
		// x / y (common case)
		if x.isCompact() {
			if y.isCompact() {
				return z.quoCompact(x, y)
			}
			return z.quoBig(&Big{
				compact:  c.Inflated,
				unscaled: *big.NewInt(x.compact),
				ctx:      x.ctx,
				form:     x.form,
				scale:    x.scale,
			}, y)
		}
		if y.isCompact() {
			return z.quoBig(x, &Big{
				compact:  c.Inflated,
				unscaled: *big.NewInt(y.compact),
				ctx:      y.ctx,
				form:     y.form,
				scale:    y.scale,
			})
		}
		return z.quoBig(x, y)
	}

	if x.form^y.form == zero || x.form&y.form != 0 {
		// 0 / 0
		// ±Inf / ±Inf
		z.form = zero
		panic(ErrNaN{"division of zero by zero or infinity by infinity"})
	}

	if x.form == zero || y.form&inf != 0 {
		// 0 / y
		// x / ±Inf
		z.form = zero
		return z
	}

	// x / 0
	// ±Inf / y

	// The spec requires the resulting infinity's sign to match
	// the  "exclusive or of the signs of the operands."
	// http://speleotrove.com/decimal/daops.html#refdivide
	//
	// Since we do not have -0, y's sign is always 1.
	if x.Signbit() {
		z.form = ninf
	} else {
		z.form = pinf
	}

	if y.form == zero {
		// Panic with ErrNaN since x / 0 is technically undefined.
		panic(ErrNaN{"division by zero"})
	}
	return z
}

func (z *Big) quoAndRound(x, y int64) *Big {
	// Quotient
	z.compact = x / y

	// ToZero means we can ignore remainder.
	if z.ctx.mode == ToZero {
		return z
	}

	// Remainder
	r := x % y

	sign := int64(1)
	if (x < 0) != (y < 0) {
		sign = -1
	}
	if r != 0 && z.needsInc(y, r, sign > 0, z.compact&1 != 0) {
		z.compact += sign
	}
	return z.Round(z.Context().Precision())
}

func (z *Big) quoCompact(x, y *Big) *Big {
	scale, ok := checked.Sub32(x.scale, y.scale)
	if !ok {
		// -x - y ∈ {-1<<31, ..., 1<<31-1}
		if x.scale < 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}

	zp := z.Context().Precision()
	xp := int32(x.Prec())
	yp := int32(y.Prec())

	// Multiply y by 10 if x' > y'
	if cmpNorm(x.compact, xp, y.compact, yp) {
		yp--
	}

	scale, ok = checked.Int32(int64(scale) + int64(yp) - int64(xp) + int64(zp))
	if !ok {
		// The wraparound from int32(int64(x)) where x ∉ {-1<<31, ..., 1<<31-1}
		// will swap its sign.
		//
		// TODO: for some reason I am not 100% sure the above accurate.
		if scale > 0 {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z
	}
	z.scale = scale

	shift, ok := checked.SumSub(zp, yp, xp)
	if !ok {
		// TODO: See above comment about wraparound.
		if scale > 0 {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z
	}

	xs, ys := x.compact, y.compact
	if shift > 0 {
		xs, ok = checked.MulPow10(x.compact, shift)
		if !ok {
			x0 := checked.MulBigPow10(big.NewInt(x.compact), shift)
			return z.quoBigAndRound(x0, big.NewInt(y.compact))
		}
		return z.quoAndRound(xs, ys)
	}

	// shift < 0
	ns, ok := checked.Sub32(xp, zp)
	if !ok {
		// -x - y ∈ {-1<<31, ..., 1<<31-1}
		if xp < 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}

	// new scale == yp, so no inflation needed.
	if ns == yp {
		return z.quoAndRound(xs, ys)
	}
	shift, ok = checked.Sub32(ns, yp)
	if !ok {
		// -x - y ∈ {-1<<31, ..., 1<<31-1}
		if ns < 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}
	ys, ok = checked.MulPow10(ys, shift)
	if !ok {
		y0 := checked.MulBigPow10(big.NewInt(y.compact), shift)
		return z.quoBigAndRound(big.NewInt(x.compact), y0)
	}
	return z.quoAndRound(xs, ys)
}

func (z *Big) quoBig(x, y *Big) *Big {
	scale, ok := checked.Sub32(x.scale, y.scale)
	if !ok {
		// -x - y ∈ {-1<<31, ..., 1<<31-1}
		if x.scale < 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}

	zp := z.Context().Precision()
	xp := int32(x.Prec())
	yp := int32(y.Prec())

	// Multiply y by 10 if x' > y'
	if cmpNormBig(&x.unscaled, xp, &y.unscaled, yp) {
		yp--
	}

	scale, ok = checked.Int32(int64(scale) + int64(yp) - int64(xp) + int64(zp))
	if !ok {
		// The wraparound from int32(int64(x)) where x ∉ {-1<<31, ..., 1<<31-1}
		// will swap its sign.
		//
		// TODO: for some reason I am not 100% sure the above accurate.
		if scale > 0 {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z
	}
	z.scale = scale

	shift, ok := checked.SumSub(zp, yp, xp)
	if !ok {
		// TODO: See above comment about wraparound.
		if scale > 0 {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z
	}
	if shift > 0 {
		xs := checked.MulBigPow10(new(big.Int).Set(&x.unscaled), shift)
		return z.quoBigAndRound(xs, &y.unscaled)
	}

	// shift < 0
	ns, ok := checked.Sub32(xp, zp)
	if !ok {
		// -x - y ∈ {-1<<31, ..., 1<<31-1}
		if xp < 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}
	shift, ok = checked.Sub32(ns, yp)
	if !ok {
		// -x - y ∈ {-1<<31, ..., 1<<31-1}
		if ns < 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}
	ys := checked.MulBigPow10(new(big.Int).Set(&y.unscaled), shift)
	return z.quoBigAndRound(&x.unscaled, ys)
}

func (z *Big) quoBigAndRound(x, y *big.Int) *Big {
	z.compact = c.Inflated

	// TODO: perhaps use a pool for the allocated big.Int?
	q, r := z.unscaled.QuoRem(x, y, new(big.Int))

	if z.ctx.mode == ToZero {
		return z
	}

	sign := int64(1)
	if (x.Sign() < 0) != (y.Sign() < 0) {
		sign = -1
	}
	odd := new(big.Int).And(q, oneInt).Sign() != 0

	if r.Sign() != 0 && z.needsIncBig(y, r, sign > 0, odd) {
		z.unscaled.Add(&z.unscaled, big.NewInt(sign))
	}
	return z.Round(z.Context().Precision())
}

// Raw directly returns x's raw compact and unscaled values. Caveat emptor:
// Neither are guaranteed to be valid. Raw is intended to support missing
// functionality outside this package and generally should be avoided.
// Additionally, Raw is the only part of this package's API which is not
// guaranteed to remain stable. This means the function could change or
// disappear at any time, even across minor version numbers.
func Raw(x *Big) (int64, *big.Int) {
	return x.compact, &x.unscaled
}

// Round rounds z down to n digits of precision and returns z. The result is
// undefined if n < 0 or z is an infinity. No rounding will occur if n == 0.
// The result of Round will always be within the interval [⌊z⌋, z].
func (z *Big) Round(n int32) *Big {
	zp := z.Prec()
	if n <= 0 || int(n) >= zp || z.form != finite {
		return z
	}
	z.SetPrecision(n)

	shift, ok := checked.Sub(int64(zp), int64(n))
	if !ok {
		// -x - y ∈ {-1<<63, ..., 1<<63-1}
		if zp < 0 {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}
	if shift <= 0 {
		return z
	}
	z.scale -= int32(shift)

	if z.isCompact() {
		val, ok := pow.Ten64(shift)
		if ok {
			return z.quoAndRound(z.compact, val)
		}
		z.unscaled.SetInt64(z.compact)
	}
	val := pow.BigTen(shift)
	return z.quoBigAndRound(&z.unscaled, &val)
}

// Scale returns x's scale.
func (x *Big) Scale() int32 {
	return x.scale
}

// Set sets z to x and returns z. The result might be rounded depending on z's
// Context.
func (z *Big) Set(x *Big) *Big {
	if z != x {
		z.compact = x.compact
		z.form = x.form
		z.scale = x.scale

		// Copy over unscaled if need be.
		if x.isInflated() {
			z.unscaled.Set(&x.unscaled)
		}
		z.Round(z.Context().Precision())
	}
	return z
}

// SetBigMantScale sets z to the given value and scale.
func (z *Big) SetBigMantScale(value *big.Int, scale int32) *Big {
	if value.Sign() == 0 {
		z.form = zero
		return z
	}
	z.scale = scale
	z.unscaled.Set(value)
	z.form = finite
	z.compact = c.Inflated
	return z
}

// SetContext sets z's Context and returns z.
func (z *Big) SetContext(ctx Context) *Big {
	z.ctx = ctx
	return z
}

// SetFloat64 sets z to the provided float64.
//
// Remember, floating-point to decimal conversions can be lossy. For example,
// the floating-point number `0.1' appears to simply be 0.1, but its actual
// value is 0.1000000000000000055511151231257827021181583404541015625.
//
// SetFloat64 is particularly lossy because will round non-integer values.
// For example, if passed the value `3.1415' it attempts to do the same as if
// SetMantScale(31415, 4) were called.
//
// To do this, it scales up the provided number by its scale. This involves
// rounding, so approximately 2.3% of decimals created from floats will have a
// rounding imprecision of ± 1 ULP.
func (z *Big) SetFloat64(value float64) *Big {
	if value == 0 {
		z.form = 0
		return z
	}

	var scale int32

	// If value is not an integer (has a fractional part) bump its value up
	// and find the appropriate scale.
	_, fr := math.Modf(value)
	if fr != 0 {
		scale = findScale(value)
		value *= math.Pow10(int(scale))
	}

	if math.IsNaN(value) {
		panic(ErrNaN{"SetFloat64(NaN)"})
	}
	if math.IsInf(value, +1) {
		z.form = pinf
		return z
	}
	if math.IsInf(value, -1) {
		z.form = ninf
		return z
	}

	// Given float64(math.MaxInt64) == math.MaxInt64.
	if value <= math.MaxInt64 {
		z.compact = int64(value)
	} else {
		if value <= math.MaxUint64 {
			z.unscaled.SetUint64(uint64(value))
		} else {
			z.unscaled.Set(bigIntFromFloat(value))
		}
		z.compact = c.Inflated
	}
	z.scale = scale
	z.form = finite
	return z
}

// SetInf sets z to -Inf if signbit is set or +Inf is signbit is not set, and
// returns z.
func (x *Big) SetInf(signbit bool) *Big {
	if signbit {
		x.form = ninf
	} else {
		x.form = pinf
	}
	return x
}

// SetMantScale sets z to the given value and scale.
func (z *Big) SetMantScale(value int64, scale int32) *Big {
	if value == 0 {
		z.form = zero
		return z
	}
	z.scale = scale
	if value == c.Inflated {
		z.unscaled.SetInt64(value)
	}
	z.compact = value
	z.form = finite
	return z
}

// SetMode sets z's RoundingMode to mode and returns z.
func (z *Big) SetMode(mode RoundingMode) *Big {
	z.ctx.mode = mode
	return z
}

// SetPrecision sets z's precision to prec and returns z.
// This method is distinct from Prec. This sets the internal context which
// dictates rounding and digits after the radix for lossy operations. The
// latter describes the number of digits in the decimal.
func (z *Big) SetPrecision(prec int32) *Big {
	z.ctx.precision = prec
	return z
}

// SetScale sets z's scale to scale and returns z.
func (z *Big) SetScale(scale int32) *Big {
	z.scale = scale
	return z
}

// SetString sets z to the value of s, returning z and a bool indicating
// success. s must be a string in one of the following formats:
//
// 	1.234
// 	1234
// 	1.234e+5
// 	1.234E-5
// 	0.000001234
// 	Inf
// 	+Inf
// 	-Inf
//
// Inf values are not required to be case-sensitive and no distinction is made
// between +Inf and Inf.
func (z *Big) SetString(s string) (*Big, bool) {
	// Inf, +Inf, or -Inf.
	if strings.EqualFold(s, "Inf") || strings.EqualFold(s, "+Inf") {
		z.form = pinf
		return z, true
	}
	if strings.EqualFold(s, "-Inf") {
		z.form = ninf
		return z, true
	}

	var scale int32

	// Check for a scientific string.
	i := strings.LastIndexAny(s, "Ee")
	if i > 0 {
		eint, err := strconv.ParseInt(s[i+1:], 10, 32)
		if err != nil {
			return nil, false
		}
		s = s[:i]
		scale = -int32(eint)
	}

	switch strings.Count(s, ".") {
	case 0:
	case 1:
		i = strings.IndexByte(s, '.')
		s = s[:i] + s[i+1:]
		scale += int32(len(s) - i)
	default:
		return nil, false
	}

	var err error
	z.form = finite
	// Numbers == 19 can be out of range, but try the edge case anyway.
	if len(s) <= 19 {
		z.compact, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			nerr, ok := err.(*strconv.NumError)
			if !ok || nerr.Err == strconv.ErrSyntax {
				return nil, false
			}
			err = nerr.Err
		} else if z.compact == 0 {
			z.form = zero
		}
	}
	if (err == strconv.ErrRange && len(s) == 19) || len(s) > 19 {
		_, ok := z.unscaled.SetString(s, 10)
		if !ok {
			return nil, false
		}
		z.compact = c.Inflated
		if z.unscaled.Sign() == 0 {
			z.form = zero
		}
	}
	z.scale = scale
	return z, true
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x is 0
//	+1 if x >  0
//
func (x *Big) Sign() int {
	// x = 0
	if x.form == zero {
		return 0
	}

	// x = +Inf
	// x = -Inf
	if x.form&inf != 0 {
		if x.form&(ninf^inf) != 0 {
			return -1
		}
		return +1
	}

	// x is finite
	if x.isCompact() {
		// See: https://github.com/golang/go/issues/16203
		if runtime.GOARCH == "amd64" {
			// Hacker's Delight, page 21, section 2-8.
			// This prevents the incorrect answer for -1 << 63.
			return int((x.compact >> 63) | int64(uint64(-x.compact)>>63))
		}
		if x.compact == 0 {
			return 0
		}
		if x.compact < 0 {
			return -1
		}
		return +1
	}
	return x.unscaled.Sign()
}

// Signbit returns true if x is negative or negative infinity.
func (x *Big) Signbit() bool {
	if x.form == ninf {
		return true
	}
	if x.isCompact() {
		return x.compact < 0
	}
	return x.unscaled.Sign() < 0
}

// String returns the scientific string representation of x.
// Special cases are:
//
//  "<nil>" if x == nil
//  "Inf"   if x.IsInf()
//
func (x *Big) String() string {
	return string(x.format(true, lower))
}

const (
	lower = 0 // opts for lowercase sci notation
	upper = 1 // opts for uppercase sci notation
)

func (x *Big) format(sci bool, opts byte) []byte {
	// Special cases.
	if x == nil {
		return []byte("<nil>")
	}
	if x.IsInf(0) {
		if x.IsInf(+1) {
			return []byte("+Inf")
		}
		return []byte("-Inf")
	}

	// Keep from allocating if x == 0.
	if x.form == zero ||
		(x.isCompact() && x.compact == 0) ||
		(x.isInflated() && x.unscaled.Sign() == 0) {
		return []byte("0")
	}

	// Fast path: return our value as-is.
	if x.scale == 0 {
		if x.isInflated() {
			// math/big.MarshalText never returns an error, only nil, so there's
			// no need to check for an error. Use MarshalText instead of Append
			// because it limits us to one allocation.
			b, _ := x.unscaled.MarshalText()
			return b
		}
		// Enough for the largest/smallest numbers we can hold plus the sign.
		var buf [20]byte
		return strconv.AppendInt(buf[0:0], x.compact, 10)
	}

	// (x.scale > 0 || x.scale < 0) && x != 0

	// We have two options: The first is to always interpret x as an unsigned
	// number and selectively add the '-' if applicable. The second is to
	// format x as a signed number and do some extra math later to determine
	// where we need to place the radix, etc. depending on whether the
	// formatted number is prefixed with a '-'.
	//
	// I'm chosing the first option because it's less gross elsewhere.
	//
	// TODO: If/when this gets merged into math/big use x.unscaled.abs.utoa

	var b []byte
	if x.isInflated() {
		if x.unscaled.Sign() < 0 {
			b, _ = x.unscaled.MarshalText()
		} else {
			var buf [1]byte
			b = x.unscaled.Append(buf[0:1], 10)
		}
	} else {
		// The 20 bytes are to hold x.compact plus its sign.
		var buf [20]byte
		if x.compact < 0 {
			b = strconv.AppendInt(buf[0:0], x.compact, 10)
		} else {
			b = strconv.AppendUint(buf[0:1], uint64(x.compact), 10)
		}
	}

	if sci {
		return x.formatSci(b, opts)
	}
	return x.formatPlain(b)
}

// formatSci returns the scientific version of x. It assumes the first byte
// is either 0 (positive) or '-' (negative).
func (x *Big) formatSci(b []byte, opts byte) []byte {
	if debug && (opts < 0 || opts > 1) {
		panic("toSciString: (bug) opts != 0 || opts != 1")
	}

	// Following quotes are from:
	// http://speleotrove.com/decimal/daconvs.html#reftostr
	//
	// Note: speleotrove's spec assumes "exponent" has the reverse sign from
	// our implementation.

	adj := -int(x.scale) + (len(b) - 2)

	// "If the exponent is less than or equal to zero and the
	// adjusted exponent is greater than or equal to -6..."
	if x.scale >= 0 && adj >= -6 {
		// "...the number will be converted to a character
		// form without using exponential notation."
		return x.formatNorm(b)
	}

	// Insert our period to turn, e.g., 0.0000000056 -> 5.6e-9 if we have
	// more than one number.
	if len(b)-1 > 1 {
		b = append(b, 0)
		copy(b[2+1:], b[2:])
		b[2] = '.'
	}
	if adj != 0 {
		b = append(b, [2]byte{'e', 'E'}[opts])

		// If negative the following strconv.Append call will add the minus sign
		// for us.
		if adj > 0 {
			b = append(b, '+')
		}
		b = strconv.AppendInt(b, int64(adj), 10)
	}
	return trim(b)
}

var zeroLiteral = []byte{'0'}

// formatPlain returns the plain string version of x.
func (x *Big) formatPlain(b []byte) []byte {
	// Just unscaled + z.scale "0"s -- no radix.
	if x.scale < 0 {
		return append(b, bytes.Repeat(zeroLiteral, -int(x.scale))...)
	}
	return x.formatNorm(b)
}

// formatNorm returns the plain version of x. It's distinct from formatPlain in
// that formatPlain calls this method once it's done its own internal checks.
// Additionally, formatSci also calls this method if it does not need to add
// the {e,E} suffix. Essentially, formatNorm decides where to place the
// radix point.
func (x *Big) formatNorm(b []byte) []byte {
	switch pad := (len(b) - 1) - int(x.scale); {
	// log10(unscaled) == scale, so immediately before str.
	case pad == 0:
		b = append([]byte{b[0], '0', '.'}, b[1:]...)

	// log10(unscaled) > scale, so somewhere inside str.
	case pad > 0:
		b = append(b, 0)
		copy(b[1+pad+1:], b[1+pad:])
		b[1+pad] = '.'

	// log10(unscaled) < scale, so before p "0s" and before str.
	default:
		b0 := append([]byte{b[0], '0', '.'}, bytes.Repeat(zeroLiteral, -pad)...)
		b = append(b0, b[1:]...)
	}
	return trim(b)
}

// trim remove unnecessary bytes from b. Unnecessary bytes are defined as:
//
// 	1) Trailing '0' bytes
// 	2) A trailing '.' after step #1
// 	3) Leading 0 bytes
//
func trim(b []byte) []byte {
	s, e := 0, len(b)-1
	for ; e >= 0; e-- {
		if b[e] != '0' {
			break
		}
	}
	if b[e] == '.' {
		e--
	}
	for ; s < e; s++ {
		if b[s] != 0 {
			break
		}
	}
	return b[s : e+1]
}

// Sub sets z to x - y and returns z.
func (z *Big) Sub(x, y *Big) *Big {
	if x.form == finite && y.form == finite {
		// TODO: Write this without using Neg to save an allocation.
		return z.Add(x, new(Big).Neg(y))
	}

	if x.form&inf != 0 && x.form == y.form {
		// +Inf - +Inf
		// -Inf - -Inf
		z.form = zero
		panic(ErrNaN{"subtraction of infinities with equal signs"})
	}

	if x.form == zero && y.form == zero {
		// ±0 - ±0
		z.form = zero
		return z
	}

	if x.form&inf != 0 || y.form == zero {
		// ±Inf - y
		// x - ±0
		return z.Set(x)
	}

	// ±0 - y
	// x - ±Inf
	return z.Neg(y)
}

// UnmarshalText implements encoding/TextUnmarshaler.
func (z *Big) UnmarshalText(data []byte) error {
	_, ok := z.SetString(string(data))
	if !ok {
		return errors.New("Big.UnmarshalText: invalid decimal format")
	}
	return nil
}
