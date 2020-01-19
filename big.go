package decimal

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/c"
)

// Big is a floating-point, arbitrary-precision decimal.
//
// It is represented as a number and a scale. If the scale is >= 0, it indicates
// the number of decimal digits after the radix. Otherwise, the number is
// multiplied by 10 to the power of the negation of the scale. More formally,
//
//    Big = number × 10**-scale
//
// with MinScale <= scale <= MaxScale. A Big may also be ±0, ±Infinity, or ±NaN
// (either quiet or signaling). Non-NaN Big values are ordered, defined as the
// result of x.Cmp(y).
//
// Additionally, each Big value has a contextual object which governs arithmetic
// operations.
type Big struct {
	// Context is the decimal's unique contextual object.
	Context Context

	// unscaled is only used if the decimal is too large to fit in compact.
	unscaled big.Int

	// compact is use if the value fits into an uint64. The scale does not
	// affect whether this field is used; typically, if a decimal has <= 20
	// digits this field will be used.
	compact uint64

	// exp is the negated scale. This means
	//
	//    number × 10**exp = number × 10**-scale
	//
	exp int

	// precision is the current precision.
	precision int

	// form indicates whether a decimal is a finite number, an infinity, or a
	// NaN value and whether it's signed or not.
	form form
}

var (
	_ fmt.Formatter            = (*Big)(nil)
	_ fmt.Scanner              = (*Big)(nil)
	_ fmt.Stringer             = (*Big)(nil)
	_ json.Unmarshaler         = (*Big)(nil)
	_ encoding.TextUnmarshaler = (*Big)(nil)
	_ decomposer               = (*Big)(nil)
)

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
	// GDA versions. Go needs to be handled manually.
	switch f {
	case finite:
		return "finite"
	case finite | signbit:
		return "-finite"
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
		return fmt.Sprintf("unknown form: %0.8b", f)
	}
}

// Payload is a NaN value's payload. A zero value indicates no payload.
type Payload uint64

//go:generate stringer -type Payload -linecomment

const (
	addinfinf      Payload = iota + 1 // addition of infinities with opposing signs
	mul0inf                           // multiplication of zero with infinity
	quo00                             // division of zero by zero
	quoinfinf                         // division of infinity by infinity
	quantinf                          // quantization of an infinity
	quantminmax                       // quantization exceeds minimum or maximum scale
	quantprec                         // quantization exceeds working precision
	subinfinf                         // subtraction of infinities with opposing signs
	absvalue                          // absolute value of NaN
	addition                          // addition with NaN as an operand
	comparison                        // comparison with NaN as an operand
	multiplication                    // multiplication with NaN as an operand
	negation                          // negation with NaN as an operand
	division                          // division with NaN as an operand
	quantization                      // quantization with NaN as an operand
	subtraction                       // subtraction with NaN as an operand
	quorem_                           // integer division or remainder has too many digits
	reminfy                           // remainder of infinity
	remx0                             // remainder by zero
	quotermexp                        // division with unlimited precision has a non-terminating decimal expansion
	invctxpltz                        // operation with a precision less than zero
	invctxpgtu                        // operation with a precision greater than MaxPrecision
	invctxrmode                       // operation with an invalid RoundingMode
	invctxomode                       // operation with an invalid OperatingMode
	invctxsltu                        // operation with a scale lesser than MinScale
	invctxsgtu                        // operation with a scale greater than MaxScale
	reduction                         // reduction with NaN as an operand
	quointprec                        // result of integer division was larger than the desired precision
	remprec                           // result of remainder operation was larger than the desired precision
)

// TODO(eric): if math.ErrNaN ever allows setting the msg field, perhaps we
// should use that instead?

// An ErrNaN is as the value passed to ``panic'' when the operating mode is set
// to Go and a decimal operation occurs that would lead to a NaN under IEEE-754
// rules.
//
// ErrNaN implements the error interface.
type ErrNaN struct{ Msg string }

func (e ErrNaN) Error() string { return e.Msg }

var _ error = ErrNaN{}

