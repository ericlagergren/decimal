// Package decimal provides a high-performance, arbitrary precision,
// floating-point decimal library.
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
// Arguments to Binary and Unary methods are allowed to alias, so the following
// is valid:
//
//     x := New(1, 0)
//     x.Add(x, x) // x == 2
//
// Unless otherwise specified, the only argument that will be modified is the
// reciever, meaning the following is valid and race-free:
//
//    x := New(1, 0)
//    var g1, g2 Big
//
//    go func() { g1.Add(x, x) }()
//    go func() { g2.Add(x, x) }()
//
// But this is not:
//
//    x := New(1, 0)
//    var g Big
//
//    go func() { g.Add(x, x) }() // BAD! RACE CONDITION!
//    go func() { g.Add(x, x) }() // BAD! RACE CONDITION!
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
)

// Big is a floating-point, arbitrary-precision decimal.
//
// It is represented as a number and a scale. If the scale is >= 0, it indicates
// the number of decimal digits after the radix. Otherwise, the number is
// multiplied by 10 to the power of the negation of the scale. More formally,
//
//   Big = number × 10**-scale
//
// with MinScale <= scale <= MaxScale. A Big may also be ±0, ±Infinity, or ±NaN
// (either quiet or signaling). Non-NaN Big values are ordered, defined as the
// result of x.Cmp(y).
//
// Additionally, each Big value has a contextual object which governs arithmetic
// operations.
type Big struct {
	// Big is laid out like this to take up as little memory as possible.

	// Context is the decimal's unique contextual object.
	Context Context

	// unscaled is only used if the decimal is too large to fit in compact.
	unscaled big.Int

	// compact is use if the value fits into an uint64. The scale does not
	// affect whether this field is used; typically, if a decimal has <= 20
	// digits this field will be used.
	compact uint64

	// exp is the negated scale, meaning
	//
	//  number × 10**exp = number ×  10**-scale
	exp int

	// form indicates whether a decimal is a finite number, an infinity, or a
	// NaN value and whether it's signed or not.
	form form

	// precision is a cached call to Precision.
	precision int
}

// form indicates whether a decimal is a finite number, an infinity, or a nan
// value and whether it's signed or not.
type form uint8

const (
	// Particular bits:
	//
	// 0: sign bit
	// 1: infinity
	// 2: signaling nan
	// 3: quiet nan
	// 4-7: unused

	finite form = 0 // default, all zeros; do not re-order this constant.

	signbit form = 1 << 0 // do not assign this; used to check for signedness.

	pinf form = 1 << 1         // may compare with ==, &, etc.
	ninf form = pinf | signbit // may compare with ==, &, etc.
	inf  form = pinf           // do not assign this; used to check for either infinity.

	snan  form = 1 << 2         // compare with bitwise & only due to ssnan
	qnan  form = 1 << 3         // compare with bitwise & only due to sqnan
	ssnan form = snan | signbit // primarily for printing, signbit
	sqnan form = qnan | signbit // primarily for printing, signbit
	nan   form = snan | qnan    // do not assign this; used to check for either NaN.

	special = inf | nan // do not assign this; used to check for a special value.
)

func (f form) String() string {
	switch f {
	case finite:
		return "finite"
	case snan:
		return "sNaN"
	case snan | signbit:
		return "-sNaN"
	case qnan:
		return "NaN"
	case qnan | signbit:
		return "-NaN"
	case pinf:
		return "Infinity"
	case ninf:
		return "-Infinity"
	default:
		return fmt.Sprintf("unknown form: %.16b", f)
	}
}

// Payload is a NaN value's payload.
type Payload uint64

const (
	addinfinf Payload = iota + 1
	mul0inf
	quo00
	quoinfinf
	quantinf
	quantminmax
	quantprec
	subinfinf
	absvalue
	addition
	comparison
	multiplication
	negation
	division
	quantization
	subtraction
)

var payloads = [...]string{
	addinfinf:      "addition of infinities with opposing signs",
	mul0inf:        "multiplication of zero with infinity",
	quo00:          "division of zero by zero",
	quoinfinf:      "division of infinity by infinity",
	quantinf:       "quantization of an infinity",
	quantminmax:    "quantization exceeds minimum or maximum scale",
	quantprec:      "quantization exceeds working precision",
	subinfinf:      "subtraction of infinities with opposing signs",
	absvalue:       "absolute value of NaN",
	addition:       "addition with NaN as an operand",
	comparison:     "comparison with NaN as an operand",
	multiplication: "multiplication with NaN as an operand",
	negation:       "negation with NaN as an operand",
	division:       "division with NaN as an operand",
	quantization:   "quantization with NaN as an operand",
	subtraction:    "subtraction with NaN as an operand",
}

func (p Payload) String() string {
	if p < Payload(len(payloads)) {
		return payloads[p]
	}
	return ""
}

// An ErrNaN is used when a decimal operation would lead to a NaN under IEEE-754
// rules. An ErrNaN implements the error interface.
type ErrNaN struct {
	Msg string
}

func (e ErrNaN) Error() string {
	return e.Msg
}

var _ error = ErrNaN{}

// CheckNaNs checks if either x or y is NaN. If so, it follows the rules of NaN
// handling set forth in the GDA specification. The second argument, y, may be
// nil. It returns true if either condition is a NaN.
func (z *Big) CheckNaNs(x, y *Big) bool {
	return z.checkNaNs(x, y, 0)
}

