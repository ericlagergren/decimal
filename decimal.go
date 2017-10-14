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
//     1. Has signals and traps, but only if you want them
//     2. Only has mutable decimals (for efficiency's sake)
//
package decimal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/c"
	"github.com/ericlagergren/decimal/internal/parse"
)

// NOTE(eric): For +/-inf/nan checks: https://play.golang.org/p/RtH3UCt5IH

// Big is a fixed-point, arbitrary-precision decimal number.
//
// A Big decimal is a number and a scale, the latter representing the number
// of digits following the radix if the scale is >= 0. Otherwise, it's the
// number * 10 ^ -scale.
type Big struct {
	// Big is laid out like this so it takes up as little memory as possible.

	// Context is the decimal's unique contextual object.
	Context Context

	// unscaled is only used if the decimal is too large to fit in compact.
	unscaled big.Int

	// compact is use if the value fits into an int64. The scale does not
	// affect whether this field is used; typically, if a decimal has <= 19
	// digits this field will be used.
	compact int64

	// scale is the number of digits following the radix. If scale is negative
	// the inflation is implied; neither the compact nor unscaled fields are
	// actually inflated.
	scale int32

	form form
}

// form represents whether the Big decimal is zero, normal, infinite, or a
// not-a-number value.
type form uint8

const (
	// zero must stay == 0 so that decimals created as literals or with new will
	// always have a value of 0.
	zero form = 0

	sign form = 1 // do not assign this; used to check for ninf and nzero.

	// nzero == sign so v <= nzero == true for nzero and zero. An alternative
	// way of thinking about it is nzero = zero | sign. Nothing assinable should
	// be smaller than nzero.
	nzero form = sign

	finite form = 1 << 1

	snan form = 1 << 2
	qnan form = 1 << 3
	nan  form = snan | qnan // do not assign this; used to check for either NaN.

	pinf form = 1 << 4
	ninf form = pinf | sign
	inf  form = pinf // do not assign this; used to check for either infinity.
)

// String is for internal use only.
func (f form) String() string {
	if !debug {
		return strconv.Itoa(int(f))
	}
	switch f {
	case zero:
		return "+zero"
	case nzero:
		return "-zero"
	case finite:
		return "finite"
	case snan:
		return "sNaN"
	case qnan:
		return "qNaN"
	case pinf:
		return "+Inf"
	case ninf:
		return "-Inf"
	case nan:
		return "bad form: nan"
	default:
		return fmt.Sprintf("unknown form: %d", f)
	}
}

// An ErrNaN panic is raised by a decimal operation that would lead to a NaN
// under IEEE-754 rules. An ErrNaN implements the error interface.
type ErrNaN struct {
	// TODO(eric): Perhaps use math/big.ErrNaN if possible in the future?
	Msg string
}

func (e ErrNaN) Error() string {
	return e.Msg
}

// checkNaNs checks if either x or y is NaN. If so, it sets z's form to either
// qnan or snan and returns the peoper Condition along with ErrNaN.
func (z *Big) checkNaNs(x, y *Big, op string) (Condition, error) {
	f := (x.form | y.form) & nan
	if f == 0 {
		return 0, nil
	}
	var cond Condition
	if f&snan != 0 {
		cond = InvalidOperation
	}
	z.form = qnan
	return cond, ErrNaN{Msg: op + " with NaN as an operand"}
}

var (
	errOverflow  = errors.New("decimal: overflow: scale is too large")
	errUnderflow = errors.New("decimal: underflow: scale is too small")
)

