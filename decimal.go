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
	"encoding"
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
	"github.com/ericlagergren/decimal/internal/compat"
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
	const debug = false
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

// TODO(eric): Perhaps use math/big.ErrNaN if possible in the future?

// An ErrNaN panic is raised by a decimal operation that would lead to a NaN
// under IEEE-754 rules. An ErrNaN implements the error interface.
type ErrNaN struct {
	Msg string
}

func (e ErrNaN) Error() string {
	return e.Msg
}

var _ error = ErrNaN{}

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
		// _arbitrary_ precision data type. Use signed Infinity instead.
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
	if debug {
		x.validate()
	}

	if x.form == finite {
		if x.isCompact() {
			z.compact = arith.Abs(x.compact)
		} else {
			z.compact = c.Inflated
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
	// |±0|
	z.form &= ^sign
	return z
}

// Add sets z to x + y and returns z.
func (z *Big) Add(x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}

	if x.form == finite && y.form == finite {
		z.form = finite
		if x.isCompact() && y.isCompact() {
			return z.addCompact(x, y).round()
		}
		return z.addBig(x, y).round()
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
	if debug {
		if x.compact == 0 || y.compact == 0 {
			panic("addCompact: operand == 0")
		}
	}

	xc, yc := x.compact, y.compact
	ok := false
	switch {
	case x.scale == y.scale:
		z.scale = x.scale
	case x.scale < y.scale:
		if xc, ok = checked.MulPow10(xc, uint64(y.scale-x.scale)); !ok {
			return z.addBig(x, y)
		}
		z.scale = y.scale
	case x.scale > y.scale:
		if yc, ok = checked.MulPow10(yc, uint64(x.scale-y.scale)); !ok {
			return z.addBig(x, y)
		}
		z.scale = x.scale
	}
	if z.compact, ok = checked.Add(xc, yc); ok {
		if z.compact == 0 {
			z.form = zero
		}
		return z
	}
	if arith.Add128(&z.unscaled, xc, yc).Sign() == 0 {
		z.form = zero
	}
	z.compact = c.Inflated
	return z
}

func (z *Big) addBig(x, y *Big) *Big {
	// TODO(eric): if debug { }

	xb, yb := &x.unscaled, &y.unscaled
	if x.isCompact() {
		xb = big.NewInt(x.compact)
	}
	if y.isCompact() {
		yb = big.NewInt(y.compact)
	}

	switch {
	case x.scale == y.scale:
		z.scale = x.scale
	case x.scale < y.scale:
		xb = checked.MulBigPow10(xb, uint64(y.scale-x.scale))
		z.scale = y.scale
	case x.scale > y.scale:
		yb = checked.MulBigPow10(yb, uint64(x.scale-y.scale))
		z.scale = x.scale
	}
	if z.unscaled.Add(xb, yb).Sign() == 0 {
		z.form = zero
	}
	z.compact = c.Inflated
	return z
}

// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
//
// It does not modify x or y. The result is undefined if either x or y are not
// a number values.
func (x *Big) Cmp(y *Big) int {
	if debug {
		x.validate()
		y.validate()
	}

	if x == y {
		return 0
	}

	// NaN cmp x
	// z cmp NaN
	// NaN cmp NaN
	if c, err := x.checkNaNs(x, y, "comparison"); err != nil {
		x.signal(c, err)
		return 0
	}

	// Fast path: different signs. Catches non-finite forms like zero and ±Inf.
	xs := x.Sign()
	ys := y.Sign()
	switch {
	case xs > ys:
		return +1
	case xs < ys:
		return -1
	case xs == 0 && ys == 0:
		return 0
	}

	// zs == xs

	// Same scales means we can compare straight across.
	if x.scale == y.scale {
		switch {
		case x.isCompact() && y.isCompact():
			if x.compact > y.compact {
				return +1
			}
			if x.compact < y.compact {
				return -1
			}
			return 0
		case x.isInflated() && y.isInflated():
			return x.unscaled.Cmp(&y.unscaled)
		default:
			// The inflated number is more than likely larger, but I'm not 100%
			// certain that inflated > compact is an invariant.
			xu, yu := &x.unscaled, &y.unscaled
			if x.isCompact() {
				xu = big.NewInt(x.compact)
			} else {
				yu = big.NewInt(y.compact)
			}
			return xu.Cmp(yu)
		}
	}

	// Signs are the same and the scales differ. Compare the lengths of their
	// integral parts; if they differ in length one number is larger.
	// E.g., 1234.01
	//        123.011
	xl := x.Precision() - int(x.scale)
	yl := y.Precision() - int(y.scale)

	if xl != yl {
		if xl < yl {
			return -xs
		}
		return xs
	}

	diff := int64(x.scale) - int64(y.scale)
	shift := uint64(arith.Abs(diff))
	if pow.Safe(shift) && x.isCompact() && y.isCompact() {
		p, _ := pow.Ten(shift)
		if diff < 0 {
			return -(xs * arith.AbsCmp128(y.compact, x.compact, p))
		}
		return xs * arith.AbsCmp128(x.compact, y.compact, p)
	}

	xw, yw := x.unscaled.Bits(), y.unscaled.Bits()
	if x.isCompact() {
		xw = arith.Words(x.compact)
	}
	if y.isCompact() {
		yw = arith.Words(y.compact)
	}

	var tmp big.Int
	if diff < 0 {
		xw = checked.MulBigPow10(tmp.SetBits(xw), shift).Bits()
	} else {
		yw = checked.MulBigPow10(tmp.SetBits(yw), shift).Bits()
	}
	return xs * arith.CmpBits(xw, yw)
}

// Copy sets z to a copy of x and returns z.
func (z *Big) Copy(x *Big) *Big {
	if debug {
		x.validate()
	}

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
	if debug {
		x.validate()
	}

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
	// TODO(eric): find a better way of doing this.
	f, _ := strconv.ParseFloat(x.String(), 64)
	return f
}

// Float sets z to x and returns z. z is allowed to be nil. The result is
// undefined if z is a NaN value.
func (x *Big) Float(z *big.Float) *big.Float {
	if debug {
		x.validate()
	}

	if z == nil {
		z = new(big.Float)
	}

	switch x.form {
	case finite:
		// TODO(eric): is there a more efficient way?
		z.SetRat(x.Rat(nil))
	case zero, snan, qnan:
		z.SetInt64(0)
	case nzero:
		z.SetInt64(0).Neg(z)
	case pinf, ninf:
		z.SetInf(x.form == pinf)
	}
	return z
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
	if debug {
		x.validate()
	}

	prec, ok := s.Precision()
	if !ok {
		prec = noPrec
	}
	width, ok := s.Width()
	if !ok {
		width = noWidth
	}

	var (
		hash    = s.Flag('#')
		dash    = s.Flag('-')
		lpZero  = s.Flag('0')
		lpSpace = width != noWidth && !dash && !lpZero
		plus    = s.Flag('+')
		space   = s.Flag(' ')
		f       = formatter{prec: prec, width: width}
		e       = x.Context.OperatingMode.get().e
	)

	// If we need to left pad then we need to first write our string into an
	// empty buffer.
	if lpZero || lpSpace {
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
		f.format(x, normal, e)
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
		f.format(x, normal, e)
		f.WriteByte(quote)
	case 'e', 'E':
		f.format(x, sci, byte(c))
	case 'f':
		if f.prec == noPrec {
			f.prec = 0
		}
		// %f's precision means "number of digits after the radix"
		if x.scale > 0 {
			f.prec += int(x.scale)
			if trail := int(x.Precision()) - int(x.scale); trail >= f.prec {
				f.prec += trail
			}
		} else {
			f.prec += int(x.Precision())
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
	if !needPad {
		return
	}

	if lpZero || lpSpace {
		if lpZero {
			io.CopyN(s, zeroReader{}, int64(width)-f.n)
		} else {
			io.CopyN(s, spaceReader{}, int64(width)-f.n)
		}
		// If we had to write into a temp buffer, copy it over to the State.
		io.Copy(s, f.w.(*bytes.Buffer))
	} else if dash {
		io.CopyN(s, spaceReader{}, int64(width)-f.n)
	}
}

var _ fmt.Formatter = (*Big)(nil)

// IsBig returns true if x, with its fractional part truncated, cannot fit
// inside an int64. If x is an infinity or a NaN value the result is undefined.
func (x *Big) IsBig() bool {
	if debug {
		x.validate()
	}

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

// Int sets z to x, truncating the fractional portion (if any) and returns z. z
// is allowed to be nil. If x is an infinity or a NaN value the result is
// undefined.
func (x *Big) Int(z *big.Int) *big.Int {
	if debug {
		x.validate()
	}

	if z == nil {
		z = new(big.Int)
	}

	if x.form != finite {
		return z
	}

	if x.isCompact() {
		z.SetInt64(x.compact)
	} else {
		z.Set(&x.unscaled)
	}
	if x.scale == 0 {
		return z
	}
	if x.scale < 0 {
		return checked.MulBigPow10(z, uint64(-x.scale))
	}
	return z.Quo(z, pow.BigTen(uint64(x.scale)))
}

// Int64 returns x as an int64, truncating the fractional portion, if any. The
// result is undefined if x is an infinity, a NaN value, or if x does not fit
// inside an int64.
func (x *Big) Int64() int64 {
	if debug {
		x.validate()
	}

	if x.form != finite {
		return 0
	}

	// x might be too large to fit into an int64 *now*, but rescaling x might
	// shrink it enough. See issue #20.
	if !x.isCompact() {
		return x.Int(nil).Int64()
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
// If sign >  0, IsNaN reports whether x is quiet NaN.
// If sign <  0, IsNaN reports whether x is signaling NaN.
// If sign == 0, IsNaN reports whether x is either NaN.
func (x *Big) IsNaN(quiet int) bool {
	return quiet >= 0 && x.form == qnan || quiet <= 0 && x.form == snan
}

// IsInt reports whether x is an integer. Infinity and NaN values are not
// integers.
func (x *Big) IsInt() bool {
	if debug {
		x.validate()
	}

	if x.form != finite {
		return x.form <= nzero
	}

	// 5000, 420
	if x.scale <= 0 {
		return true
	}

	xp := x.Precision()

	// 0.001
	// 0.5
	if int(x.scale) >= xp {
		return false
	}

	// 44.00
	// 1.000
	if x.isCompact() {
		for v := x.compact; v%10 == 0; v /= 10 {
			xp--
		}
	} else {
		v := new(big.Int).Set(&x.unscaled)
		r := new(big.Int)
		for {
			v.QuoRem(v, c.TenInt, r)
			if r.Sign() != 0 {
				break
			}
			xp--
		}
	}
	return xp <= int(x.scale)
}

// MarshalText implements encoding.TextMarshaler.
func (x *Big) MarshalText() ([]byte, error) {
	if debug {
		x.validate()
	}

	var (
		b bytes.Buffer
		f = formatter{w: &b, prec: noPrec, width: noWidth}
	)
	f.format(x, normal, 'e')
	return b.Bytes(), nil
}

// Mul sets z to x * y and returns z.
func (z *Big) Mul(x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}

	if x.form == finite && y.form == finite {
		z.form = finite
		if x.isCompact() && y.isCompact() {
			return z.mulCompact(x, y).round()
		}
		return z.mulBig(x, y).round()
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
	if debug {
		if x.compact == 0 || y.compact == 0 {
			panic("mulCompact: zero operand")
		}
	}

	scale, ok := checked.Add32(x.scale, y.scale)
	if !ok {
		// x + -y ∈ [-1<<31, 1<<31-1]
		return z.xflow(x.scale > 0, true)
	}
	z.scale = scale

	prod, ok := checked.Mul(x.compact, y.compact)
	if ok {
		z.compact = prod
	} else {
		arith.Mul128(&z.unscaled, x.compact, y.compact)
		z.compact = c.Inflated
	}
	z.form = finite
	return z
}

// mulBig sets z to x * y. Both x or y or both should be inflated.
func (z *Big) mulBig(x, y *Big) *Big {
	if debug {
		if x.isCompact() && y.isCompact() {
			panic("mulBig: both are compact")
		}
	}

	if x.isCompact() {
		arith.MulInt64(&z.unscaled, &y.unscaled, x.compact)
	} else if y.isCompact() {
		arith.MulInt64(&z.unscaled, &x.unscaled, y.compact)
	} else {
		z.unscaled.Mul(&x.unscaled, &y.unscaled)
	}

	z.compact = c.Inflated
	scale, ok := checked.Add32(x.scale, y.scale)
	if !ok {
		// x + -y ∈ [-1<<31, 1<<31-1]
		return z.xflow(x.scale > 0, true)
	}
	z.scale = scale
	z.form = finite
	return z
}

// Neg sets z to -x and returns z. If x is positive infinity, z will be set to
// negative infinity and visa versa. If x == 0, z will be set to zero as well.
// NaN has no negative representation, and will result in an error.
func (z *Big) Neg(x *Big) *Big {
	if debug {
		x.validate()
	}

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
// undefined if x is an infinity or a NaN value.
func (x *Big) Precision() int {
	if debug {
		x.validate()
	}

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
	if debug {
		x.validate()
		y.validate()
	}

	if x.form == finite && y.form == finite {
		// set z.form == finite inside the quo* methods.
		// x / y (common case)
		if x.isCompact() && y.isCompact() {
			return z.quoCompact(x, y)
		}
		return z.quo(x, y)
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

func (z *Big) quoCompact(x, y *Big) *Big {
	return z.quoCoreCompact(
		x.compact, x.scale, x.Precision(),
		y.compact, y.scale, y.Precision(),
	)
}

// quoCoreCompact implements division of two compact decimals.
func (z *Big) quoCoreCompact(
	x int64, xs int32, xp int,
	y int64, ys int32, yp int,
) *Big {
	sdiff, ok := checked.Sub32(xs, ys)
	if !ok {
		// -x - y ∈ [-1<<31, 1<<31-1]
		return z.xflow(ys > 0, true)
	}

	// Multiply y by 10 if x' > y'
	if cmpNorm(x, xp, y, yp) {
		yp--
	}

	zp := z.Context.Precision()
	scale, ok := checked.Int32(int64(sdiff) + int64(yp) - int64(xp) + int64(zp))
	if !ok {
		// The wraparound from int32(int64(x)) where x ∉ [-1<<31, 1<<31-1] will
		// swap its sign.
		return z.xflow(scale < 0, false)
	}
	z.scale = scale

	shift := zp + yp - xp
	if shift > 0 { // shift > 0
		if sx, ok := checked.MulPow10(x, uint64(shift)); ok {
			return z.quoAndRoundCompact(sx, y)
		}
		xb := checked.MulBigPow10(big.NewInt(x), uint64(shift))
		return z.quoAndRoundBig(xb, big.NewInt(y))
	}
	ns := xp - zp
	if ns == yp {
		return z.quoAndRoundCompact(x, y)
	}
	// shift < 0
	shift = ns - yp
	if sy, ok := checked.MulPow10(y, uint64(shift)); ok {
		return z.quoAndRoundCompact(x, sy)
	}
	yb := checked.MulBigPow10(big.NewInt(y), uint64(shift))
	return z.quoAndRoundBig(big.NewInt(x), yb)
}

func (z *Big) quoAndRoundCompact(x, y int64) *Big {
	z.form = finite

	// Quotient
	z.compact = x / y

	// ToZero means we can ignore remainder.
	if z.Context.RoundingMode == ToZero {
		z.Context.Conditions |= Rounded | Inexact
		return z
	}

	// Remainder
	r := x % y
	if r == 0 {
		return z.simplify()
	}

	rc := 1
	if r2, ok := checked.Mul(r, 2); ok {
		rc = arith.AbsCmp(r2, y)
	}

	pos := (x < 0) == (y < 0)
	if z.needsInc(rc, pos) {
		z.Context.Conditions |= Rounded | Inexact
		if pos {
			z.compact++
		} else {
			z.compact--
		}
	} else if z.compact == 0 {
		if pos {
			z.form = zero
		} else {
			z.form = nzero
		}
	}
	return z
}

func (z *Big) simplify() *Big {
	if int(z.scale) == z.Context.Precision() {
		return z
	}
	ok := false
	prec := z.Context.Precision()
	for arith.Abs(z.compact) >= 10 && int(z.scale) > prec {
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

func (z *Big) quo(x, y *Big) *Big {
	return z.quoCore(
		&x.unscaled, x.compact, x.scale, x.Precision(),
		&y.unscaled, y.compact, y.scale, y.Precision(),
	)
}

// see quoCompactCore. xc and yc override xb and yb, respectively, if they !=
// c.Inflated. If both xc and yc != c.Inflated quoCompactCore will be called.
// This method should be used sparingly.
func (z *Big) quoCore(
	xb *big.Int, xc int64, xs int32, xp int,
	yb *big.Int, yc int64, ys int32, yp int,
) *Big {
	sdiff, ok := checked.Sub32(xs, ys)
	if !ok {
		// -x - y ∈ [-1<<31, 1<<31-1]
		return z.xflow(ys > 0, true)
	}

	// TODO(eric): re-work this quo* methods. I don't like how they're laid out.

	if xc != c.Inflated {
		if yc != c.Inflated {
			return z.quoCoreCompact(xc, xs, xp, yc, ys, yp)
		}
		xb = big.NewInt(xc)
	}
	if yc != c.Inflated {
		yb = big.NewInt(yc)
	}

	// Multiply y by 10 if x' > y'
	if cmpNormBig(xb, xp, yb, yp) {
		yp--
	}

	zp := z.Context.Precision()
	scale, ok := checked.Int32(int64(sdiff) + int64(yp) - int64(xp) + int64(zp))
	if !ok {
		// The wraparound from int32(int64(x)) where x ∉ [-1<<31, 1<<31-1] will
		// swap its sign.
		return z.xflow(scale < 0, true)
	}
	z.scale = scale

	shift := zp + yp - xp
	if shift > 0 {
		xb = checked.MulBigPow10(new(big.Int).Set(xb), uint64(shift))
		return z.quoAndRoundBig(xb, yb)
	}

	ns := xp - zp
	if ns == yp {
		return z.quoAndRoundBig(xb, yb)
	}

	shift = ns - yp
	yb = checked.MulBigPow10(new(big.Int).Set(yb), uint64(shift))
	return z.quoAndRoundBig(xb, yb)
}

func (z *Big) quoAndRoundBig(x, y *big.Int) *Big {
	z.form = finite
	z.compact = c.Inflated

	if z.Context.RoundingMode == ToZero {
		z.Context.Conditions |= Rounded | Inexact
		z.unscaled.Quo(x, y)
		return z
	}

	// Can't use z.unscaled.Sign because it z might == 0
	pos := x.Sign() == y.Sign()
	_, r := z.unscaled.QuoRem(x, y, new(big.Int))
	if r.Sign() == 0 {
		return z.simplifyBig()
	}

	var rc int
	rv := r.Int64()
	// Drop into integers if we can.
	if compat.IsInt64(r) && compat.IsInt64(y) && (rv <= math.MaxInt64/2 && rv > -math.MaxInt64/2) {
		rc = arith.AbsCmp(rv*2, y.Int64())
	} else {
		rc = compat.BigCmpAbs(r.Mul(r, c.TwoInt), y)
	}

	if z.needsInc(rc, pos) {
		z.Context.Conditions |= Rounded | Inexact
		if pos {
			z.unscaled.Add(&z.unscaled, c.OneInt)
		} else {
			z.unscaled.Sub(&z.unscaled, c.OneInt)
		}
	} else if z.unscaled.Sign() == 0 {
		if pos {
			z.form = zero
		} else {
			z.form = nzero
		}
	}
	return z
}

func (z *Big) simplifyBig() *Big {
	if int(z.scale) == z.Context.Precision() {
		return z
	}
	if compat.IsInt64(&z.unscaled) {
		z.compact = z.unscaled.Int64()
		return z.simplify()
	}
	var (
		ok   = false
		prec = z.Context.Precision()
		tmp  = new(big.Int)
	)
	// arith.BigAbsAlias won't work here.
	for tmp.SetBits(z.unscaled.Bits()).Cmp(c.TenInt) >= 0 && int(z.scale) > prec {
		if z.unscaled.Bit(0) != 0 || tmp.Mod(&z.unscaled, c.TenInt).Sign() != 0 {
			break
		}
		z.unscaled.Div(&z.unscaled, c.TenInt)
		z.Context.Conditions |= Rounded
		if z.scale, ok = checked.Sub32(z.scale, 1); !ok {
			return z.xflow(false, z.Sign() < 0)
		}
	}
	return z
}

// Rat sets z to x returns z. z is allowed to be nil. The result is undefined if
// x is an infinity or NaN value.
func (x *Big) Rat(z *big.Rat) *big.Rat {
	if debug {
		x.validate()
	}

	if z == nil {
		z = new(big.Rat)
	}

	if x.form != finite {
		return z.SetInt64(0)
	}

	x0 := new(Big).Copy(x)
	if x0.scale > 0 {
		x0.scale = 0
	}
	num := x0.Int(nil)

	var denom *big.Int
	if x.scale > 0 {
		if shift, ok := pow.Ten(uint64(x.scale)); ok {
			denom = new(big.Int).SetUint64(shift)
		} else {
			tmp := new(big.Int).SetUint64(uint64(x.scale))
			denom = tmp.Exp(c.TenInt, tmp, nil)
		}
	} else {
		denom = big.NewInt(1)
	}
	return z.SetFrac(num, denom)
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

func (z *Big) round() *Big {
	zp := z.Context.Precision()
	if zp != 0 && z.Context.OperatingMode == GDA {
		return z.Round(zp)
	}
	return z
}

// Round rounds z down to n digits of precision and returns z. The result is
// undefined if z is not finite. No rounding will occur if n == 0. The result of
// Round will always be within the interval [⌊10**x⌋, z] where x = the precision
// of z.
func (z *Big) Round(n int) *Big {
	if debug {
		z.validate()
	}

	if n <= 0 || z.form != finite {
		return z
	}

	zp := z.Precision()
	if n >= zp {
		return z
	}

	shift := zp - n
	if shift > MaxScale {
		return z.xflow(false, true)
	}

	scale, ok := checked.Sub32(z.scale, int32(shift))
	if !ok {
		return z.xflow(false, true)
	}
	z.scale = scale

	z.Context.Conditions |= Rounded

	if z.isCompact() {
		if val, ok := pow.TenInt(uint64(shift)); ok {
			return z.quoAndRoundCompact(z.compact, val)
		}
		z.unscaled.SetInt64(z.compact)
	}
	return z.quoAndRoundBig(&z.unscaled, pow.BigTen(uint64(shift)))
}

// Quantize sets z to the number equal in value and sign to z with the scale, n.
func (z *Big) Quantize(n int32) *Big {
	if debug {
		z.validate()
	}

	if z.form != finite {
		if z.form <= nzero {
			z.scale = n
		} else {
			z.form = qnan
		}
		return z
	}

	if z.scale == n {
		return z
	}

	shift := n - z.scale
	if shift == 0 {
		return z
	}
	z.scale = n

	if z.isCompact() {
		if shift > 0 {
			if zc, ok := checked.MulPow10(z.compact, uint64(shift)); ok {
				z.compact = zc
				return z
			}
			// shift < 0
		} else if yc, ok := pow.TenInt(uint64(-shift)); ok {
			return z.quoAndRoundCompact(z.compact, yc)
		}
		z.unscaled.SetInt64(z.compact)
	}
	z.compact = c.Inflated
	if shift > 0 {
		checked.MulBigPow10(&z.unscaled, uint64(shift))
		return z
	}
	return z.quoAndRoundBig(&z.unscaled, pow.BigTen(uint64(-shift)))
}

// Scale returns x's scale.
func (x *Big) Scale() int32 { return x.scale }

// Scan implements fmt.Scanner.
func (z *Big) Scan(state fmt.ScanState, verb rune) error {
	toks, err := state.Token(true, func(c rune) bool {
		// Small enough range that we balance speed with accuracy.
		return c >= '+' && c <= 'e'
	})
	if err != nil {
		return err
	}
	if _, ok := z.SetString(string(toks)); !ok {
		if z.Context.Err != nil {
			return z.Context.Err
		}
		return errors.New("Big.Scan: SetString failed")
	}
	return nil
}

var _ fmt.Scanner = (*Big)(nil)

// Set sets z to x and returns z. The result might be rounded depending on z's
// Context.
func (z *Big) Set(x *Big) *Big {
	if debug {
		x.validate()
	}

	if z != x {
		z.compact = x.compact
		z.form = x.form
		z.scale = x.scale

		// Copy over unscaled if need be.
		if x.isInflated() {
			z.unscaled.Set(&x.unscaled)
		}
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

// SetFloat sets z to x and returns z.
func (z *Big) SetFloat(x *big.Float) *Big {
	if x.IsInf() {
		if x.Signbit() {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z
	}

	if x.Sign() == 0 {
		if x.Signbit() {
			z.form = nzero
		} else {
			z.form = zero
		}
		return z
	}

	z.scale = 0
	x0 := x
	if !x.IsInt() {
		x0 = new(big.Float).Copy(x)
		for !x0.IsInt() {
			x0.Mul(x0, c.TenFloat)
			z.scale++
		}
	}

	if mant, acc := x0.Int64(); acc == big.Exact {
		z.compact = mant
	} else {
		z.compact = c.Inflated
		x0.Int(&z.unscaled)
	}
	z.form = finite
	return z
}

// SetFloat64 sets z to exactly x. It's an exact conversion, meaning
// SetFloat64(0.1) results in a decimal with a value of
// 0.1000000000000000055511151231257827021181583404541015625. Use SetMantScale
// or SetString if you require exact conversions.
func (z *Big) SetFloat64(x float64) *Big {
	if x == 0 {
		z.form = zero
		return z
	}
	if math.IsNaN(x) {
		z.form = qnan
		return z.signal(InvalidOperation, ErrNaN{"SetFloat64(NaN)"})
	}
	if math.IsInf(x, 0) {
		if math.IsInf(x, 1) {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z.signal(InvalidOperation, errors.New("SetFloat(Inf)"))
	}
	return z.SetRat(new(big.Rat).SetFloat64(x))
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
	// Construct the required parts manually. The alternative is something like
	//
	//   num := new(Big).SetBigMantScale(x.Num(), 0)
	//   denom := new(Big).SetBigMantScale(x.Denom(), 0)
	//   return z.Quo(num, denom)
	//
	// But requires allocations we can avoid.

	if x.Sign() == 0 {
		z.form = zero
		return z
	}

	if x.IsInt() {
		z.form = finite
		return z.SetBigMantScale(x.Num(), 0)
	}

	xb, xc, xp := x.Num(), c.Inflated, 0
	if compat.IsInt64(xb) {
		xc = xb.Int64()
		xp = arith.Length(xc)
	} else {
		xp = arith.BigLength(xb)
	}

	yb, yc, yp := x.Denom(), c.Inflated, 0
	if compat.IsInt64(yb) {
		yc = yb.Int64()
		yp = arith.Length(yc)
	} else {
		yp = arith.BigLength(yb)
	}

	z.form = finite
	if xc == c.Inflated || yc == c.Inflated {
		return z.quoCore(xb, xc, 0, xp, yb, yc, 0, yp)
	}
	return z.quoCoreCompact(xc, 0, xp, yc, 0, yp)
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
// ``Inf'' and ``NaN'' map to ``+Inf'' and ``qNaN', respectively. NaN values may
// have optional diagnostic information, represented as trailing digits; for
// example, ``NaN123''. These digits are otherwise ignored but are included for
// robustness.
func (z *Big) SetString(s string) (*Big, bool) {
	// TODO(eric): write a scan(io.Reader) method.
	if debug {
		defer func() { z.validate() }()
	}

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

	switch parse.ParseSpecial(s) {
	case parse.QNaN:
		z.form = qnan
		return z, true
	case parse.SNaN:
		z.form = snan
		return z, true
	case parse.PInf:
		z.form = pinf
		return z, true
	case parse.NInf:
		z.form = ninf
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
// The result is undefined if x is a NaN value.
func (x *Big) Sign() int {
	if debug {
		x.validate()
	}

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
	if debug {
		x.validate()
	}

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
	var (
		b bytes.Buffer
		f = formatter{w: &b, prec: noPrec, width: noWidth}
		e = x.Context.OperatingMode.get().e
	)
	f.format(x, normal, e)
	return b.String()
}

var _ fmt.Stringer = (*Big)(nil)

// Sub sets z to x - y and returns z.
func (z *Big) Sub(x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}

	if x.form == finite && y.form == finite {
		z.form = finite
		if x.isCompact() && y.isCompact() {
			return z.subCompact(x, y).round()
		}
		return z.subBig(x, y).round()
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

// subCompact sets z to x - y and returns z.
func (z *Big) subCompact(x, y *Big) *Big {
	if debug {
		if x.compact == 0 || y.compact == 0 {
			panic("subCompact: operand == 0")
		}
	}

	xc, yc := x.compact, y.compact
	ok := false
	switch {
	case x.scale == y.scale:
		z.scale = x.scale
	case x.scale < y.scale:
		if xc, ok = checked.MulPow10(xc, uint64(y.scale-x.scale)); !ok {
			return z.subBig(x, y)
		}
		z.scale = y.scale
	case x.scale > y.scale:
		if yc, ok = checked.MulPow10(yc, uint64(x.scale-y.scale)); !ok {
			return z.subBig(x, y)
		}
		z.scale = x.scale
	}
	if z.compact, ok = checked.Sub(xc, yc); ok {
		if z.compact == 0 {
			z.form = zero
		}
		return z
	}
	if arith.Sub128(&z.unscaled, xc, yc).Sign() == 0 {
		z.form = zero
	}
	z.compact = c.Inflated
	return z
}

func (z *Big) subBig(x, y *Big) *Big {
	// TODO(eric): if debug { }

	xb, yb := &x.unscaled, &y.unscaled
	if x.isCompact() {
		xb = big.NewInt(x.compact)
	}
	if y.isCompact() {
		yb = big.NewInt(y.compact)
	}

	switch {
	case x.scale == y.scale:
		z.scale = x.scale
	case x.scale < y.scale:
		xb = checked.MulBigPow10(xb, uint64(y.scale-x.scale))
		z.scale = y.scale
	case x.scale > y.scale:
		yb = checked.MulBigPow10(yb, uint64(x.scale-y.scale))
		z.scale = x.scale
	}
	if z.unscaled.Sub(xb, yb).Sign() == 0 {
		z.form = zero
	}
	z.compact = c.Inflated
	return z
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (z *Big) UnmarshalText(data []byte) error {
	// TODO(eric): get rid of the allocation here.
	if _, ok := z.SetString(string(data)); !ok {
		return errors.New("Big.UnmarshalText: invalid decimal format")
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*Big)(nil)

// validate ensures x's internal state is correct. There's no need for it to
// have good performance since it's for debug == true only.
func (x *Big) validate() {
	defer func() {
		if err, ok := recover().(error); ok {
			fmt.Printf("%#v\n", x)
			panic(err)
		}
	}()
	if x.form == finite {
		if x.isCompact() && x.compact == 0 {
			panic("finite and compact == 0")
		}
		if x.isInflated() && x.unscaled.Sign() == 0 {
			panic("finite and unscaled == 0")
		}
	}
}