func (z *Big) checkNaNs(x, y *Big, op Payload) bool {
	f := (x.form | y.form) & nan
	if f == 0 {
		return false
	}

	form := qnan
	var cond Condition
	if f&snan != 0 {
		cond = InvalidOperation
		if x.form&snan != 0 {
			form |= (x.form & signbit)
		} else {
			form |= (y.form & signbit)
		}
	} else if x.form&nan != 0 {
		form |= (x.form & signbit)
	} else {
		form |= (y.form & signbit)
	}
	z.setNaN(cond, form, op)
	return true
}

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
		z.Context.Conditions |= Overflow | Inexact | Rounded
		return z
	}

	z.exp = MinScale
	z.compact = 0
	z.form = finite
	if neg {
		z.form |= signbit
	}
	z.Context.Conditions |= Underflow | Inexact | Rounded | Subnormal
	return z
}

// These methods are here to prevent typos.

func (x *Big) isCompact() bool  { return x.compact != c.Inflated }
func (x *Big) isInflated() bool { return !x.isCompact() }
func (x *Big) isSpecial() bool  { return x.form&(inf|nan) != 0 }

// norm normalizes z's mantissa and returns z.
func (z *Big) norm() *Big {
	if z.IsFinite() && z.isInflated() && iscompact(&z.unscaled) {
		z.compact = z.unscaled.Uint64()
	}
	return z
}

func (x *Big) adjusted() int { return (x.exp + x.Precision()) - 1 }
func (x *Big) emax() int     { return MaxScale - (x.Precision() - 1) }
func (x *Big) emin() int     { return MinScale - (x.Precision() - 1) }
func (x *Big) etiny() int    { return x.emin() - (precision(x) - 1) }
func (x *Big) etop() int     { return x.emax() - (precision(x) - 1) }

// Abs sets z to the absolute value of x and returns z.
func (z *Big) Abs(x *Big) *Big {
	if debug {
		x.validate()
	}
	if !z.checkNaNs(x, x, absvalue) {
		z.copyAbs(x)
		z.form = x.form & ^signbit
	}
	return z.round()
}

// Add sets z to x + y and returns z.
func (z *Big) Add(x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}

	if x.IsFinite() && y.IsFinite() {
		neg := z.add(x, x.Signbit(), y, y.Signbit())
		z.form = finite
		if neg {
			z.form |= signbit
		}
		z.precision = 0
		return z.norm().round()
	}

	// NaN + NaN
	// NaN + y
	// x + NaN
	if z.checkNaNs(x, y, addition) {
		return z
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 && x.form^y.form == signbit {
			// +Inf + -Inf
			// -Inf + +Inf
			return z.setNaN(InvalidOperation, qnan, addinfinf)
		}
		// ±Inf + y
		// +Inf + +Inf
		// -Inf + -Inf
		return z.Set(x)
	}
	// x + ±Inf
	return z.Set(y)
}

func (z *Big) add(x *Big, xn bool, y *Big, yn bool) (neg bool) {
	hi, lo := x, y
	hineg, loneg := xn, yn
	if hi.exp < lo.exp {
		hi, lo = lo, hi
		hineg, loneg = loneg, hineg
	}

	if neg, ok := z.tryTinyAdd(hi, hineg, lo, loneg); ok {
		return neg
	}

	if hi.isCompact() {
		if lo.isCompact() {
			neg = z.addCompact(hi.compact, hineg, lo.compact, loneg, uint64(hi.exp-lo.exp))
		} else {
			neg = z.addMixed(&lo.unscaled, loneg, lo.exp, hi.compact, hineg, hi.exp)
		}
	} else if lo.isCompact() {
		neg = z.addMixed(&hi.unscaled, hineg, hi.exp, lo.compact, loneg, lo.exp)
	} else {
		neg = z.addBig(&hi.unscaled, hineg, &lo.unscaled, loneg, uint64(hi.exp-lo.exp))
	}
	z.exp = lo.exp
	return neg
}

// tryTinyAdd returns true if hi + lo requires a huge shift that will produce
// the same results as a smaller shift. E.g., 3 + 0e+9999999999999999 with a
// precision of 5 doesn't need to be shifted by a large number.
func (z *Big) tryTinyAdd(hi *Big, hineg bool, lo *Big, loneg bool) (neg, ok bool) {
	if hi.compact == 0 {
		return false, false
	}

	exp := hi.exp - 1
	if hp, zp := hi.Precision(), precision(z); hp <= zp {
		exp += hp - zp - 1
	}

	if lo.adjusted() >= exp {
		return false, false
	}

	var tiny uint64
	if lo.compact != 0 {
		tiny = 1
	}
	tinyneg := loneg

	if hi.isCompact() {
		shift := uint64(hi.exp - exp)
		neg = z.addCompact(hi.compact, hineg, tiny, tinyneg, shift)
	} else {
		neg = z.addMixed(&hi.unscaled, hineg, hi.exp, tiny, tinyneg, exp)
	}
	z.exp = exp
	return neg, true
}

// addCompact sets z to x + y and returns z.
func (z *Big) addCompact(hi uint64, hineg bool, lo uint64, loneg bool, shift uint64) bool {
	neg := hineg
	if hi, ok := checked.MulPow10(hi, shift); ok {
		// Try regular addition and fall back to 128-bit addition.
		if loneg == hineg {
			if z.compact, ok = checked.Add(hi, lo); !ok {
				arith.Add128(&z.unscaled, hi, lo)
				z.compact = c.Inflated
			}
		} else {
			if z.compact, ok = checked.Sub(hi, lo); !ok {
				neg = !neg
				arith.Sub128(&z.unscaled, lo, hi)
				z.compact = c.Inflated
			}
		}
		// "Otherwise, the sign of a zero result is 0 unless either both
		// operands were negative or the signs of the operands were different
		// and the rounding is round-floor."
		return (z.compact == 0 && z.Context.RoundingMode == ToNegativeInf && neg) || neg
	}

	{
		hi := z.unscaled.SetUint64(hi)
		hi = checked.MulBigPow10(hi, hi, shift)
		if hineg == loneg {
			arith.Add(&z.unscaled, hi, lo)
		} else {
			// lo had to be promoted to a big.Int, so by definition it'll be
			// larger than hi. Therefore, we do not need to negate neg, nor do
			// we need to check to see if the result == 0.
			arith.Sub(&z.unscaled, hi, lo)
		}
	}
	z.compact = c.Inflated
	return neg
}