// Append appends to buf the string form of x, as generated by x.Text, and
// returns the extended buffer.
func (x *Big) Append(buf []byte, fmt byte, prec int) []byte {
	if x == nil {
		return append(buf, "<nil>"...)
	}

	if x.isSpecial() {
		switch x.Context.OperatingMode {
		case GDA:
			buf = append(buf, x.form.String()...)
			if x.IsNaN(0) && x.compact != 0 {
				buf = strconv.AppendUint(buf, x.compact, 10)
			}
		case Go:
			if x.IsNaN(0) {
				buf = append(buf, "NaN"...)
			} else if x.IsInf(0) {
				buf = append(buf, "+Inf"...)
			} else {
				buf = append(buf, "Inf"...)
			}
		default:
			buf = append(buf, '?')
		}
		return buf
	}

	neg := x.Signbit()
	if neg {
		buf = append(buf, '-')
	}

	dec := make([]byte, 0, x.Precision())
	if x.isCompact() {
		dec = strconv.AppendUint(dec[:0], uint64(x.compact), 10)
	} else {
		dec = x.unscaled.Append(dec[:0], 10)
	}

	// Normalize x such that 0 <= x < 1.
	//
	//    actual     decimal        normalized      exp+len
	//    ---------- -------------- --------------- ---------
	//    123400     1234*10^2      0.1234*10^6      2+4=6
	//    12340      1234*10^1      0.1234*10^5      1+4=5
	//    1234       1234*10^0      0.1234*10^4      0+4=4
	//    1234.0     12340*10^-1    0.12340*10^4    -1+5=4
	//    1234.00    123400*10^-2   0.123400*10^4   -2+6=4
	//    123.4      1234*10^-1     0.1234*10^3     -1+4=3
	//    12.34      1234*10^-2     0.1234*10^2     -2+4=2
	//    1.234      1234*10^-3     0.1234*10^1     -3+4=1
	//    1.0        10*10-1        0.10*10^1       -1+2=1
	//    0.0        00*10^-1       0.00*10^1       -1+2=1
	//    0.1234     1234*10^-4     0.1234*10^0     -4+4=0
	//    0.001234   1234*10^-5     0.1234*10^-1    -5+4=-1
	//
	norm := x.exp + len(dec)

	if prec < 0 {
		dec = roundShortest(dec, norm)
		switch fmt {
		case 'e', 'E':
			prec = len(dec) - 1
		case 'f', 'F':
			prec = max(len(dec)-norm, 0)
		case 'g', 'G':
			prec = len(dec)
		}
	} else {
		switch fmt {
		case 'e', 'E':
			dec = round(dec, 1+prec)
		case 'f', 'F':
			dec = round(dec, len(dec)+norm+prec)
		case 'g', 'G':
			if prec == 0 {
				prec = 1
			}
			dec = round(dec, prec)
		}
	}

	switch fmt {
	case 'g', 'G':
		// Decide whether to use regular or exponential notation.
		//
		//    Next, the adjusted exponent is calculated; this is the exponent,
		//    plus the number of characters in the converted coefficient, less
		//    one. That is, exponent+(clength-1), where clength is the length of
		//    the coefficient in decimal digits.
		//
		//    If the exponent is less than or equal to zero and the adjusted
		//    exponent is greater than or equal to -6 the number will be converted
		//    to a character form without using exponential notation.
		//
		// - http://speleotrove.com/decimal/daconvs.html#reftostr
		if -6 <= norm-1 && x.exp <= 0 {
			if prec > norm {
				prec = len(dec)
			}
			return fmtF(buf, max(prec-norm, 0), dec, norm)
		}
		if prec > len(dec) {
			prec = len(dec)
		}
		return fmtE(buf, fmt+'e'-'g', prec-1, dec, norm)
	case 'e', 'E':
		return fmtE(buf, fmt, prec, dec, norm)
	case 'f', 'F':
		return fmtF(buf, prec, dec, norm)
	default:
		if neg {
			buf = buf[:len(buf)-1]
		}
		return append(buf, '%', fmt)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func round(dec []byte, n int) []byte {
	// TODO(eric): actually round
	if n < 0 || n >= len(dec) {
		return dec
	}
	return dec[:n]
}

func roundShortest(dec []byte, exp int) []byte {
	// exp < 0        : 0.[000]ddd
	// exp >= len(dec): ddddd
	if exp < 0 || exp >= len(dec) {
		return dec
	}
	i := len(dec) - 1
	for i >= len(dec)-exp {
		if dec[i] != '0' {
			break
		}
		i--
	}
	return dec[:i+1]
}

func fmtE(buf []byte, fmt byte, prec int, dec []byte, exp int) []byte {
	if len(dec) > 0 {
		buf = append(buf, dec[0])
	} else {
		buf = append(buf, '0')
	}
	if prec > 0 {
		buf = append(buf, '.')
		i := 1
		m := min(len(dec), prec+1)
		if i < m {
			buf = append(buf, dec[i:m]...)
			i = m
		}
		buf = appendZeros(buf, prec-i+1) // i <= prec
	}

	if len(dec) > 0 {
		exp--
	}
	buf = append(buf, fmt)
	if exp < 0 {
		buf = append(buf, '-')
		exp = -exp
	} else {
		buf = append(buf, '+')
	}
	if exp < 10 {
		buf = append(buf, '0') // 01, 02, ..., 09
	}
	return strconv.AppendUint(buf, uint64(exp), 10) // exp >= 0
}

func fmtF(buf []byte, prec int, dec []byte, exp int) []byte {
	if exp > 0 {
		m := min(len(dec), exp)
		buf = append(buf, dec[:m]...)
		buf = appendZeros(buf, exp-m)
	} else {
		buf = append(buf, '0')
	}

	if prec > 0 {
		buf = append(buf, '.')
		for i := 0; i < prec; i++ {
			c := byte('0')
			if j := i + exp; 0 <= j && j < len(dec) {
				c = dec[j]
			}
			buf = append(buf, c)
		}
	}
	return buf
}

// appeendZeros appends n '0's to buf.
func appendZeros(buf []byte, n int) []byte {
	const zeros = "0000000000000000"
	for n >= 0 {
		if n < len(zeros) {
			buf = append(buf, zeros[:n]...)
		} else {
			buf = append(buf, zeros...)
		}
		n -= len(zeros)
	}
	return buf
}

// CheckNaNs checks if either x or y is NaN. If so, it follows the rules of NaN
// handling set forth in the GDA specification. The second argument, y, may be
// nil. It returns true if either condition is a NaN.
func (z *Big) CheckNaNs(x, y *Big) bool {
	return z.invalidContext(z.Context) || z.checkNaNs(x, y, 0)
}

func (z *Big) checkNaNs(x, y *Big, op Payload) bool {
	var yform form
	if y != nil {
		yform = y.form
	}
	f := (x.form | yform) & nan
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

func (z *Big) xflow(exp int, over, neg bool) *Big {
	// over == overflow
	// neg == intermediate result < 0
	if over {
		// TODO(eric): actually choose the largest finite number in the current
		// precision. This is legacy now.
		//
		// NOTE(eric): in some situations, the decimal library tells us to set
		// z to "the largest finite number that can be represented in the
		// current precision..." Use signed Infinity instead.
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

	var sign form
	if neg {
		sign = signbit
	}
	z.setZero(sign, exp)
	z.Context.Conditions |= Underflow | Inexact | Rounded | Subnormal
	return z
}

// These methods are here to prevent typos.

func (x *Big) isCompact() bool  { return x.compact != c.Inflated }
func (x *Big) isInflated() bool { return !x.isCompact() }
func (x *Big) isSpecial() bool  { return x.form&(inf|nan) != 0 }
func (x *Big) isZero() bool     { return x.compact == 0 }

func (x *Big) adjusted() int { return (x.exp + x.Precision()) - 1 }
func (c Context) etiny() int { return MinScale - (precision(c) - 1) }

// Abs sets z to the absolute value of x and returns z.
func (z *Big) Abs(x *Big) *Big {
	if debug {
		x.validate()
	}
	if !z.invalidContext(z.Context) && !z.checkNaNs(x, x, absvalue) {
		z.Context.round(z.copyAbs(x))
	}
	return z
}

// Add sets z to x + y and returns z.
func (z *Big) Add(x, y *Big) *Big { return z.Context.Add(z, x, y) }

// Class returns the ``class'' of x, which is one of the following:
//
//    sNaN
//    NaN
//    -Infinity
//    -Normal
//    -Subnormal
//    -Zero
//    +Zero
//    +Subnormal
//    +Normal
//    +Infinity
//
func (x *Big) Class() string {
	if x.IsNaN(0) {
		if x.IsNaN(+1) {
			return "NaN"
		}
		return "sNaN"
	}
	if x.Signbit() {
		if x.IsInf(0) {
			return "-Infinity"
		}
		if x.isZero() {
			return "-Zero"
		}
		if x.IsSubnormal() {
			return "-Subnormal"
		}
		return "-Normal"
	}
	if x.IsInf(0) {
		return "+Infinity"
	}
	if x.isZero() {
		return "+Zero"
	}
	if x.IsSubnormal() {
		return "+Subnormal"
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
func (x *Big) Cmp(y *Big) int { return cmp(x, y, false) }

// CmpAbs compares |x| and |y| and returns:
//
//   -1 if |x| <  |y|
//    0 if |x| == |y|
//   +1 if |x| >  |y|
//
// It does not modify x or y. The result is undefined if either x or y are NaN.
// For an abstract comparison with NaN values, see misc.CmpTotalAbs.
func (x *Big) CmpAbs(y *Big) int { return cmp(x, y, true) }

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
	if (x.form|y.form)&nan != 0 {
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
		return x.unscaled.CmpAbs(&y.unscaled)
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
	if arith.Safe(shift) && x.isCompact() && y.isCompact() {
		p, _ := arith.Pow10(shift)
		if diff < 0 {
			return arith.CmpShift(x.compact, y.compact, p)
		}
		return -arith.CmpShift(y.compact, x.compact, p)
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
		yw = arith.MulBigPow10(&tmp, tmp.SetBits(copybits(yw)), shift).Bits()
	} else {
		xw = arith.MulBigPow10(&tmp, tmp.SetBits(copybits(xw)), shift).Bits()
	}
	return arith.CmpBits(xw, yw)
}

// Copy sets z to a copy of x and returns z.
func (z *Big) Copy(x *Big) *Big {
	if debug {
		x.validate()
	}
	if z != x {
		sign := x.form & signbit
		z.copyAbs(x)
		z.form |= sign
	}
	return z
}

// copyAbs sets z to a copy of |x| and returns z.
func (z *Big) copyAbs(x *Big) *Big {
	if z != x {
		z.precision = x.Precision()
		z.exp = x.exp
		z.compact = x.compact
		if x.IsFinite() && x.isInflated() {
			z.unscaled.Set(&x.unscaled)
		}
	}
	z.form = x.form & ^signbit
	return z
}

// CopySign sets z to x with the sign of y and returns z. It accepts NaN values.
func (z *Big) CopySign(x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}
	// Pre-emptively capture signbit in case z == y.
	sign := y.form & signbit
	z.copyAbs(x)
	z.form |= sign
	return z
}

// Float64 returns x as a float64 and a bool indicating whether x can fit into
// a float64 without truncation, overflow, or underflow. Special values are
// considered exact; however, special values that occur because the magnitude of
// x is too large to be represented as a float64 are not.
func (x *Big) Float64() (f float64, ok bool) {
	if debug {
		x.validate()
	}

	if !x.IsFinite() {
		switch x.form {
		case pinf, ninf:
			return math.Inf(int(x.form & signbit)), true
		case snan, qnan:
			return math.NaN(), true
		case ssnan, sqnan:
			return math.Copysign(math.NaN(), -1), true
		}
	}

	const (
		maxPow10    = 22        // largest exact power of 10
		maxMantissa = 1<<53 + 1 // largest exact mantissa
	)
	switch xc := x.compact; {
	case !x.isCompact():
		fallthrough
	//lint:ignore ST1015 convoluted, but on purpose
	default:
		f, _ = strconv.ParseFloat(x.String(), 64)
		ok = !math.IsInf(f, 0) && !math.IsNaN(f)
	case xc == 0:
		ok = true
	case x.IsInt():
		if xc, ok := x.Int64(); ok {
			f = float64(xc)
		} else if xc, ok := x.Uint64(); ok {
			f = float64(xc)
		}
		ok = xc < maxMantissa || (xc&(xc-1)) == 0
	case x.exp == 0:
		f = float64(xc)
		ok = xc < maxMantissa || (xc&(xc-1)) == 0
	case x.exp > 0:
		f = float64(x.compact) * math.Pow10(x.exp)
		ok = x.compact < maxMantissa && x.exp < maxPow10
	case x.exp < 0:
		f = float64(x.compact) / math.Pow10(-x.exp)
		ok = x.compact < maxMantissa && x.exp > -maxPow10
	}

	if x.form&signbit != 0 {
		f = math.Copysign(f, -1)
	}
	return f, ok
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
		if x.isZero() {
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

// Format implements fmt.Formatter, recognizing the same format verbs as Append.
//
// When used in conjunction with any of the recognized verbs, Format honors all
// flags in the manner described for floating point numbers in the ``fmt''
// package.
//
// All other unrecognized format and flag combinations (such as %#v) are passed
// through to the ``fmt'' package.
func (x *Big) Format(s fmt.State, c rune) {
	if debug {
		x.validate()
	}

	prec, hasPrec := s.Precision()
	if !hasPrec {
		prec = 6
	}

	switch c {
	case 'e', 'E', 'f', 'F':
		// OK
	case 'g', 'G':
		if !hasPrec {
			prec = -1
		}
	case 'v':
		// Handle %#v as a special case.
		if !s.Flag('#') {
			c = 'g'
			break
		}
		fallthrough
	default:
		type Big struct {
			Context   Context
			unscaled  big.Int
			compact   uint64
			exp       int
			precision int
			form      form
		}
		fmt.Fprintf(s, makeFormat(s, c), (*Big)(x))
		return
	}

	cap := 10
	if prec > 0 {
		cap += prec
	}
	buf := x.Append(make([]byte, 0, cap), byte(c), prec)

	var sign string
	switch {
	case buf[0] == '-':
		sign = "-"
		buf = buf[1:]
	case buf[0] == '+':
		// +Inf
		sign = "+"
		if s.Flag(' ') {
			sign = "  "
		}
		buf = buf[1:]
	case s.Flag('+'):
		sign = "+"
	case s.Flag(' '):
		sign = " "
	}

	// Sharp flag requires a decimal point.
	if s.Flag('#') {
		digits := prec
		if digits < 0 {
			digits = 6
		}
		tail := make([]byte, 0, 6)
		hasDot := false
		for i := 0; i < len(buf); i++ {
			switch buf[i] {
			case '.':
				hasDot = true
			case 'e', 'E':
				tail = append(tail, buf[i:]...)
				buf = buf[:i]
			default:
				digits--
			}
		}
		if !hasDot {
			buf = append(buf, '.')
		}
		for ; digits > 0; digits-- {
			buf = append(buf, '0')
		}
		buf = append(buf, tail...)
	}

	var padding int
	if width, ok := s.Width(); ok && width > len(sign)+len(buf) {
		padding = width - len(sign) - len(buf)
	}

	switch {
	case s.Flag('0') && !x.IsInf(0):
		writeN(s, sign, 1)
		writeN(s, "0", padding)
		s.Write(buf)
	case s.Flag('-'):
		writeN(s, sign, 1)
		s.Write(buf)
		writeN(s, " ", padding)
	default:
		writeN(s, " ", padding)
		writeN(s, sign, 1)
		s.Write(buf)
	}
}

// makeFormat recreates a format string.
func makeFormat(s fmt.State, c rune) string {
	var b strings.Builder
	b.WriteByte('%')
	for _, c := range "+-# 0" {
		if s.Flag(int(c)) {
			b.WriteRune(c)
		}
	}
	if width, ok := s.Width(); ok {
		b.WriteString(strconv.Itoa(width))
	}
	if prec, ok := s.Precision(); ok {
		b.WriteByte('.')
		b.WriteString(strconv.Itoa(prec))
	}
	b.WriteRune(c)
	return b.String()
}

// writeN writes s to w n times, if s != "".
func writeN(w io.Writer, s string, n int) {
	if s == "" {
		return
	}
	type stringWriter interface {
		WriteString(string) (int, error)
	}
	if sw, ok := w.(stringWriter); ok {
		for ; n > 0; n-- {
			sw.WriteString(s)
		}
	} else {
		tmp := []byte(s)
		for ; n > 0; n-- {
			w.Write(tmp)
		}
	}
}

// FMA sets z to (x * y) + u without any intermediate rounding.
func (z *Big) FMA(x, y, u *Big) *Big { return z.Context.FMA(z, x, y, u) }

// Int sets z to x, truncating the fractional portion (if any) and returns z.
//
// z is allowed to be nil. If x is an infinity or a NaN value the result is
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
	return bigScalex(z, z, x.exp)
}

// Int64 returns x as an int64, truncating towards zero. The returned boolean
// indicates whether the conversion to an int64 was successful.
func (x *Big) Int64() (int64, bool) {
	if debug {
		x.validate()
	}

	if !x.IsFinite() {
		return 0, false
	}

	// x might be too large to fit into an int64 *now*, but rescaling x might
	// shrink it enough. See issue #20.
	if !x.isCompact() {
		xb := x.Int(nil)
		return xb.Int64(), xb.IsInt64()
	}

	u := x.compact
	if x.exp != 0 {
		var ok bool
		if u, ok = scalex(u, x.exp); !ok {
			return 0, false
		}
	}
	su := int64(u)
	if su >= 0 || x.Signbit() && su == -su {
		if x.Signbit() {
			su = -su
		}
		return su, true
	}
	return 0, false
}

// Uint64 returns x as a uint64, truncating towards zero. The returned boolean
// indicates whether the conversion to a uint64 was successful.
func (x *Big) Uint64() (uint64, bool) {
	if debug {
		x.validate()
	}

	if !x.IsFinite() || x.Signbit() {
		return 0, false
	}

	// x might be too large to fit into an uint64 *now*, but rescaling x might
	// shrink it enough. See issue #20.
	if !x.isCompact() {
		xb := x.Int(nil)
		return xb.Uint64(), xb.IsUint64()
	}

	b := x.compact
	if x.exp == 0 {
		return b, true
	}
	return scalex(b, x.exp)
}

// IsFinite returns true if x is finite.
func (x *Big) IsFinite() bool { return x.form & ^signbit == 0 }

// IsNormal returns true if x is normal.
func (x *Big) IsNormal() bool {
	return x.IsFinite() && x.adjusted() >= x.Context.minScale()
}

// IsSubnormal returns true if x is subnormal.
func (x *Big) IsSubnormal() bool {
	return x.IsFinite() && x.adjusted() < x.Context.minScale()
}

// IsInf returns true if x is an infinity according to sign.
//
//    If sign >  0, IsInf reports whether x is positive infinity.
//    If sign <  0, IsInf reports whether x is negative infinity.
//    If sign == 0, IsInf reports whether x is either infinity.
//
func (x *Big) IsInf(sign int) bool {
	return sign >= 0 && x.form == pinf || sign <= 0 && x.form == ninf
}

// IsNaN returns true if x is NaN.
//
//    If sign >  0, IsNaN reports whether x is quiet NaN.
//    If sign <  0, IsNaN reports whether x is signaling NaN.
//    If sign == 0, IsNaN reports whether x is either quiet or signaling NaN.
//
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

	// 0, 5000, 40
	if x.isZero() || x.exp >= 0 {
		return true
	}

	xp := x.Precision()
	exp := x.exp

	// 0.00d
	// 0.d
	if -exp >= xp {
		return false
	}

	// 44.dd
	// 1.ddd
	if x.isCompact() {
		for v := x.compact; v%10 == 0; v /= 10 {
			exp++
		}
		// Avoid the overhead of copying x.unscaled if we know for a fact it's not
		// an integer.
	} else if x.unscaled.Bit(0) == 0 {
		v := new(big.Int).Set(&x.unscaled)
		r := new(big.Int)
		for {
			v.QuoRem(v, c.TenInt, r)
			if r.Sign() != 0 {
				break
			}
			exp++
		}
	}
	return exp >= 0
}

// MarshalText implements encoding.TextMarshaler, formatting x like x.String.
func (x *Big) MarshalText() ([]byte, error) {
	if debug {
		x.validate()
	}
	return x.Append(make([]byte, 0, 10), 'g', -1), nil
}

// Mul sets z to x * y and returns z.
func (z *Big) Mul(x, y *Big) *Big { return z.Context.Mul(z, x, y) }

// Neg sets z to -x and returns z. If x is positive infinity, z will be set to
// negative infinity and visa versa. If x == 0, z will be set to zero as well.
// NaN will result in an error.
func (z *Big) Neg(x *Big) *Big {
	if debug {
		x.validate()
	}
	if !z.invalidContext(z.Context) && !z.checkNaNs(x, x, negation) {
		xform := x.form // copy in case z == x
		z.copyAbs(x)
		if !z.IsFinite() || z.compact != 0 || z.Context.RoundingMode == ToNegativeInf {
			z.form = xform ^ signbit
		}
	}
	return z.Context.round(z)
}

// New creates a new Big decimal with the given value and scale. For example:
//
//    New(1234, 3) // 1.234
//    New(42, 0)   // 42
//    New(4321, 5) // 0.04321
//    New(-1, 0)   // -1
//    New(3, -10)  // 30 000 000 000
//
func New(value int64, scale int) *Big {
	return new(Big).SetMantScale(value, scale)
}

// Payload returns the payload of x, provided x is a NaN value. If x is not a
// NaN value, the result is undefined.
func (x *Big) Payload() Payload {
	if !x.IsNaN(0) {
		return 0
	}
	return Payload(x.compact)
}

// Precision returns the precision of x. That is, it returns the number of
// digits in the unscaled form of x. x == 0 has a precision of 1. The result is
// undefined if x is not finite.
func (x *Big) Precision() int {
	// Cannot call validate since validate calls this method.
	if !x.IsFinite() {
		return 0
	}
	if x.precision == 0 {
		return 1
	}
	return x.precision
}

// Quantize sets z to the number equal in value and sign to z with the scale, n.
// The rounding of z is performed according to the rounding mode set in z.Context.RoundingMode.
// In order to perform truncation, set z.Context.RoundingMode to ToZero.
func (z *Big) Quantize(n int) *Big { return z.Context.Quantize(z, n) }

// Quo sets z to x / y and returns z.
func (z *Big) Quo(x, y *Big) *Big { return z.Context.Quo(z, x, y) }

// QuoInt sets z to x / y with the remainder truncated. See QuoRem for more
// details.
func (z *Big) QuoInt(x, y *Big) *Big { return z.Context.QuoInt(z, x, y) }

// QuoRem sets z to the quotient x / y and r to the remainder x % y, such that
// x = z * y + r, and returns the pair (z, r).
func (z *Big) QuoRem(x, y, r *Big) (*Big, *Big) {
	return z.Context.QuoRem(z, x, y, r)
}

// Rat sets z to x and returns z.
//
// z is allowed to be nil. The result is undefined if x is an infinity or NaN
// value.
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

	// Fast path for decimals <= math.MaxInt64.
	if x.IsInt() {
		if u, ok := x.Int64(); ok {
			// If profiled we can call scalex ourselves and save the overhead of
			// calling Int64. But I doubt it'll matter much.
			return z.SetInt64(u)
		}
	}

	num := new(big.Int)
	if x.isCompact() {
		num.SetUint64(x.compact)
	} else {
		num.Set(&x.unscaled)
	}
	if x.exp > 0 {
		arith.MulBigPow10(num, num, uint64(x.exp))
	}
	if x.Signbit() {
		num.Neg(num)
	}

	denom := c.OneInt
	if x.exp < 0 {
		denom = new(big.Int)
		if shift, ok := arith.Pow10(uint64(-x.exp)); ok {
			denom.SetUint64(shift)
		} else {
			denom.Set(arith.BigPow10(uint64(-x.exp)))
		}
	}
	return z.SetFrac(num, denom)
}

// Raw directly returns x's raw compact and unscaled values.
//
// Caveat emptor: Neither are guaranteed to be valid. Raw is intended to support
// missing functionality outside this package and generally should be avoided.
//
// Additionally, Raw is the only part of this package's API which is not
// guaranteed to remain stable. This means the function could change or disappear
// without warning at any time, even across minor versions.
func Raw(x *Big) (*uint64, *big.Int) { return &x.compact, &x.unscaled }

// Reduce reduces a finite z to its most simplest form.
func (z *Big) Reduce() *Big { return z.Context.Reduce(z) }

// Rem sets z to the remainder x % y. See QuoRem for more details.
func (z *Big) Rem(x, y *Big) *Big { return z.Context.Rem(z, x, y) }

// Round rounds z down to n digits of precision and returns z. The result is
// undefined if z is not finite. No rounding will occur if n <= 0. The result of
// Round will always be within the interval [⌊10**x⌋, z] where x = the precision
// of z.
func (z *Big) Round(n int) *Big {
	ctx := z.Context
	ctx.Precision = n
	return ctx.Round(z)
}

// RoundToInt rounds z down to an integral value.
func (z *Big) RoundToInt() *Big { return z.Context.RoundToInt(z) }

// Scale returns x's scale.
func (x *Big) Scale() int { return -x.exp }

// Scan implements fmt.Scanner.
func (z *Big) Scan(state fmt.ScanState, verb rune) error {
	return z.scan(byteReader{state})
}

// Set sets z to x and returns z. The result might be rounded depending on z's
// Context, and even if z == x.
func (z *Big) Set(x *Big) *Big { return z.Context.round(z.Copy(x)) }

// setShared sets z to x, but does not copy—z may possibly alias x.
func (z *Big) setShared(x *Big) *Big {
	if debug {
		x.validate()
	}

	if z != x {
		z.precision = x.Precision()
		z.compact = x.compact
		z.form = x.form
		z.exp = x.exp
		z.unscaled = x.unscaled
	}
	return z
}

// SetBigMantScale sets z to the given value and scale.
func (z *Big) SetBigMantScale(value *big.Int, scale int) *Big {
	// Do this first in case value == z.unscaled. Don't want to clobber the sign.
	z.form = finite
	if value.Sign() < 0 {
		z.form |= signbit
	}

	z.unscaled.Abs(value)
	z.compact = c.Inflated
	z.precision = arith.BigLength(value)

	if z.unscaled.IsUint64() {
		if v := z.unscaled.Uint64(); v != c.Inflated {
			z.compact = v
		}
	}

	z.exp = -scale
	return z
}

// SetFloat sets z to exactly x and returns z.
func (z *Big) SetFloat(x *big.Float) *Big {
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
		if neg {
			z.form |= signbit
		}
		z.compact = 0
		z.precision = 1
		return z
	}

	z.exp = 0
	x0 := new(big.Float).Copy(x).SetPrec(big.MaxPrec)
	x0.Abs(x0)
	if !x.IsInt() {
		for !x0.IsInt() {
			x0.Mul(x0, c.TenFloat)
			z.exp--
		}
	}

	if mant, acc := x0.Uint64(); acc == big.Exact {
		z.compact = mant
		z.precision = arith.Length(mant)
	} else {
		z.compact = c.Inflated
		x0.Int(&z.unscaled)
		z.precision = arith.BigLength(&z.unscaled)
	}
	z.form = finite
	if neg {
		z.form |= signbit
	}
	return z
}

// SetFloat64 sets z to exactly x.
func (z *Big) SetFloat64(x float64) *Big {
	if x == 0 {
		var sign form
		if math.Signbit(x) {
			sign = signbit
		}
		return z.setZero(sign, 0)
	}
	if math.IsNaN(x) {
		var sign form
		if math.Signbit(x) {
			sign = signbit
		}
		return z.setNaN(0, qnan|sign, 0)
	}
	if math.IsInf(x, 0) {
		if math.IsInf(x, 1) {
			z.form = pinf
		} else {
			z.form = ninf
		}
		return z
	}

	// The gist of the following is lifted from math/big/rat.go, but adapted for
	// base-10 decimals.

	const expMask = 1<<11 - 1
	bits := math.Float64bits(x)
	mantissa := bits & (1<<52 - 1)
	exp := int((bits >> 52) & expMask)
	if exp == 0 { // denormal
		exp -= 1022
	} else { // normal
		mantissa |= 1 << 52
		exp -= 1023
	}

	if mantissa == 0 {
		return z.SetUint64(0)
	}

	shift := 52 - exp
	for mantissa&1 == 0 && shift > 0 {
		mantissa >>= 1
		shift--
	}

	z.exp = 0
	z.form = finite | form(bits>>63)

	if shift > 0 {
		z.unscaled.SetUint64(uint64(shift))
		z.unscaled.Exp(c.FiveInt, &z.unscaled, nil)
		arith.Mul(&z.unscaled, &z.unscaled, mantissa)
		z.exp = -shift
	} else {
		// TODO(eric): figure out why this doesn't work for _some_ numbers. See
		// https://github.com/ericlagergren/decimal/issues/89
		//
		// z.compact = mantissa << uint(-shift)
		// z.precision = arith.Length(z.compact)

		z.compact = c.Inflated
		z.unscaled.SetUint64(mantissa)
		z.unscaled.Lsh(&z.unscaled, uint(-shift))
	}
	return z.norm()
}

// SetInf sets z to -Inf if signbit is set or +Inf is signbit is not set, and
// returns z.
func (z *Big) SetInf(signbit bool) *Big {
	if signbit {
		z.form = ninf
	} else {
		z.form = pinf
	}
	return z
}

// SetMantScale sets z to the given value and scale.
func (z *Big) SetMantScale(value int64, scale int) *Big {
	z.SetUint64(arith.Abs(value))
	z.exp = -scale // compiler should optimize out z.exp = 0 in SetUint64
	if value < 0 {
		z.form |= signbit
	}
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
// returns z. No conditions are raised.
func (z *Big) SetNaN(signal bool) *Big {
	if signal {
		z.form = snan
	} else {
		z.form = qnan
	}
	z.compact = 0 // payload
	return z
}

// SetRat sets z to to the possibly rounded value of x and return z.
func (z *Big) SetRat(x *big.Rat) *Big {
	if x.IsInt() {
		return z.Context.round(z.SetBigMantScale(x.Num(), 0))
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

// Regexp matches any valid string representing a decimal that can be passed to
// SetString.
var Regexp = regexp.MustCompile(`(?i)(([+-]?(\d+\.\d*|\.?\d+)([eE][+-]?\d+)?)|(inf(infinity)?))|([+-]?([sq]?nan\d*))`)

// SetString sets z to the value of s, returning z and a bool indicating success.
// s must be a string in one of the following formats:
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

func (z *Big) setTriple(compact uint64, sign form, exp int) *Big {
	z.compact = compact
	z.precision = arith.Length(compact)
	z.exp = exp
	z.form = finite | sign
	return z
}

func (z *Big) setZero(sign form, exp int) *Big {
	z.compact = 0
	z.precision = 1
	z.exp = exp
	z.form = finite | sign
	return z
}

// SetUint64 is shorthand for SetMantScale(x, 0) for an unsigned integer.
func (z *Big) SetUint64(x uint64) *Big {
	z.compact = x
	if x == c.Inflated {
		z.unscaled.SetUint64(x)
	}
	z.precision = arith.Length(x)
	z.exp = 0
	z.form = finite
	return z
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
//    -1 if x <  0
//     0 if x == 0
//    +1 if x >  0
//
// No distinction is made between +0 and -0. The result is undefined if x is a
// NaN value.
func (x *Big) Sign() int {
	if debug {
		x.validate()
	}

	if (x.IsFinite() && x.isZero()) || x.IsNaN(0) {
		return 0
	}
	if x.form&signbit != 0 {
		return -1
	}
	return 1
}

// Signbit reports whether x is negative, negative zero, negative infinity, or
// negative NaN.
func (x *Big) Signbit() bool {
	if debug {
		x.validate()
	}
	return x.form&signbit != 0
}

// String formats x like x.Text('g', -1).
func (x *Big) String() string {
	return string(x.Text('g', -1))
}

// Sub sets z to x - y and returns z.
func (z *Big) Sub(x, y *Big) *Big { return z.Context.Sub(z, x, y) }

// Text appends x to buf using the specified format and returns the resulting
// slice.
//
// The following verbs are recognized:
//
//    'e': -d.dddd±edd
//    'E': -d.dddd±Edd
//    'f': -dddd.dd
//    'F': same as 'f'
//    'g': same as 'f' or 'e', depending on x
//    'G': same as 'F' or 'E', depending on x
//
// For every format the precision is the number of digits following the radix,
// except in the case of 'g' and 'G' where the precision is the number of
// significant digits.
func (x *Big) Text(fmt byte, prec int) string {
	return string(x.Append(make([]byte, 0, 10), fmt, prec))
}

// UnmarshalJSON implements json.Unmarshaler.
func (z *Big) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}
	return z.UnmarshalText(data)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (z *Big) UnmarshalText(data []byte) error {
	return z.scan(bytes.NewReader(data))
}

// validate ensures x's internal state is correct. There's no need for it to
// have good performance since it's for debug == true only.
func (x *Big) validate() {
	defer func() {
		if err := recover(); err != nil {
			pc, _, _, ok := runtime.Caller(4)
			if caller := runtime.FuncForPC(pc); ok && caller != nil {
				fmt.Println("called by:", caller.Name())
			}
			println(x)
			panic(err)
		}
	}()
	switch x.form {
	case finite, finite | signbit:
		if x.isInflated() {
			if x.unscaled.IsUint64() && x.unscaled.Uint64() != c.Inflated {
				panic(fmt.Sprintf("inflated but unscaled == %d", x.unscaled.Uint64()))
			}
			if x.unscaled.Sign() < 0 {
				panic("x.unscaled.Sign() < 0")
			}
			if bl, xp := arith.BigLength(&x.unscaled), x.precision; bl != xp {
				panic(fmt.Sprintf("BigLength (%d) != x.Precision (%d)", bl, xp))
			}
		}
		if x.isCompact() {
			if bl, xp := arith.Length(x.compact), x.Precision(); bl != xp {
				panic(fmt.Sprintf("BigLength (%d) != x.Precision() (%d)", bl, xp))
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