func (z *Big) xflow(over, neg bool) *Big {
	// over == overflow
	// neg == intermediate result < 0
	if over {
		// NOTE(eric): in some situations, the decimal library tells us to set
		// z to "the largest finite number that can be represented in the
		// current precision..." This is unreasonable, since this is an
		// _arbitrary_ precision library. Use signed Infinity instead.
		//
		// Because of the logic above, every rounding mode works out to the
		// following.
		if neg {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z.signal(Overflow|Inexact|Rounded, errOverflow)
	}

	z.scale = MinScale
	if neg {
		z.form = nzero
	} else {
		z.form = zero
	}
	return z.signal(Underflow|Inexact|Rounded|Subnormal, errUnderflow)
}

// These methods are here to prevent typos.

func (x *Big) isCompact() bool  { return x.compact != c.Inflated }
func (x *Big) isInflated() bool { return !x.isCompact() }

// Abs sets z to the absolute value of x and returns z.
func (z *Big) Abs(x *Big) *Big {
	if x.form == finite {
		if x.isCompact() {
			z.compact = arith.Abs(x.compact)
		} else {
			z.unscaled.Abs(&x.unscaled)
		}
		z.scale = x.scale
		z.form = finite
		return z
	}

	// |NaN|
	c, err := z.checkNaNs(x, x, "abs")
	if err != nil {
		return z.signal(c, err)
	}

	// |±Inf|
	x.form &= ^sign
	return z
}

// Add sets z to x + y and returns z.
func (z *Big) Add(x, y *Big) *Big {
	if x.form == finite && y.form == finite {
		z.form = finite
		if x.isCompact() && y.isCompact() {
			return z.addCompact(x, y)
		}
		return z.addBig(x, y)
	}

	// NaN + NaN
	// NaN + y
	// x + NaN
	if c, err := z.checkNaNs(x, y, "addition"); err != nil {
		return z.signal(c, err)
	}

	if x.form&y.form == inf && x.form^y.form == sign {
		// +Inf + -Inf
		// -Inf + +Inf
		z.form = qnan
		return z.signal(
			InvalidOperation,
			ErrNaN{"addition of infinities with opposing signs"},
		)
	}

	if x.form <= nzero && y.form <= nzero {
		// ±0 + ±0
		z.form = x.form & y.form
		return z
	}

	if x.form&inf != 0 || y.form <= nzero {
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
	// Fast path: if the scales are the same we can simply add without adjusting
	// either number.
	if x.scale == y.scale {
		z.scale = x.scale
		sum, ok := checked.Add(x.compact, y.compact)
		if ok {
			z.compact = sum
			if sum == 0 {
				z.form = zero
			}
		} else {
			xt := getInt64(x.compact)
			yt := getInt64(y.compact)
			z.unscaled.Add(xt, yt)
			putInt(xt)
			putInt(yt)

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

	scaled := checked.MulBigPow10(getInt64(lo.compact), inc)

	unscaled := getInt64(hi.compact)
	z.unscaled.Add(scaled, unscaled)

	putInt(scaled)
	putInt(unscaled)

	z.compact = c.Inflated
	if z.unscaled.Sign() == 0 {
		z.form = zero
	}
	return z
}

func (z *Big) addBig(x, y *Big) *Big {
	xb, yb := &x.unscaled, &y.unscaled
	if x.isCompact() {
		xb = big.NewInt(x.compact)
	} else if y.isCompact() {
		yb = big.NewInt(y.compact)
	}

	z.compact = c.Inflated
	if x.scale == y.scale {
		z.scale = x.scale
		if z.unscaled.Add(xb, yb).Sign() == 0 {
			z.form = zero
		}
		return z
	}

	his, los := x.scale, y.scale
	hi, lo := xb, yb
	if his < los {
		hi, lo = lo, hi
		his, los = los, his
	}
	// Inflate lo so we can add with matching scales.
	lo = checked.MulBigPow10(getInt(lo), his-los)
	if z.unscaled.Add(hi, lo).Sign() == 0 {
		z.form = zero
	}
	putInt(lo)
	z.scale = his
	return z
}

// BitLen returns the absolute value of x in bits. The result is undefined if
// x is an infinity or not a number value.
func (x *Big) BitLen() int {
	if x.form != finite {
		return 0
	}

	// If using an artificially inflated number determine the
	// bitlen using the number of digits.
	//
	// http://www.exploringbinary.com/number-of-bits-in-a-decimal-integer/
	if x.scale < 0 {
		// log2(10)
		const ln210 = 3.321928094887362347870319429489390175864831393024580612054

		// Number of zeros in scale + digits in z.
		d := -int(x.scale) + x.Precision()
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
// It does not modify z or x. The result is undefined if either z or x are not
// a number values.
func (z *Big) Cmp(x *Big) int {
	if z == x {
		return 0
	}

	// NaN cmp x
	// z cmp NaN
	// NaN cmp NaN
	if c, err := z.checkNaNs(z, x, "comparison"); err != nil {
		z.signal(c, err)
		return 0
	}

	// Fast path: different signs. Catches non-finite forms like zero and ±Inf.
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
				zu = getInt64(z.compact)
				defer putInt(zu)
			} else {
				xu = getInt64(x.compact)
				defer putInt(xu)
			}
			return zu.Cmp(xu)
		}
	}

	// Signs are the same and the scales differ. Compare the lengths of their
	// integral parts; if they differ in length one number is larger.
	// E.g., 1234.01
	//        123.011
	zl := int64(z.Precision() - int(z.scale))
	xl := int64(x.Precision() - int(x.scale))

	if zl < xl {
		return -zs
	}
	if zl > xl {
		return zs
	}

	// We have to inflate one of the numbers. Designate z as hi and x as lo.
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
	if debug && !ok {
		// TODO(eric): I'm like 99% positive this can't be reached.
		panic("should not be reached")
	}

	// Inflate lo.
	if xc != c.Inflated {
		if nx, ok := checked.MulPow10(xc, diff); !ok {
			// Can't fit in an int64, use big.Int.
			xm = checked.MulBigPow10(getInt64(xc), diff)
			defer putInt(xm)
			xc = c.Inflated
		} else {
			xc = nx
		}
	} else {
		xm = checked.MulBigPow10(getInt(xm), diff)
		defer putInt(xm)
	}

	// Swap back to original.
	if swap {
		zc, xc = xc, zc
		zm, xm = xm, zm
	}

	if zc != c.Inflated {
		if xc != c.Inflated {
			if zc > xc {
				return +1
			}
			if zc < xc {
				return -1
			}
			return 0
		}
		tmp := getInt64(zc)
		cmp := tmp.Cmp(xm)
		putInt(tmp)
		return cmp
	}
	if xc != c.Inflated {
		tmp := getInt64(xc)
		cmp := zm.Cmp(tmp)
		putInt(tmp)
		return cmp
	}
	return zm.Cmp(xm)
}

// Copy sets z to a copy of x and returns z.
func (z *Big) Copy(x *Big) *Big {
	if z != x {
		z.compact = x.compact
		z.Context = x.Context
		z.form = x.form
		z.scale = x.scale

		// Copy over unscaled if need be.
		if x.isInflated() {
			z.unscaled.Set(&x.unscaled)
		}
	}
	return z
}

// Float64 returns x as a float64.
func (x *Big) Float64() float64 {
	if x.form != finite {
		switch x.form {
		case pinf, ninf:
			return math.Inf(int(x.form & sign))
		case snan, qnan:
			return math.NaN()
		case nzero:
			return math.Copysign(0, -1)
		default: // zero
			return 0
		}
	}
	if x.isCompact() {
		if x.scale == 0 {
			return float64(x.compact)
		}
		const maxMantissa = 1 << 52
		if arith.Abs(x.compact) < maxMantissa {
			const maxPow10 = 22
			if x.scale > 0 && x.scale < maxPow10 {
				return float64(x.compact) / math.Pow10(int(x.scale))
			}
			if x.scale < 0 && x.scale < -maxPow10 {
				return float64(x.compact) * math.Pow10(int(-x.scale))
			}
		}
	}
	// No other way of doing it.
	f, _ := strconv.ParseFloat(x.String(), 64)
	return f
}

// Format implements the fmt.Formatter interface. The following verbs are
// supported:
//
// 	%s: -dddd.dd or -d.dddd±edd, depending on x
// 	%d: same as %s
// 	%v: same as %s
// 	%e: -d.dddd±edd
// 	%E: -d.dddd±Edd
// 	%f: -dddd.dd
// 	%g: same as %f
//
// Precision and width are honored in the same manner as the fmt package. In
// short, width is the minimum width of the formatted number. Given %f,
// precision is the number of digits following the radix. Given %g, precision
// is the number of significant digits.
//
// Format honors all flags (such as '+' and ' ') in the same manner as the fmt
// package, except for '#'. Unless used in conjunction with %v, %q, or %p, the
// '#' flag will be ignored; decimals have no defined hexadeximal or octal
// representation.
//
// %+v, %#v, %T, %#p, and %p all honor the formats specified in the fmt
// package's documentation.
func (x *Big) Format(s fmt.State, c rune) {
	prec, ok := s.Precision()
	if !ok {
		prec = noPrec
	}
	width, ok := s.Width()
	if !ok {
		width = noWidth
	}

	var (
		hash   = s.Flag('#')
		dash   = s.Flag('-')
		lpZero = s.Flag('0')
		plus   = s.Flag('+')
		space  = s.Flag(' ')
		f      = formatter{prec: prec, width: width}
	)

	// If we need to left pad then we need to first write our string into an
	// empty buffer.
	if lpZero {
		f.w = new(bytes.Buffer)
	} else {
		f.w = stateWrapper{s}
	}

	if plus {
		f.sign = '+'
	} else if space {
		f.sign = ' '
	}

	// noE is a placeholder for formats that do not use scientific notation
	// and don't require 'e' or 'E'
	const noE = 0
	switch c {
	case 's', 'd':
		f.format(x, normal, 'e')
	case 'q':
		// The fmt package's docs specify that the '+' flag
		// "guarantee[s] ASCII-only output for %q (%+q)"
		f.sign = 0

		// Since no other escaping is needed we can do it ourselves and save
		// whatever overhead running it through fmt.Fprintf would cause.
		quote := byte('"')
		if hash {
			quote = '`'
		}
		f.WriteByte(quote)
		f.format(x, normal, 'e')
		f.WriteByte(quote)
	case 'e', 'E':
		f.format(x, sci, byte(c))
	case 'f':
		if f.prec == noPrec {
			f.prec = 0
		}
		// %f's precision means "number of digits after the radix"
		if x.scale > 0 {
			if trail := x.Precision() - int(x.scale); trail < f.prec {
				f.prec += int(x.scale)
			} else {
				f.prec = int(x.scale) + trail
			}
		} else {
			f.prec += x.Precision()
		}
		f.format(x, plain, noE)
	case 'g':
		// %g's precision means "number of significant digits"
		f.format(x, plain, noE)

	// Make sure we return from the following two cases.
	case 'v':
		// %v == %s
		// TODO(eric): make this neater.
		if !hash && !plus {
			f.format(x, normal, 'e')
			break
		}

		// This is the easiest way of doing it. Note we can't use type Big Big,
		// even though it's declared inside a function. Go thinks it's
		// recursive. At least the fields are checked at compile time.
		type Big struct {
			Context  Context
			unscaled big.Int
			compact  int64
			scale    int32
			form     form
		}
		specs := ""
		if dash {
			specs += "-"
		} else if lpZero {
			specs += "0"
		}
		if hash {
			specs += "#"
		} else if plus {
			specs += "+"
		} else if space {
			specs += " "
		}
		fmt.Fprintf(s, "%"+specs+"v", (*Big)(x))
		return
	default:
		fmt.Fprintf(s, "%%!%c(*decimal.Big=%s)", c, x.String())
		return
	}

	needPad := f.n < int64(width)
	if needPad && lpZero {
		io.CopyN(s, zeroReader{}, int64(width)-f.n)
		needPad = false
	}

	// TODO(eric): find a better way of doing this.
	// If we had to write into a temp buffer, copy it over to the State.
	if r, ok := f.w.(*bytes.Buffer); ok {
		io.Copy(s, r)
	}

	// Right pad if need be.
	if needPad && dash {
		io.CopyN(s, spaceReader{}, int64(width)-f.n)
	}
}

// IsBig returns true if x, with its fractional part truncated, cannot fit
// inside an int64. If x is an infinity or a not a number value the result is
// undefined.
func (x *Big) IsBig() bool {
	// x.form != finite == zero, infinity, or nan
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
// x is an infinity or a not a number value the result is undefined.
func (x *Big) Int() *big.Int {
	if x.form != finite {
		return new(big.Int)
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
// result is undefined if x is an infinity, a not a number value, or if x does
// not fit inside an int64.
func (x *Big) Int64() int64 {
	if x.form != finite {
		return 0
	}

	// x might be too large to fit into an int64 *now*, but rescaling x might
	// shrink it enough. See issue #20.
	if !x.isCompact() {
		return x.Int().Int64()
	}

	b := x.compact
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
func (x *Big) IsFinite() bool { return x.form == finite }

// IsInf returns true if x is an infinity according to sign.
// If sign >  0, IsInf reports whether x is positive infinity.
// If sign <  0, IsInf reports whether x is negative infinity.
// If sign == 0, IsInf reports whether x is either infinity.
func (x *Big) IsInf(sign int) bool {
	return sign >= 0 && x.form == pinf || sign <= 0 && x.form == ninf
}

// IsNaN returns true if x is NaN.
// If sign >  0, IsNaN reports whether x is signaling NaN.
// If sign <  0, IsNaN reports whether x is quiet NaN.
// If sign == 0, IsNaN reports whether x is either NaN.
func (x *Big) IsNaN(signal int) bool {
	return signal >= 0 && x.form == snan || signal <= 0 && x.form == qnan
}

// IsInt reports whether x is an integer. Inf and NaN values are not integers.
func (x *Big) IsInt() bool {
	if x.form != finite {
		return x.form <= nzero
	}
	// The x.Cmp(one) check is necessary because x might be a decimal *and*
	// Prec <= 0 if x < 1.
	//
	// E.g., 0.1:  scale == 1, prec == 1
	//       0.01: scale == 2, prec == 1
	//
	// TODO(eric): avoid Cmp.
	return x.scale <= 0 || (x.Precision() <= int(x.scale) && x.Cmp(one) > 0)
}

// MarshalText implements encoding/TextMarshaler.
func (x *Big) MarshalText() ([]byte, error) {
	var (
		b bytes.Buffer
		f = formatter{w: &b, prec: noPrec, width: noWidth}
	)
	f.format(x, normal, 'e')
	return b.Bytes(), nil
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

	// NaN * NaN
	// NaN * y
	// x * NaN
	c, err := z.checkNaNs(x, y, "multiplication")
	if err != nil {
		return z.signal(c, err)
	}

	if x.form <= nzero && y.form&inf != 0 || x.form&inf != 0 && y.form <= nzero {
		// 0 * ±Inf
		// ±Inf * 0
		z.form = qnan
		return z.signal(
			InvalidOperation,
			ErrNaN{"multiplication of zero with infinity"},
		)
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
		// x + -y ∈ [-1<<31, 1<<31-1]
		return z.xflow(x.scale > 0, true)
	}

	prod, ok := checked.Mul(x.compact, y.compact)
	if ok {
		z.compact = prod
	} else {
		xt := getInt64(x.compact)
		yt := getInt64(y.compact)
		z.unscaled.Mul(xt, yt)
		putInt(xt)
		putInt(yt)
		z.compact = c.Inflated
	}
	z.scale = scale
	z.form = finite
	return z
}

func (z *Big) mulMixed(comp, non *Big) *Big {
	if debug && comp.isInflated() {
		panic("decimal.Mul (bug) comp.isInflated() == true")
	}
	if comp.scale == non.scale {
		scale, ok := checked.Add32(comp.scale, non.scale)
		if !ok {
			// x + -y ∈ [-1<<31, 1<<31-1]
			return z.xflow(comp.scale > 0, true)
		}
		tmp := getInt64(comp.compact)
		z.unscaled.Mul(tmp, &non.unscaled)
		putInt(tmp)

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
		// x + -y ∈ [-1<<31, 1<<31-1]
		return z.xflow(x.scale > 0, true)
	}
	z.unscaled.Mul(&x.unscaled, &y.unscaled)
	z.compact = c.Inflated
	z.scale = scale
	z.form = finite
	return z
}

// Neg sets z to -x and returns z. If x is positive infinity, z will be set to
// negative infinity and visa versa. If x == 0, z will be set to zero as well.
// NaN has no negative representation, and will result in an error.
func (z *Big) Neg(x *Big) *Big {
	if x.form == finite {
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

	// - NaN
	if c, err := z.checkNaNs(x, x, "negation"); err != nil {
		return z.signal(c, err)
	}

	// - ±Inf
	// - ±0
	z.form = x.form ^ sign
	return z
}

// New creates a new Big decimal with the given value and scale. For example:
//
//  New(1234, 3) // 1.234
//  New(42, 0)   // 42
//  New(4321, 5) // 0.04321
//  New(-1, 0)   // -1
//  New(3, -10)  // 30 000 000 000
//
func New(value int64, scale int32) *Big {
	return new(Big).SetMantScale(value, scale)
}

// Precision returns the precision of x. That is, it returns the number of
// digits in the unscaled form of x. x == 0 has a precision of 1. The result is
// undefined if x is an infinity or a not a number value.
func (x *Big) Precision() int {
	if x.form != finite {
		if x.form <= nzero {
			return 1
		}
		return 0
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
				Context:  x.Context,
				form:     x.form,
				scale:    x.scale,
			}, y)
		}
		if y.isCompact() {
			return z.quoBig(x, &Big{
				compact:  c.Inflated,
				unscaled: *big.NewInt(y.compact),
				Context:  y.Context,
				form:     y.form,
				scale:    y.scale,
			})
		}
		return z.quoBig(x, y)
	}

	// NaN / NaN
	// NaN / y
	// x / NaN
	c, err := z.checkNaNs(x, y, "division")
	if err != nil {
		return z.signal(c, err)
	}

	if x.form <= nzero && y.form <= nzero || (x.form&inf != 0 && y.form&inf != 0) {
		// 0 / 0
		// ±Inf / ±Inf
		z.form = qnan
		return z.signal(
			InvalidOperation,
			ErrNaN{"division of zero by zero or infinity by infinity"},
		)
	}

	if x.form <= nzero || y.form&inf != 0 {
		// 0 / y
		// x / ±Inf
		z.form = zero
		return z
	}

	// The spec requires the resulting infinity's sign to match
	// the "exclusive or of the signs of the operands."
	// http://speleotrove.com/decimal/daops.html#refdivide
	if xs, ys := x.Signbit(), y.Signbit(); (xs != ys) && (xs || ys) {
		z.form = ninf
	} else {
		z.form = pinf
	}

	if x.form&inf != 0 {
		// ±Inf / y
		return z
	}
	// x / 0
	return z.signal(DivisionByZero, errors.New("division by zero"))
}

func (z *Big) quoAndRound(x, y int64) *Big {
	// Quotient
	z.compact = x / y

	// ToZero means we can ignore remainder.
	if z.Context.RoundingMode == ToZero {
		return z
	}

	// Remainder
	r := x % y

	sign := int64(1)
	if (x < 0) != (y < 0) {
		sign = -1
	}
	if r != 0 {
		if z.needsInc(y, r, sign > 0, z.compact&1 != 0) {
			z.compact += sign
		}
		return z
	}
	if z.scale != z.Context.Precision() {
		return z.simplify()
	}
	return z
}

func (z *Big) quoCompact(x, y *Big) *Big {
	sdiff, ok := checked.Sub32(x.scale, y.scale)
	if !ok {
		// -x - y ∈ [-1<<31, 1<<31-1]
		return z.xflow(y.scale > 0, true)
	}

	zp := z.Context.Precision()
	xp := int32(x.Precision())
	yp := int32(y.Precision())

	// Multiply y by 10 if x' > y'
	if cmpNorm(x.compact, xp, y.compact, yp) {
		yp--
	}

	scale, ok := checked.Int32(int64(sdiff) + int64(yp) - int64(xp) + int64(zp))
	if !ok {
		// The wraparound from int32(int64(x)) where x ∉ [-1<<31, 1<<31-1]
		// will swap its sign.
		return z.xflow(scale < 0, false)
	}
	z.scale = scale

	shift, ok := checked.SumSub(zp, yp, xp)
	if !ok {
		return z.xflow(scale < 0, false)
	}

	xs, ys := x.compact, y.compact
	if shift > 0 {
		xs, ok = checked.MulPow10(x.compact, shift)
		if !ok {
			x0 := checked.MulBigPow10(getInt64(x.compact), shift)
			y0 := getInt64(y.compact)

			z = z.quoBigAndRound(x0, y0)

			putInt(x0)
			putInt(y0)
			return z
		}
		return z.quoAndRound(xs, ys)
	}

	// shift < 0
	ns, ok := checked.Sub32(xp, zp)
	if !ok {
		// -x - y ∈ [-1<<31, 1<<31-1]
		return z.xflow(zp > 0, true)
	}

	// No inflation needed.
	if ns == yp {
		return z.quoAndRound(xs, ys)
	}

	shift, ok = checked.Sub32(ns, yp)
	if !ok {
		// -x - y ∈ [-1<<31, 1<<31-1]
		return z.xflow(yp > 0, true)
	}
	ys, ok = checked.MulPow10(ys, shift)
	if !ok {
		x0 := getInt64(x.compact)
		y0 := checked.MulBigPow10(getInt64(y.compact), shift)

		z = z.quoBigAndRound(x0, y0)

		putInt(y0)
		putInt(x0)

		return z
	}
	return z.quoAndRound(xs, ys)
}

func (z *Big) quoBig(x, y *Big) *Big {
	scale, ok := checked.Sub32(x.scale, y.scale)
	if !ok {
		// -x - y ∈ [-1<<31, 1<<31-1]
		return z.xflow(y.scale > 0, true)
	}

	zp := z.Context.Precision()
	xp := int32(x.Precision())
	yp := int32(y.Precision())

	// Multiply y by 10 if x' > y'
	if cmpNormBig(&x.unscaled, xp, &y.unscaled, yp) {
		yp--
	}

	scale, ok = checked.Int32(int64(scale) + int64(yp) - int64(xp) + int64(zp))
	if !ok {
		// The wraparound from int32(int64(x)) where x ∉ [-1<<31, 1<<31-1] will
		// swap its sign.
		return z.xflow(scale < 0, true)
	}
	z.scale = scale

	shift, ok := checked.SumSub(zp, yp, xp)
	if !ok {
		// TODO(eric): See above comment about wraparound.
		return z.xflow(shift < 0, true)
	}
	if shift > 0 {
		xs := checked.MulBigPow10(getInt(&x.unscaled), shift)
		defer putInt(xs)
		return z.quoBigAndRound(xs, &y.unscaled)
	}

	// shift < 0
	ns, ok := checked.Sub32(xp, zp)
	if !ok {
		// -x - y ∈ [-1<<31, ..., 1<<31-1]
		return z.xflow(zp > 0, true)
	}
	shift, ok = checked.Sub32(ns, yp)
	if !ok {
		// -x - y ∈ {-1<<31, ..., 1<<31-1}
		return z.xflow(zp > 0, true)
	}
	ys := checked.MulBigPow10(getInt(&y.unscaled), shift)
	defer putInt(ys)
	return z.quoBigAndRound(&x.unscaled, ys)
}

func (z *Big) quoBigAndRound(x, y *big.Int) *Big {
	z.compact = c.Inflated

	r := get()
	defer putInt(r)
	q, r := z.unscaled.QuoRem(x, y, r)

	if z.Context.RoundingMode == ToZero {
		return z
	}

	sign := int64(1)
	if (x.Sign() < 0) != (y.Sign() < 0) {
		sign = -1
	}
	tmp := get().And(q, oneInt)
	odd := tmp.Sign() != 0
	defer putInt(tmp)

	if r.Sign() != 0 {
		if z.needsIncBig(y, r, sign > 0, odd) {
			tmp.SetInt64(int64(sign))
			z.unscaled.Add(&z.unscaled, tmp)
		}
		return z
	}
	if z.scale != z.Context.Precision() {
		return z.simplifyBig()
	}
	return z
}

func (z *Big) simplify() *Big {
	ok := false
	prec := z.Context.Precision()
	for arith.Abs(z.compact) >= 10 && z.scale > prec {
		if z.compact&1 != 0 || z.compact%10 != 0 {
			break
		}
		z.compact /= 10
		z.Context.Conditions |= Rounded
		if z.scale, ok = checked.Sub32(z.scale, 1); !ok {
			return z.xflow(false, z.compact < 0)
		}
	}
	return z
}

func (z *Big) simplifyBig() *Big {
	var (
		ok   = false
		prec = z.Context.Precision()
		tmp  = get()
	)
	for arith.BigAbs(&z.unscaled).Cmp(tenInt) >= 0 && z.scale > prec {
		if tmp.And(&z.unscaled, oneInt).Cmp(oneInt) != 0 ||
			tmp.Mod(&z.unscaled, tenInt).Sign() != 0 {
			break
		}
		z.unscaled.Div(&z.unscaled, tenInt)
		z.Context.Conditions |= Rounded
		if z.scale, ok = checked.Sub32(z.scale, 1); !ok {
			putInt(tmp)
			return z.xflow(false, z.Sign() < 0)
		}
	}
	putInt(tmp)
	return z
}

// Rat returns x as a *big.Rat. The result is undefined if x is an infinity or
// not a number value.
func (x *Big) Rat() *big.Rat {
	if x.form != finite {
		return new(big.Rat)
	}

	x0 := new(Big).Copy(x)
	if x0.scale > 0 {
		x0.scale = 0
	}
	num := x0.Int()

	var denom *big.Int
	if x.scale > 0 {
		if shift, ok := pow.Ten64(int64(x.scale)); ok {
			denom = big.NewInt(shift)
		} else {
			denom = new(big.Int).Exp(tenInt, big.NewInt(int64(x.scale)), nil)
		}
	} else {
		denom = big.NewInt(1)
	}
	return new(big.Rat).SetFrac(num, denom)
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
// undefined if n < 0 or z is not finite. No rounding will occur if n == 0.
// The result of Round will always be within the interval [⌊z⌋, z].
func (z *Big) Round(n int32) *Big {
	if n <= 0 || z.form != finite {
		return z
	}

	zp := z.Precision()
	if int(n) >= zp {
		return z
	}

	shift, ok := checked.Sub(int64(zp), int64(n))
	if !ok {
		return z.xflow(zp < 0, z.Signbit())
	}
	if shift <= 0 {
		return z
	}

	z.Context.SetPrecision(n)
	z.Context.Conditions |= Rounded
	z.scale -= int32(shift)

	if z.isCompact() {
		if val, ok := pow.Ten64(shift); ok {
			return z.quoAndRound(z.compact, val)
		}
		z.unscaled.SetInt64(z.compact)
	}
	val := pow.BigTen(shift)
	return z.quoBigAndRound(&z.unscaled, &val)
}

// Scale returns x's scale.
func (x *Big) Scale() int32 { return x.scale }

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

		// TODO(eric): should we round even if z == x?
		z.Round(z.Context.Precision())
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

// SetFloat64 sets z to the provided float64.
//
// Because certain numbers cannot be exactly represented as floating-point
// numbers, SetFloat64 "rounds" its input in order to break it into simple
// mantissa and scale parts as if SetMantScale were called. If SetFloat64 took
// its input as-is, the result of calling SetFloat64(0.1) would be
// 0.1000000000000000055511151231257827021181583404541015625.
//
// Approximately 2.3% of decimals created from floats will have an rounding
// imprecision of ± 1 ULP.
func (z *Big) SetFloat64(value float64) *Big {
	if value == 0 {
		z.form = 0
		return z
	}

	if math.IsNaN(value) {
		z.form = qnan
		return z.signal(InvalidOperation, ErrNaN{"SetFloat64(NaN)"})
	}
	if math.IsInf(value, 0) {
		if math.IsInf(value, 1) {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z.signal(InvalidOperation, errors.New("SetFloat(Inf)"))
	}

	var scale int32

	// If value is not an integer (has a fractional part) bump up its value and
	// find the appropriate scale.
	if _, fr := math.Modf(value); fr != 0 {
		scale = findScale(value)
		value *= math.Pow10(int(scale))
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

// SetInf sets x to -Inf if signbit is set or +Inf is signbit is not set, and
// returns x.
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

// SetNaN sets z to a signaling NaN if signal is true or quiet NaN otherwise and
// returns z.
func (z *Big) SetNaN(signal bool) *Big {
	if signal {
		z.form = snan
	} else {
		z.form = qnan
	}
	return z
}

// SetRat sets z to to the possibly rounded value of x and return z.
func (z *Big) SetRat(x *big.Rat) *Big {
	// TODO(eric): once we modify quoBig to accept two *big.Ints then this
	// method can just be ``return z.quoBig(x.Num(), x.Denom())''
	a := new(Big).SetBigMantScale(x.Num(), 0)
	b := new(Big).SetBigMantScale(x.Denom(), 0)
	return z.Quo(a, b)
}

// SetScale sets z's scale to scale and returns z.
func (z *Big) SetScale(scale int32) *Big {
	z.scale = scale
	return z
}

// Regexp matches any valid string representing a decimal that can be pased to
// SetString.
var Regexp = regexp.MustCompile(`(?i)(((\+|-)?(\d+\.\d*|\.?\d+)([eE][+-]?\d+)?)|(inf(infinity)?))|((\+|-)?([sq]?nan))`)

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
// 	NaN
// 	qNaN
// 	sNaN
//
// Inf and NaN map to +Inf and qNaN, respectively. NaN values may have optional
// diagnostic information, represented as trailing digits; for example,
// ``NaN123''. These digits are otherwise ignored but are included for
// robustness.
func (z *Big) SetString(s string) (*Big, bool) {
	if s == "" {
		return z.signal(ConversionSyntax, errors.New(`SetString("")`)), false
	}

	// http://speleotrove.com/decimal/daconvs.html#refnumsyn
	//
	//   sign           ::=  '+' | '-'
	//   digit          ::=  '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' |
	//                       '8' | '9'
	//   indicator      ::=  'e' | 'E'
	//   digits         ::=  digit [digit]...
	//   decimal-part   ::=  digits '.' [digits] | ['.'] digits
	//   exponent-part  ::=  indicator [sign] digits
	//   infinity       ::=  'Infinity' | 'Inf'
	//   nan            ::=  'NaN' [digits] | 'sNaN' [digits]
	//   numeric-value  ::=  decimal-part [exponent-part] | infinity
	//   numeric-string ::=  [sign] numeric-value | [sign] nan
	//
	// We deviate a little by being a tad bit more forgiving. For instance,
	// we allow case-insensitive nan and infinity values.

	// Eliminate overhead when searching for inf and nan values.
	if s[len(s)-1] > '9' {
		switch parse.ParseSpecial(s) {
		case parse.QNaN:
			z.form = qnan
		case parse.SNaN:
			z.form = snan
		case parse.PInf:
			z.form = pinf
		case parse.NInf:
			z.form = ninf
		default:
			return z.signal(
				ConversionSyntax,
				fmt.Errorf("SetString: invalid input: %q", s),
			), false
		}
		return z, true
	}

	var scale int32

	// Check for a scientific string.
	if i := strings.LastIndexAny(s, "Ee"); i > 0 {
		eint, err := strconv.ParseInt(s[i+1:], 10, 32)
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrSyntax {
				z.form = qnan
				return z.signal(ConversionSyntax, err), false
			}
			// strconv.ErrRange.
			return z.xflow(eint < 0, s[0] == '-'), false
		}
		s = s[:i]
		scale = -int32(eint)
	}

	switch strings.Count(s, ".") {
	case 0:
		// OK
	case 1:
		i := strings.IndexByte(s, '.')
		s = s[:i] + s[i+1:]
		sc, ok := checked.Add32(scale, int32(len(s)-i))
		if !ok {
			// It's impossible for the scale to underflow here since the rhs will
			// always be [0, len(s)]
			return z.xflow(true, s[0] == '-'), false
		}
		scale = sc
	default:
		return z.signal(
			ConversionSyntax,
			errors.New("SetString: too many '.' in input"),
		), false
	}

	var err error
	z.form = finite
	// Numbers == 19 can be out of range, but try the edge case anyway.
	if len(s) <= 19 {
		if z.compact, err = strconv.ParseInt(s, 10, 64); err != nil {
			nerr, ok := err.(*strconv.NumError)
			if !ok || nerr.Err == strconv.ErrSyntax {
				z.form = qnan
				return z.signal(ConversionSyntax, err), false
			}
			err = nerr.Err
		} else if z.compact == 0 {
			if s[0] == '-' {
				z.form = nzero
			} else {
				z.form = zero
			}
		}
		if z.compact == c.Inflated {
			z.unscaled.SetInt64(z.compact)
		}
	}
	if (err == strconv.ErrRange && len(s) == 19) || len(s) > 19 {
		if _, ok := z.unscaled.SetString(s, 10); !ok {
			return z.signal(
				ConversionSyntax,
				// TODO(eric): a better error message?
				errors.New("SetString: invalid syntax"),
			), false
		}
		z.compact = c.Inflated
		if z.unscaled.Sign() == 0 {
			if s[0] == '-' {
				z.form = nzero
			} else {
				z.form = zero
			}
		}
	}
	z.scale = scale
	return z, true
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
//
// The result is undefined if x is a not a number value.
func (x *Big) Sign() int {
	if x.form != finite {
		switch x.form {
		case zero, nzero:
			return 0
		case ninf:
			return -1
		case pinf:
			return +1
		default:
			return 0
		}
	}

	// x is finite.
	if x.isCompact() {
		// TODO(eric): remove this conditional when we drop support for Go 1.7.
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

func (x *Big) signal(c Condition, err error) *Big {
	switch ctx := &x.Context; ctx.OperatingMode {
	case Go:
		// Go mode always panics on NaNs.
		if _, ok := err.(ErrNaN); ok {
			panic(err)
		}
	case GDA:
		ctx.Conditions = c
		if c&ctx.Traps != 0 {
			ctx.Err = err
		}
	default:
		ctx.Conditions = c | InvalidContext
		ctx.Err = fmt.Errorf("invalid OperatingMode: %d", ctx.OperatingMode)
		x.form = qnan
	}
	return x
}

// Signbit returns true if x is negative, negative infinity, or negative zero.
func (x *Big) Signbit() bool {
	if x.form != finite {
		return x.form == ninf || x.form == nzero
	}
	if x.isCompact() {
		return x.compact < 0
	}
	return x.unscaled.Sign() < 0
}

// String returns the string representation of x. It's equivalent to the %s verb
// discussed in the Format method's documentation. Special cases depend on the
// OperatingMode. The defaults (for OperatingMode == Go) are:
//
//  "<nil>" if x == nil
//  "+Inf"  if x.IsInf(+1)
//  "+Inf"  if x.IsInf(0)
//  "-Inf"  if x.IsInf(-1)
//
func (x *Big) String() string {
	// TODO(eric): use a pool?
	var (
		b bytes.Buffer
		f = formatter{w: &b, prec: noPrec, width: noWidth}
	)
	f.format(x, normal, 'e')
	return b.String()
}

// Sub sets z to x - y and returns z.
func (z *Big) Sub(x, y *Big) *Big {
	if x.form == finite && y.form == finite {
		// TODO(eric): Write this without using Neg to save an allocation.
		return z.Add(x, new(Big).Neg(y))
	}

	// NaN - NaN
	// NaN - y
	// x - NaN
	c, err := z.checkNaNs(x, y, "subtraction")
	if err != nil {
		return z.signal(c, err)
	}

	if x.form&inf != 0 && x.form == y.form {
		// +Inf - +Inf
		// -Inf - -Inf
		z.form = qnan
		return z.signal(
			InvalidOperation,
			ErrNaN{"subtraction of infinities with equal signs"},
		)
	}

	if x.form <= nzero && y.form <= nzero {
		// ±0 - ±0
		z.form = zero
		return z
	}

	if x.form&inf != 0 || y.form <= nzero {
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
	// TODO(eric): get rid of the allocation here.
	if _, ok := z.SetString(string(data)); !ok {
		return errors.New("Big.UnmarshalText: invalid decimal format")
	}
	return nil
}