func (z *Big) addMixed(x *big.Int, xneg bool, xs int, y uint64, yn bool, ys int) bool {
	switch {
	case xs < ys:
		shift := uint64(ys - xs)
		if y0, ok := checked.MulPow10(y, shift); ok {
			y = y0
			break
		}

		// See comment in addCompact.
		yb := alias(&z.unscaled, x).SetUint64(y)
		yb = checked.MulBigPow10(yb, yb, shift)

		neg := xneg
		if xneg == yn {
			z.unscaled.Add(x, yb)
		} else {
			if x.Cmp(yb) >= 0 {
				z.unscaled.Sub(x, yb)
			} else {
				neg = !neg
				z.unscaled.Sub(yb, x)
			}
		}
		if z.unscaled.Sign() == 0 {
			z.compact = 0
		} else {
			z.compact = c.Inflated
		}
		return (z.compact == 0 && z.Context.RoundingMode == ToNegativeInf && neg) || neg
	case xs > ys:
		x = checked.MulBigPow10(&z.unscaled, x, uint64(xs-ys))
	}

	if xneg == yn {
		arith.Add(&z.unscaled, x, y)
	} else {
		// x > y
		arith.Sub(&z.unscaled, x, y)
	}

	z.compact = c.Inflated
	return xneg
}

func (z *Big) addBig(hi *big.Int, hineg bool, lo *big.Int, loneg bool, shift uint64) bool {
	if shift != 0 {
		hi = checked.MulBigPow10(alias(&z.unscaled, lo), hi, shift)
	}
	neg := hineg
	if hineg == loneg {
		z.unscaled.Add(hi, lo)
	} else {
		if hi.Cmp(lo) >= 0 {
			z.unscaled.Sub(hi, lo)
		} else {
			neg = !neg
			z.unscaled.Sub(lo, hi)
		}
	}
	if z.unscaled.Sign() == 0 {
		z.compact = 0
	} else {
		z.compact = c.Inflated
	}
	return z.compact != 0 && neg
}

// Class returns the ``class'' of x, which is one of the following:
//
//  sNaN
//  NaN
//  -Infinity
//  -Normal
//  -Subnormal
//  -Zero
//  +Zero
//  +Subnormal
//  +Normal
//  +Infinity
//
func (x *Big) Class() string {
	if !x.IsFinite() {
		return x.form.String()
	}
	if x.compact == 0 {
		if x.Signbit() {
			return "-Zero"
		}
		return "Zero"
	}
	if x.IsSubnormal() {
		if x.Signbit() {
			return "-Subnormal"
		}
		return "+Subnormal"
	}
	if x.Signbit() {
		return "-Normal"
	}
	return "+Normal"
}

// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
//
// It does not modify x or y. The result is undefined if either x or y are NaN.
// For an abstract comparison with NaN values, see misc.CmpTotal.
func (x *Big) Cmp(y *Big) int {
	return cmp(x, y, false)
}

// CmpAbs compares |x| and |y| and returns:
//
//   -1 if |x| <  |y|
//    0 if |x| == |y|
//   +1 if |x| >  |y|
//
// It does not modify x or y. The result is undefined if either x or y are NaN.
// For an abstract comparison with NaN values, see misc.CmpTotalAbs.
func (x *Big) CmpAbs(y *Big) int {
	return cmp(x, y, true)
}

// cmp is the implementation for both Cmp and CmpAbs.
func cmp(x, y *Big, abs bool) int {
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
	if x.checkNaNs(x, y, comparison) {
		return 0
	}

	// Fast path: Catches non-finite forms like zero and ±Inf, possibly signed.
	xs := x.ord(abs)
	ys := y.ord(abs)
	if xs != ys {
		if xs > ys {
			return +1
		}
		return -1
	}
	switch xs {
	case 0, +2, -2:
		return 0
	default:
		r := cmpabs(x, y)
		if xs < 0 && !abs {
			r = -r
		}
		return r
	}
}

func cmpabs(x, y *Big) int {
	// Same scales means we can compare straight across.
	if x.exp == y.exp {
		if x.isCompact() {
			if y.isCompact() {
				return arith.Cmp(x.compact, y.compact)
			}
			return -1 // y.isInflateed
		}
		if y.isCompact() {
			return +1 // !x.isCompact
		}
		return compat.BigCmpAbs(&x.unscaled, &y.unscaled)
	}

	// Signs are the same and the scales differ. Compare the lengths of their
	// integral parts; if they differ in length one number is larger.
	// E.g., 1234.01
	//        123.011
	xl := x.adjusted()
	yl := y.adjusted()

	if xl != yl {
		if xl < yl {
			return -1
		}
		return +1
	}

	diff := int64(x.exp) - int64(y.exp)
	shift := uint64(arith.Abs(diff))
	if pow.Safe(shift) && x.isCompact() && y.isCompact() {
		p, _ := pow.Ten(shift)
		if diff < 0 {
			return arith.AbsCmp128(x.compact, y.compact, p)
		}
		return -arith.AbsCmp128(y.compact, x.compact, p)
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
		yw = checked.MulBigPow10(&tmp, tmp.SetBits(copybits(yw)), shift).Bits()
	} else {
		xw = checked.MulBigPow10(&tmp, tmp.SetBits(copybits(xw)), shift).Bits()
	}
	return arith.CmpBits(xw, yw)
}

// Copy sets z to a copy of x and returns z.
func (z *Big) Copy(x *Big) *Big {
	if debug {
		x.validate()
	}
	z.copyAbs(x)
	z.form |= x.form & signbit
	return z
}

// copyAbs sets z to a copy of |x| and returns z.
func (z *Big) copyAbs(x *Big) *Big {
	if debug {
		x.validate()
	}

	if z != x {
		z.compact = x.compact
		z.form = x.form & ^signbit
		z.exp = x.exp
		if x.isInflated() {
			z.unscaled.Set(&x.unscaled)
		}
	}
	return z
}

// CopySign sets z to x with the sign of y and returns z. It accepts NaN values.
func (z *Big) CopySign(x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}
	z.Set(x)
	z.form = x.form | (y.form & signbit)
	return z

}

// Float64 returns x as a float64.
func (x *Big) Float64() float64 {
	if debug {
		x.validate()
	}

	if !x.IsFinite() {
		switch x.form {
		case pinf, ninf:
			return math.Inf(int(x.form & signbit))
		case snan, qnan:
			return math.NaN()
		case ssnan, sqnan:
			return math.Copysign(math.NaN(), -1)
		}
	}
	if x.isCompact() {
		if x.exp == 0 {
			return float64(x.compact)
		}
		const maxMantissa = 1 << 52
		if x.compact < maxMantissa {
			const maxPow10 = 22

			var f float64
			if x.exp > 0 && x.exp < maxPow10 {
				f = float64(x.compact) / math.Pow10(x.exp)
			} else if x.exp < 0 && x.exp < -maxPow10 {
				f = float64(x.compact) * math.Pow10(-x.exp)
			}
			if x.form&signbit != 0 {
				math.Copysign(f, -1)
			}
			return f
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
	case finite, finite | signbit:
		if x.compact == 0 {
			z.SetUint64(0)
		} else {
			z.SetRat(x.Rat(nil))
		}
	case pinf, ninf:
		z.SetInf(x.form == pinf)
	default: // snan, qnan, ssnan, sqnan:
		z.SetUint64(0)
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
		e       = sciE[x.Context.OperatingMode]
	)

	// If we need to left pad then we need to first write our string into an
	// empty buffer.
	tmpbuf := lpZero || lpSpace
	if tmpbuf {
		f.w = new(compat.Builder)
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
		if x.exp < 0 {
			f.prec -= x.exp
			if trail := x.Precision() + x.exp; trail >= f.prec {
				f.prec += trail
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
		if !hash && !plus {
			f.format(x, normal, 'e')
			break
		}

		// This is the easiest way of doing it. Note we can't use type Big Big,
		// even though it's declared inside a function. Go thinks it's recursive.
		// At least the fields are checked at compile time.
		type Big struct {
			Context   Context
			unscaled  big.Int
			compact   uint64
			exp       int
			form      form
			precision int
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

	// Need padding out to width.
	if f.n < int64(width) {
		switch pad := int64(width) - f.n; {
		case dash:
			io.CopyN(s, spaceReader{}, pad)
		case lpZero:
			io.CopyN(s, zeroReader{}, pad)
		case lpSpace:
			io.CopyN(s, spaceReader{}, pad)
		}
	}

	if tmpbuf {
		io.Copy(s, f.w.(*compat.Builder))
	}
}

var _ fmt.Formatter = (*Big)(nil)

// FMA sets z to (x * y) + u without any intermediate rounding.
func (z *Big) FMA(x, y, u *Big) *Big {
	z.mul(x, y, true)
	if z.Context.Conditions&InvalidOperation != 0 {
		return z
	}
	return z.Add(z, u)
}

// IsBig returns true if x, with its fractional part truncated, cannot fit
// inside an uint64. If x is an infinity or a NaN value the result is undefined.
func (x *Big) IsBig() bool {
	if debug {
		x.validate()
	}

	// !x.IsFinite() == zero, infinity, or nan
	if !x.IsFinite() {
		return false
	}
	// x.scale <= -20 is too large for a uint64.
	if x.exp <= -20 {
		return true
	}

	var v uint64
	if x.isCompact() {
		if x.exp == 0 {
			return false
		}
		v = x.compact
	} else {
		if x.unscaled.Cmp(c.MinInt64) <= 0 || x.unscaled.Cmp(c.MaxInt64) > 0 {
			return true
		}
		// Repeat this line twice so we don't have to call x.unscaled.Int64.
		if x.exp == 0 {
			return false
		}
		v = x.unscaled.Uint64()
	}
	_, ok := scalex(v, x.exp)
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

	if !x.IsFinite() {
		return z
	}

	if x.isCompact() {
		z.SetUint64(x.compact)
	} else {
		z.Set(&x.unscaled)
	}
	if x.Signbit() {
		z.Neg(z)
	}
	if x.exp == 0 {
		return z
	}
	if x.exp > 0 {
		return checked.MulBigPow10(z, z, uint64(x.exp))
	}
	return z.Quo(z, pow.BigTen(uint64(-x.exp)))
}

// Int64 returns x as an int64, truncating the fractional portion, if any. The
// result is undefined if x is an infinity, a NaN value, or if x does not fit
// inside an uint64.
func (x *Big) Int64() int64 {
	u := x.Uint64()
	if u > math.MaxInt64 {
		return 0
	}
	b := int64(u)
	if x.form&signbit != 0 {
		b = -b
	}
	return b
}

// Uint64 returns x as a uint64, truncating the fractional portion, if any. The
// result is undefined if x is an infinity, a NaN value, or if x does not fit
// inside an uint64.
func (x *Big) Uint64() uint64 {
	if debug {
		x.validate()
	}

	if !x.IsFinite() {
		return 0
	}

	// x might be too large to fit into an uint64 *now*, but rescaling x might
	// shrink it enough. See issue #20.
	if !x.isCompact() {
		return x.Int(nil).Uint64()
	}

	b := x.compact
	if x.exp == 0 {
		return b
	}
	b, ok := scalex(b, x.exp)
	if !ok {
		return 0
	}
	return b
}

// IsFinite returns true if x is finite.
func (x *Big) IsFinite() bool { return x.form & ^signbit == 0 }

// IsNormal returns true if x is normal.
func (x *Big) IsNormal() bool {
	return x.IsFinite() && x.adjusted() >= x.emin()
}

// IsSubnormal returns true if x is subnormal.
func (x *Big) IsSubnormal() bool {
	return x.IsFinite() && x.adjusted() < x.emin()
}

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
	return quiet >= 0 && x.form&qnan == qnan || quiet <= 0 && x.form&snan == snan
}

// IsInt reports whether x is an integer. Infinity and NaN values are not
// integers.
func (x *Big) IsInt() bool {
	if debug {
		x.validate()
	}

	if !x.IsFinite() {
		return false
	}

	// 5000, 420
	if x.exp >= 0 {
		return true
	}

	xp := x.Precision()

	// 0.001
	// 0.5
	if -x.exp >= xp {
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
	return xp <= -x.exp
}

// MarshalText implements encoding.TextMarshaler.
func (x *Big) MarshalText() ([]byte, error) {
	if debug {
		x.validate()
	}

	var (
		b compat.Builder
		f = formatter{w: &b, prec: noPrec, width: noWidth}
		e = sciE[x.Context.OperatingMode]
	)
	f.format(x, normal, e)
	return b.Bytes(), nil
}

// Mul sets z to x * y and returns z.
func (z *Big) Mul(x, y *Big) *Big {
	return z.mul(x, y, false).test()
}

// mul is the implementation of Mul, but with a boolean to toggle rounding. This
// is useful for FMA.
func (z *Big) mul(x, y *Big, fma bool) *Big {
	if debug {
		x.validate()
		y.validate()
	}

	sign := x.form&signbit ^ y.form&signbit

	if x.IsFinite() && y.IsFinite() {
		// Multiplication is simple, so inline it.
		if x.isCompact() {
			if y.isCompact() {
				if prod, ok := checked.Mul(x.compact, y.compact); ok {
					z.compact = prod
				} else {
					// Overflow: use 128 bit multiplication.
					arith.Mul128(&z.unscaled, x.compact, y.compact)
					z.compact = c.Inflated
				}
			} else { // y.isInflated
				arith.MulInt64(&z.unscaled, &y.unscaled, x.compact)
				z.compact = c.Inflated
			}
		} else if y.isCompact() { // x.isInflated
			arith.MulInt64(&z.unscaled, &x.unscaled, y.compact)
			z.compact = c.Inflated
		} else {
			z.unscaled.Mul(&x.unscaled, &y.unscaled)
			z.compact = c.Inflated
		}

		z.form = finite | sign
		z.exp = x.exp + y.exp
		z.precision = 0
		z.norm()
		if !fma {
			return z.round()
		}
		return z
	}

	// NaN * NaN
	// NaN * y
	// x * NaN
	if z.checkNaNs(x, y, multiplication) {
		return z
	}

	if (x.IsInf(0) && y.compact != 0) ||
		(y.IsInf(0) && x.compact != 0) ||
		(y.IsInf(0) && x.IsInf(0)) {
		// ±Inf * y
		// x * ±Inf
		// ±Inf * ±Inf
		return z.SetInf(sign != 0)
	}

	// 0 * ±Inf
	// ±Inf * 0
	return z.setNaN(InvalidOperation, qnan, mul0inf)
}

// Neg sets z to -x and returns z. If x is positive infinity, z will be set to
// negative infinity and visa versa. If x == 0, z will be set to zero as well.
// NaN will result in an error.
func (z *Big) Neg(x *Big) *Big {
	if debug {
		x.validate()
	}

	if !z.checkNaNs(x, x, negation) {
		z.copyAbs(x)
		z.form = x.form ^ signbit
	}
	return z.round()
}

// New creates a new Big decimal with the given value and scale. For example:
//
//  New(1234, 3) // 1.234
//  New(42, 0)   // 42
//  New(4321, 5) // 0.04321
//  New(-1, 0)   // -1
//  New(3, -10)  // 30 000 000 000
//
func New(value int64, scale int) *Big {
	return new(Big).SetMantScale(value, scale)
}

// Payload returns the payload of x, provided x is a NaN value. If x is not a
// NaN value the result is undefined.
func (x *Big) Payload() Payload {
	if !x.IsNaN(0) {
		return 0
	}
	return Payload(x.compact)
}

// Precision returns the precision of x. That is, it returns the number of
// digits in the unscaled form of x. x == 0 has a precision of 1. The result is
// undefined if x is an infinity or a NaN value.
func (x *Big) Precision() int {
	if debug {
		x.validate()
	}

	if !x.IsFinite() {
		return 0
	}
	if x.precision == 0 {
		if x.isCompact() {
			x.precision = arith.Length(x.compact)
		} else {
			x.precision = arith.BigLength(&x.unscaled)
		}
	}
	return x.precision
}

// Quantize sets z to the number equal in value and sign to z with the scale, n.
func (z *Big) Quantize(n int) *Big {
	if debug {
		z.validate()
	}

	n = -n
	if z.isSpecial() {
		if z.form&inf != 0 {
			return z.setNaN(InvalidOperation, qnan, quantinf)
		}
		z.checkNaNs(z, z, quantization)
		return z
	}

	if n > z.emax() || n < z.etiny() {
		return z.setNaN(InvalidOperation, qnan, quantminmax)
	}

	shift := z.exp - n
	if z.Precision()+shift > precision(z) {
		return z.setNaN(InvalidOperation, qnan, quantprec)
	}

	z.exp = n
	if shift == 0 || z.compact == 0 {
		return z
	}

	if shift < 0 {
		z.Context.Conditions |= Rounded
	}

	neg := z.Signbit()
	if z.isCompact() {
		if shift > 0 {
			if zc, ok := checked.MulPow10(z.compact, uint64(shift)); ok {
				z.compact = zc
				z.precision = 0
				return z
			}
			// shift < 0
		} else if yc, ok := pow.Ten(uint64(-shift)); ok {
			z.precision = 0
			return z.quoAndRoundCompact(z.compact, neg, yc, false)
		}
		z.unscaled.SetUint64(z.compact)
		z.compact = c.Inflated
	}

	if shift > 0 {
		_ = checked.MulBigPow10(&z.unscaled, &z.unscaled, uint64(shift))
		z.precision = 0
		return z
	}
	z.quoAndRoundBig(&z.unscaled, neg, pow.BigTen(uint64(-shift)), false)
	z.precision = 0
	return z
}

// Quo sets z to x / y and returns z.
func (z *Big) Quo(x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}

	sign := x.form&signbit ^ y.form&signbit

	if x.IsFinite() && y.IsFinite() {
		if y.compact == 0 {
			if x.compact == 0 {
				// 0 / 0
				return z.setNaN(InvalidOperation|DivisionUndefined, qnan, quo00)
			}
			// x / 0
			z.Context.Conditions |= DivisionByZero
			return z.SetInf(sign != 0)
		}
		if x.compact == 0 {
			// 0 / y
			z.exp = x.exp - y.exp
			z.compact = 0
			z.form = finite | sign
			return z.test()
		}

		if x.isCompact() && y.isCompact() {
			z.quoCompact(x, y)
		} else {
			z.quo(x, y)
		}
		z.precision = 0
		return z
	}

	// NaN / NaN
	// NaN / y
	// x / NaN
	if z.checkNaNs(x, y, division) {
		return z
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 {
			// ±Inf / ±Inf
			return z.setNaN(InvalidOperation, qnan, quoinfinf)
		}
		// ±Inf / y
		return z.SetInf(sign != 0)
	}
	// x / ±Inf
	z.form = finite | sign
	z.exp = z.etiny()
	z.Context.Conditions |= Clamped
	return z
}

func (z *Big) quoCompact(x, y *Big) *Big {
	return z.quoCoreCompact(
		x.compact, x.Signbit(), x.exp, x.Precision(),
		y.compact, y.Signbit(), y.exp, y.Precision(),
	)
}

// quoCoreCompact implements division of two compact decimals.
func (z *Big) quoCoreCompact(
	x uint64, xn bool, xs, xp int,
	y uint64, yn bool, ys, yp int,
) *Big {
	if cmpNorm(x, xp, y, yp) {
		yp--
	}

	zp := precision(z)
	shift := zp + yp - xp
	z.exp = (xs - ys) - shift
	if shift > 0 {
		if sx, ok := checked.MulPow10(x, uint64(shift)); ok {
			return z.quoAndRoundCompact(sx, xn, y, yn)
		}
		xb := z.unscaled.SetUint64(x)
		xb = checked.MulBigPow10(xb, xb, uint64(shift))
		return z.quoAndRoundBig(xb, xn, new(big.Int).SetUint64(y), yn)
	}
	if shift < 0 {
		if sy, ok := checked.MulPow10(y, uint64(-shift)); ok {
			return z.quoAndRoundCompact(x, xn, sy, yn)
		}
		yb := z.unscaled.SetUint64(y)
		yb = checked.MulBigPow10(yb, yb, uint64(-shift))
		return z.quoAndRoundBig(new(big.Int).SetUint64(x), xn, yb, yn)
	}
	return z.quoAndRoundCompact(x, xn, y, yn)
}

func (z *Big) quoAndRoundCompact(x uint64, xneg bool, y uint64, yneg bool) *Big {
	z.form = finite

	pos := xneg == yneg
	if !pos {
		z.form |= signbit
	}

	z.compact = x / y
	r := x % y
	if r == 0 {
		return z
	}

	z.Context.Conditions |= Inexact | Rounded
	if z.Context.RoundingMode == ToZero {
		return z
	}

	rc := 1
	if r2, ok := checked.Mul(r, 2); ok {
		rc = arith.Cmp(r2, y)
	}

	if z.needsInc(rc, pos) {
		z.Context.Conditions |= Rounded
		if pos {
			z.compact++
		} else {
			z.compact--
		}
	}
	return z
}

func (z *Big) quo(x, y *Big) *Big {
	return z.quoCore(
		&x.unscaled, x.compact, x.Signbit(), x.exp, x.Precision(),
		&y.unscaled, y.compact, y.Signbit(), y.exp, y.Precision(),
	)
}

// see quoCompactCore. xc and yc override xb and yb, respectively, if they !=
// c.Inflated. If both xc and yc != c.Inflated quoCompactCore will be called.
// This method should be used sparingly.
func (z *Big) quoCore(
	xb *big.Int, xc uint64, xn bool, xs, xp int,
	yb *big.Int, yc uint64, yn bool, ys, yp int,
) *Big {
	z.precision = 0
	// TODO(eric): re-work the quo methods. I don't like how they're laid out.
	if xc != c.Inflated {
		if yc != c.Inflated {
			return z.quoCoreCompact(xc, xn, xs, xp, yc, yn, ys, yp)
		}
		xb = new(big.Int).SetUint64(xc)
	}
	if yc != c.Inflated {
		yb = new(big.Int).SetUint64(yc)
	}

	if cmpNormBig(&z.unscaled, xb, xp, yb, yp) {
		yp--
	}

	zp := precision(z)
	shift := zp + yp - xp
	z.exp = (xs - ys) - shift
	if shift > 0 {
		tmp := alias(&z.unscaled, yb)
		xb = checked.MulBigPow10(tmp, xb, uint64(shift))
	} else {
		shift = xp - zp - yp
		tmp := alias(&z.unscaled, xb)
		yb = checked.MulBigPow10(tmp, yb, uint64(shift))
	}
	return z.quoAndRoundBig(xb, xn, yb, yn)
}

func (z *Big) quoAndRoundBig(x *big.Int, xneg bool, y *big.Int, yneg bool) *Big {
	z.form = finite
	z.compact = c.Inflated

	pos := xneg == yneg
	if !pos {
		z.form |= signbit
	}

	// q == z.unscaled, but it's easier to type q.
	q, r := z.unscaled.QuoRem(x, y, new(big.Int))
	if r.Sign() == 0 {
		return z.norm()
	}

	z.Context.Conditions |= Inexact | Rounded
	if z.Context.RoundingMode == ToZero {
		return z.norm()
	}

	var rc int
	rv := r.Uint64()
	// Drop into integers if we can.
	if arith.IsUint64(r) && arith.IsUint64(y) && rv <= math.MaxUint64/2 {
		rc = arith.Cmp(rv*2, y.Uint64())
	} else {
		rc = compat.BigCmpAbs(r.Mul(r, c.TwoInt), y)
	}

	if z.needsInc(rc, pos) {
		z.Context.Conditions |= Rounded
		if pos {
			arith.Add(q, q, 1)
		} else {
			arith.Sub(q, q, 1)
		}
	}
	return z.norm()
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

	if !x.IsFinite() {
		return z.SetInt64(0)
	}

	var num *big.Int
	if x.isCompact() {
		num = new(big.Int).SetUint64(x.compact)
	} else {
		num = new(big.Int).Set(&x.unscaled)
	}
	if x.Signbit() {
		num.Neg(num)
	}

	var denom *big.Int
	if x.exp < 0 {
		if shift, ok := pow.Ten(uint64(-x.exp)); ok {
			denom = new(big.Int).SetUint64(shift)
		} else {
			tmp := new(big.Int).SetUint64(uint64(-x.exp))
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
func Raw(x *Big) (uint64, *big.Int) {
	return x.compact, &x.unscaled
}

func (z *Big) round() *Big {
	if mode(z) == GDA {
		if zp := precision(z); zp != UnlimitedPrecision {
			return z.Round(zp)
		}
	}
	return z
}

// Round rounds z down to n digits of precision and returns z. The result is
// undefined if z is not finite. No rounding will occur if n <= 0. The result of
// Round will always be within the interval [⌊10**x⌋, z] where x = the precision
// of z.
func (z *Big) Round(n int) *Big {
	if debug {
		z.validate()
	}

	if n <= 0 || z.isSpecial() {
		return z
	}

	zp := z.Precision()
	if zp <= n {
		return z
	}

	shift := zp - n
	if shift > MaxScale {
		return z.xflow(false, true)
	}
	z.exp += shift

	z.Context.Conditions |= Rounded

	neg := z.Signbit()
	if z.isCompact() {
		if z.compact == 0 {
			z.precision = n
			return z
		}
		if val, ok := pow.Ten(uint64(shift)); ok {
			z.precision = n
			return z.quoAndRoundCompact(z.compact, neg, val, false)
		}
		z.unscaled.SetUint64(z.compact)
		z.compact = c.Inflated
	}
	z.quoAndRoundBig(&z.unscaled, neg, pow.BigTen(uint64(shift)), false)
	z.precision = n
	return z
}

// Scale returns x's scale.
func (x *Big) Scale() int { return -x.exp }

// Scan implements fmt.Scanner.
func (z *Big) Scan(state fmt.ScanState, verb rune) error {
	return z.scan(byteReader{state})
}

var _ fmt.Scanner = (*Big)(nil)

// Set sets z to x and returns z. The result might be rounded depending on z's
// Context, and even if z aliases x.
func (z *Big) Set(x *Big) *Big {
	if debug {
		x.validate()
	}

	if z != x {
		z.compact = x.compact
		z.form = x.form
		z.exp = x.exp

		// Copy over unscaled if need be.
		if x.isInflated() {
			z.unscaled.Set(&x.unscaled)
		}
	}
	return z.round()
}

// SetBigMantScale sets z to the given value and scale.
func (z *Big) SetBigMantScale(value *big.Int, scale int) *Big {
	if iscompact(value) {
		z.compact = value.Uint64()
	} else {
		z.unscaled.Abs(value)
		z.compact = c.Inflated
	}
	z.form = finite
	if value.Sign() < 0 {
		z.form |= signbit
	}
	z.exp = -scale
	z.precision = 0
	return z
}

// SetFloat sets z to x and returns z.
func (z *Big) SetFloat(x *big.Float) *Big {
	z.precision = 0

	if x.IsInf() {
		if x.Signbit() {
			z.form = ninf
		} else {
			z.form = pinf
		}
		return z
	}

	neg := x.Signbit()
	if x.Sign() == 0 {
		z.compact = 0
		if neg {
			z.form |= signbit
		}
		return z
	}

	z.exp = 0
	x0 := x
	if !x.IsInt() {
		x0 = new(big.Float).Copy(x)
		for !x0.IsInt() {
			x0.Mul(x0, c.TenFloat)
			z.exp--
		}
	}

	if mant, acc := x0.Uint64(); acc == big.Exact {
		z.compact = mant
	} else {
		z.compact = c.Inflated
		x0.Int(&z.unscaled)
	}
	z.form = finite
	if neg {
		z.form |= signbit
	}
	return z
}

// SetFloat64 sets z to exactly x. It's an exact conversion, meaning
// SetFloat64(0.1) results in a decimal with a value of
// 0.1000000000000000055511151231257827021181583404541015625. Use SetMantScale
// or SetString if you require exact conversions.
func (z *Big) SetFloat64(x float64) *Big {
	z.precision = 0

	if x == 0 {
		z.compact = 0
		z.form = finite
		if math.Signbit(x) {
			z.form |= signbit
		}
		return z
	}
	if math.IsNaN(x) {
		var sign form
		if math.Signbit(x) {
			sign = 1
		}
		return z.setNaN(InvalidOperation, qnan|sign, 0)
	}
	if math.IsInf(x, 0) {
		if math.IsInf(x, 1) {
			z.form = pinf
		} else {
			z.form = ninf
		}
		z.Context.Conditions |= InvalidOperation
		return z
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
func (z *Big) SetMantScale(value int64, scale int) *Big {
	z.precision = 0
	z.compact = arith.Abs(value)
	z.form = finite
	if value < 0 {
		z.form |= signbit
	}
	z.exp = -scale
	return z
}

// setNaN is an internal NaN-setting method that panics when the OperatingMode
// is Go.
func (z *Big) setNaN(c Condition, f form, p Payload) *Big {
	z.form = f
	z.compact = uint64(p)
	z.Context.Conditions |= c
	if z.Context.OperatingMode == Go {
		panic(ErrNaN{Msg: z.Context.Conditions.String()})
	}
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
	// Catches 0 case.
	if x.IsInt() {
		z.form = finite
		return z.SetBigMantScale(x.Num(), 0)
	}

	var num, denom Big
	num.SetBigMantScale(x.Num(), 0)
	denom.SetBigMantScale(x.Denom(), 0)
	return z.Quo(&num, &denom)
}

// SetScale sets z's scale to scale and returns z.
func (z *Big) SetScale(scale int) *Big {
	z.exp = -scale
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
// 	NaN
// 	qNaN
// 	sNaN
//
// Each value may be preceded by an optional sign, ``-'' or ``+''. ``Inf'' and
// ``NaN'' map to ``+Inf'' and ``qNaN'', respectively. NaN values may have
// optional diagnostic information, represented as trailing digits; for example,
// ``NaN123''. These digits are otherwise ignored but are included for
// robustness.
func (z *Big) SetString(s string) (*Big, bool) {
	if err := z.scan(strings.NewReader(s)); err != nil {
		return nil, false
	}
	return z, true
}

// ord returns similar to Sign except -Inf is -2 and +Inf is +2.
func (x *Big) ord(abs bool) int {
	if x.form&inf != 0 {
		if x.form == pinf || abs {
			return +2
		}
		return -2
	}
	r := x.Sign()
	if abs && r < 0 {
		r = -r
	}
	return r
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

	if (x.IsFinite() && x.compact == 0) || x.IsNaN(0) {
		return 0
	}
	if x.form&signbit != 0 {
		return -1
	}
	return 1
}

// Signbit returns true if x is negative, negative infinity, negative zero, or
// negative NaN.
func (x *Big) Signbit() bool {
	if debug {
		x.validate()
	}
	return x.form&signbit != 0
}

// String returns the string representation of x. It's equivalent to the %s verb
// discussed in the Format method's documentation. Special cases depend on the
// OperatingMode.
func (x *Big) String() string {
	var (
		b compat.Builder
		f = formatter{w: &b, prec: noPrec, width: noWidth}
		e = sciE[x.Context.OperatingMode]
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

	if x.IsFinite() && y.IsFinite() {
		neg := z.add(x, x.Signbit(), y, !y.Signbit())
		z.form = finite
		if neg {
			z.form |= signbit
		}
		z.precision = 0
		return z.norm().round()
	}

	// NaN - NaN
	// NaN - y
	// x - NaN
	if z.checkNaNs(x, y, subtraction) {
		return z
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 && (x.form&signbit == y.form&signbit) {
			// -Inf - -Inf
			// -Inf - -Inf
			return z.setNaN(InvalidOperation, qnan, subinfinf)
		}
		// ±Inf - y
		// -Inf - +Inf
		// +Inf - -Inf
		return z.Set(x)
	}
	// x - ±Inf
	return z.Neg(y)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (z *Big) UnmarshalText(data []byte) error {
	return z.scan(bytes.NewReader(data))
}

var _ encoding.TextUnmarshaler = (*Big)(nil)

// validate ensures x's internal state is correct. There's no need for it to
// have good performance since it's for debug == true only.
func (x *Big) validate() {
	defer func() {
		if err := recover(); err != nil {
			pc, _, _, ok := runtime.Caller(4)
			if caller := runtime.FuncForPC(pc); ok && caller != nil {
				fmt.Println("called by:", caller.Name())
			}
			type Big struct {
				Context   Context
				unscaled  big.Int
				compact   uint64
				exp       int
				form      form
				precision int
			}
			fmt.Printf("%#v\n", (*Big)(x))
			panic(err)
		}
	}()
	switch x.form {
	case finite, finite | signbit:
		if x.isInflated() {
			if iscompact(&x.unscaled) {
				panic(fmt.Sprintf("inflated but unscaled == %d", x.unscaled.Uint64()))
			}
			if x.unscaled.Sign() < 0 {
				panic("x.unscaled.Sign() < 0")
			}
		}
	case snan, ssnan, qnan, sqnan, pinf, ninf:
		// OK
	case nan:
		panic(x.form.String())
	default:
		panic(fmt.Sprintf("invalid form %s", x.form))
	}
}
