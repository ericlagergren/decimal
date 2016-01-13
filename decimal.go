package decimal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Precision and scale limits.
const (
	MinScale = math.MinInt64 + 1 // smallest allowed scale.
	MaxScale = math.MaxInt64     // largest allowed scale.
	MinPrec  = 0                 // smallest allowed precision.
	MaxPrec  = math.MaxInt64     // maximum allowed precision.
)

// Constant values for ease of use.
var (
	Zero = New(0, 0)
	One  = New(1, 0)
	Five = New(5, 0)
	Ten  = New(10, 0)
)

// Decimal should be packed.
// Currently on (64-bit architectures):
// 		compact:  8
// 		scale:    4
// 		ctx:      4 + 1  (5)
// 		mantissa: 1 + 24 (4)
// 		----------------
// 		total:    8 + 4 + 5 + 4 = 48 bytes.

// Decimal represents a multi-precision, fixed-point
// decimal number.
//
// Decimal = v * (10 ^ scale)
//
// A positive scale represents the number of digits to the right
// of the radix. A negative scale represents the value of the
// Decimal (at this point an integer) multiplied by the negated
// value of the scale.
//
// A Decimal's precision represents the number of decimal digits
// that should follow the radix after a lossy arithmetic operation.
type Decimal struct {
	// If |mantissa| <= math.MaxInt64 then the mantissa
	// will be stored in this field.
	compact  int64
	scale    int64
	ctx      Context
	mantissa big.Int
}

func NaN() *Decimal {
	return &Decimal{scale: overflown}
}

// New returns a new fixed-point decimal.
func New(value int64, scale int64) *Decimal {
	if scale == overflown {
		panic("decimal.New: scale is too small")
	}
	return &Decimal{
		compact: value,
		scale:   scale,
	}
}

// NewFromString returns a new Decimal from a string representation.
//
// Example:
//
//     d1, err := NewFromString("-123.45")
//     d2, err := NewFromString(".0001")
//
func NewFromString(value string) (*Decimal, error) {
	originalInput := value

	var exp int64

	// Check if number is using scientific notation.
	eIndex := strings.IndexAny(value, "Ee")
	if eIndex != -1 {
		expInt, err := strconv.ParseInt(value[eIndex+1:], 10, 64)
		if expInt == overflown {
			panic("decimal.NewFromString: scale is too small")
		}
		if err != nil {
			if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
				return nil, fmt.Errorf("decimal.NewFromString: can't convert %s to decimal: fractional part too long", value)
			}
			return nil, fmt.Errorf("decimal.NewFromString: can't convert %s to decimal: exponent is not numeric", value)
		}
		value = value[:eIndex]
		exp = int64(-expInt)
	}

	var intString string

	switch parts := strings.Split(value, "."); len(parts) {
	case 1:
		// There is no decimal point, we can just parse the original string as
		// a whole integer
		intString = value
	case 2:
		intString = parts[0] + parts[1]
		exp += int64(len(parts[1]))
	default:
		return nil, fmt.Errorf("decimal.NewFromString: can't convert %s to decimal: too many .s", value)
	}

	dValue := new(big.Int)
	_, ok := dValue.SetString(intString, 10)
	if !ok {
		return nil, fmt.Errorf("decimal.NewFromString: can't convert %s to decimal", value)
	}

	if exp < math.MinInt64 || exp > math.MaxInt64 {
		// NOTE(vadim): I doubt a string could realistically be this long
		return nil, fmt.Errorf("decimal.NewFromString: can't convert %s to decimal: fractional part too long", originalInput)
	}

	// Determine if we can fit the value into an int64.
	// We stuff the value into a big.Int first since it's
	// guaranteed to fit and best-case we only parse the
	// string once.
	//
	// TODO(eric): Heuristics to check whether or not we can
	// fit the input inside an int64 before we stuff it into
	// a big.Int.

	z := Decimal{
		compact:  overflown,
		scale:    int64(exp),
		mantissa: *dValue,
	}
	return z.Shrink(), nil
}

// NewFromFloat converts a float64 to Decimal.
//
// Keep in mind that float -> decimal conversions can be lossy.
// For example, 0.1 appears to be "just" 0.1, but in reality it's
// 0.1000000000000000055511151231257827021181583404541015625
// (see: fmt.Printf("%.55f", 0.1))
//
// In order to cope with this, the number of decimal digits in the float
// are calculated as closely as possible use that as the scale.
//
// Approximately 2.3% of decimals created from floats will have a rounding
// imprecision of ± 1 ULP.
//
// Example:
//
//     NewFromFloat(123.45678901234567).String() // output: "123.4567890123456"
//     NewFromFloat(.00000000000000001).String() // output: "0.00000000000000001"
//
// NOTE: this will panic on NaN, +/-inf
func NewFromFloat(value float64) *Decimal {
	return NewFromFloatWithScale(value, prec(value))
}

// NewFromFloatWithScale converts a float64 to Decimal, with an arbitrary
// number of fractional digits.
//
// Example:
//
//     NewFromFloatWithScale(123.456, 2).String() // output: "123.46"
//
func NewFromFloatWithScale(value float64, scale int64) *Decimal {
	if scale == overflown {
		panic("decimal: scale is too small")
	}

	value *= pow10(scale)
	if math.IsNaN(value) || math.IsInf(value, 0) {
		panic(fmt.Sprintf("decimal: cannot create a Decimal from %v", value))
	}

	z := Decimal{scale: scale}

	// Given float64(math.MaxInt64) == math.MaxInt64
	if value <= math.MaxInt64 {
		// TODO(eric):
		// Should we put an integer that's so close to overflowing inside
		// the compact member?
		z.compact = int64(value)
	} else {
		// Given float64(math.MaxUint64) == math.MaxUint64
		if value <= math.MaxUint64 {
			z.mantissa.SetUint64(uint64(value))
		} else {
			z.mantissa.Set(bigIntFromFloat(value))
		}
		z.compact = overflown
	}
	return &z
}

// "stolen" from https://golang.org/pkg/math/big/#Rat.SetFloat64
// Removed non-finite case because we already check for
// Inf/NaN values
func bigIntFromFloat(f float64) *big.Int {
	const expMask = 1<<11 - 1
	bits := math.Float64bits(f)
	mantissa := bits & (1<<52 - 1)
	exp := int((bits >> 52) & expMask)
	if exp == 0 { // denormal
		exp -= 1022
	} else { // normal
		mantissa |= 1 << 52
		exp -= 1023
	}

	shift := 52 - exp

	// Optimization (?): partially pre-normalise.
	for mantissa&1 == 0 && shift > 0 {
		mantissa >>= 1
		shift--
	}

	if shift < 0 {
		shift = -shift
	}

	var a big.Int
	a.SetUint64(mantissa)
	return a.Lsh(&a, uint(shift))
}

// prec determines the precision of a float64.
func prec(f float64) (precision int64) {
	if math.IsNaN(f) ||
		math.IsInf(f, 0) ||
		math.Floor(f) == f {
		return 0
	}

	e := float64(1)
	for cmp := round(f*e) / e; !math.IsNaN(cmp) && cmp != f; cmp = round(f*e) / e {
		e *= 10
	}
	return int64(math.Ceil(math.Log10(e)))
}

// Abs sets z to the absolute value of x and returns z.
func (z *Decimal) Abs(x *Decimal) *Decimal {
	if x.compact != overflown {
		z.compact = abs(x.compact)
	} else {
		z.mantissa.Abs(&x.mantissa)
		z.compact = overflown
	}
	z.scale = x.scale
	return z
}

// Add sets z to x + y and returns z.
func (z *Decimal) Add(x, y *Decimal) *Decimal {
	// The Mul method follows the same steps as Adz, so I'll detail the
	// formula in the various add methods.
	if x.compact != overflown {
		if y.compact != overflown {
			return z.addCompact(x, y)
		}
		return z.addHalf(x, y)
	}
	if y.compact != overflown {
		return z.addHalf(y, x)
	}
	return z.addBig(x, y)
}

// addCompact set d to the sum of x and y and returns z.
// Each case depends on the scales.
func (z *Decimal) addCompact(x, y *Decimal) *Decimal {
	// Fast path: we don't need to adjust anything.
	// Just check for overflows (if so, use a big.Int)
	// and return the result.
	if x.scale == y.scale {
		z.scale = x.scale
		sum := sum(x.compact, y.compact)
		if sum != overflown {
			z.compact = sum
		} else {
			z.mantissa.Add(big.NewInt(x.compact), big.NewInt(y.compact))
			z.compact = overflown
		}
		return z
	}

	// Guess the high and low scale. If we guess wrong, swap.
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	// Find which power of 10 we have to multiply our low value by in order
	// to equalize their scales.
	inc := safeScale(lo.compact, hi.scale, sub(hi.scale, lo.scale))

	z.scale = hi.scale

	// Expand the low value (checking for overflows) and
	// find the sum (checking for overflows).
	//
	// If we overflow at all use a big.Int to calculate the sum.
	scaledLo := mulPow10(lo.compact, inc)
	if scaledLo != overflown {
		sum := sum(hi.compact, scaledLo)
		if sum != overflown {
			z.compact = sum
			return z
		}
	}

	scaled := mulBigPow10(big.NewInt(lo.compact), inc)
	z.mantissa.Add(scaled, big.NewInt(hi.compact))
	z.compact = overflown
	return z
}

// addHalf adds a compact Decimal with a non-compact
// Decimal.
// Let the first arg be the compact and the second the non-compact.
func (z *Decimal) addHalf(comp, nc *Decimal) *Decimal {
	if comp.compact == overflown {
		panic("decimal.Add: (bug) comp should != overflown")
	}
	if comp.scale == nc.scale {
		z.mantissa.Add(big.NewInt(comp.compact), &nc.mantissa)
		z.scale = comp.scale
		z.compact = overflown
		return z
	}
	// Since we have to rescale we need to add two big.Ints
	// together because big.Int doesn't have an API for
	// increasing its value by an integer.
	return z.addBig(&Decimal{
		mantissa: *big.NewInt(comp.compact),
		scale:    comp.scale,
	}, nc)
}

func (z *Decimal) addBig(x, y *Decimal) *Decimal {
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	inc := safeScale(lo.compact, hi.scale, sub(hi.scale, lo.scale))
	scaled := mulBigPow10(&lo.mantissa, inc)
	z.mantissa.Add(&hi.mantissa, scaled)
	z.compact = overflown
	z.scale = hi.scale
	return z
}

// and sets z to to x & n and returns z.
func (z *Decimal) and(x *Decimal, n int64) *Decimal {
	if x.compact != overflown {
		z.compact = x.compact & n
	} else {
		z.mantissa.And(&x.mantissa, big.NewInt(n))
		z.compact = overflown
	}
	return z
}

// Binomial sets z to the binomial coefficient of (n, k) and returns z.
func (z *Decimal) Binomial(n, k int64) *Decimal {
	if n/2 < k && k <= n {
		k = n - k
	}
	var a, b Decimal
	a.MulRange(n-k+1, n)
	b.MulRange(1, k)
	return z.Div(&a, &b)
}

// BitLen returns the absolute value of z in bits.
func (z *Decimal) BitLen() int64 {
	if z.compact != overflown {
		x := z.compact
		if z.scale < 0 {
			x = mulPow10(x, -z.scale)
		}
		if x != overflown {
			return (64 - clz(x))
		}
	}
	x := &z.mantissa
	if z.scale < 0 {
		// Double check because we fall through if
		// mulPow10(x, -z.scale) returns overflown.
		if z.compact != overflown {
			x = mulBigPow10(big.NewInt(z.compact), -z.scale)
		} else {
			x = mulBigPow10(x, -z.scale)
		}
	}
	return int64(x.BitLen())
}

// Bytes returns the absolute value of z as a big-endian
// byte slice.
func (z *Decimal) Bytes() []byte {
	if z.compact != overflown {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(abs(z.compact)))
		return b[:]
	}
	return z.mantissa.Bytes()
}

// Ceil sets z to the nearest integer value greater than or equal to x
// and returns z.
func (z *Decimal) Ceil(x *Decimal) *Decimal {
	z.Floor(z.Neg(x))
	return z.Neg(z)
}

// Cmp compares d and x and returns:
//
//   -1 if z <  x
//    0 if z == x
//   +1 if z >  x
//
// It does not modify d or x.
func (z *Decimal) Cmp(x *Decimal) int {
	// Check for same pointers.
	if z == x {
		return 0
	}

	// Same scales means we can compare straight across.
	if z.scale == x.scale &&
		z.compact != overflown && x.compact != overflown {
		if z.compact > x.compact {
			return +1
		}
		if z.compact < x.compact {
			return -1
		}
		return 0
	}

	// Different scales -- check signs and/or if they're
	// both zero.

	ds := z.Sign()
	xs := x.Sign()
	switch {
	case ds > xs:
		return +1
	case ds < xs:
		return -1
	case ds == 0 && xs == 0:
		return 0
	}

	// Scales aren't equal, the signs are the same, and both
	// are non-zero.
	dl := z.Ilog10() - z.scale
	xl := x.Ilog10() - x.scale
	if dl > xl {
		return +1
	}
	if dl < xl {
		return -1
	}

	// We need to inflate one of the numbers.

	dc := z.compact // hi
	xc := x.compact // lo

	var swap bool

	hi, lo := z, x
	if hi.scale < lo.scale {
		hi, lo = lo, hi
		dc, xc = xc, dc
		swap = true // d is lo
	}

	diff := hi.scale - lo.scale
	if diff <= math.MaxInt64 {
		xc = mulPow10(xc, diff)
		if xc == overflown && dc == overflown {
			// d is lo
			if swap {
				return mulBigPow10(&z.mantissa, diff).
					Cmp(&x.mantissa)
			}
			// x is lo
			return z.mantissa.Cmp(mulBigPow10(&x.mantissa, diff))
		}
	}

	if swap {
		dc, xc = xc, dc
	}

	if dc != overflown {
		if xc != overflown {
			return cmpAbs(dc, xc)
		}
		return big.NewInt(dc).Cmp(&x.mantissa)
	}
	if xc != overflown {
		return z.mantissa.Cmp(big.NewInt(xc))
	}
	return z.mantissa.Cmp(&x.mantissa)
}

// Dim sets z to the maximum of x - y or 0 and returns z.
func (z *Decimal) Dim(x, y *Decimal) *Decimal {
	x0 := new(Decimal).Sub(x, y)
	return Max(x0, New(0, 0))
}

// DivMod sets z to the quotient x div y and m to the modulus x mod y and
// returns the pair (z, m) for y != 0. If y == 0, a division-by-zero run-time panic occurs
func (z *Decimal) DivMod(x, y, m *Decimal) (div *Decimal, moz *Decimal) {
	if y.ez() {
		panic("decimal.DivMod: division by zero")
	}

	if x.ez() {
		z.compact = 0
		z.scale = safeScale2(x.scale, sub(x.scale, y.scale))
		return z, m.SetInt64(0)
	}

	if x.compact != overflown {
		if y.compact != overflown {
			if m.compact != overflown {
				return z.divCompact(x, y, m)
			}
			return z.divBig(x, y, &Decimal{
				mantissa: *big.NewInt(m.compact),
				scale:    m.scale,
			})
		}
		return z.divBig(&Decimal{
			mantissa: *big.NewInt(x.compact),
			scale:    x.scale,
		}, y, m)
	}
	if y.compact != overflown {
		return z.divBig(x, &Decimal{
			mantissa: *big.NewInt(y.compact),
			scale:    y.scale,
		}, m)
	}
	return z.divBig(x, y, m)
}

// Div sets z to the quotient x/y for y != 0 and returns z. If y == 0, a
// division-by-zero run-time panic occurs.
func (z *Decimal) Div(x, y *Decimal) *Decimal {
	var r Decimal
	div, _ := z.DivMod(x, y, &r)
	return div
}

func (z *Decimal) needsInc(x, r int64, pos, odd bool) bool {
	m := 1
	if r > math.MinInt64/2 || r <= math.MaxInt64/2 {
		m = cmpAbs(r<<1, x)
	}
	return z.ctx.Mode.needsInc(m, pos, odd)
}

func (z *Decimal) needsIncBig(x, r *big.Int, pos, odd bool) bool {
	var x0 big.Int
	m := cmpBigAbs(*x0.Mul(r, twoInt), *x)
	return z.ctx.Mode.needsInc(m, pos, odd)
}

func (z *Decimal) divCompact(x, y, m *Decimal) (div *Decimal, moz *Decimal) {

	shift := z.Prec()

	// Shifts >= 19 are guaranteed to overflow.
	if shift < 19 {
		// We're still not guaranteed to not overflow.
		x0 := prod(x.compact, pow10int64(shift))
		if x0 != overflown {
			q := x0 / y.compact
			r := x0 % y.compact
			sign := int64(1)
			if (x.compact < 0) != (y.compact < 0) {
				sign = -1
			}
			z.compact = q
			if r != 0 && z.needsInc(y.compact, r, sign > 0, q&1 != 0) {
				z.compact += sign
			}
			z.scale = safeScale2(x.scale, x.scale-y.scale+shift)
			return z.SetPrec(shift), m.SetInt64(r)
		}
	}
	return z.divBig(
		&Decimal{
			mantissa: *big.NewInt(x.compact),
			scale:    x.scale,
		},
		&Decimal{
			mantissa: *big.NewInt(y.compact),
			scale:    y.scale,
		},
		&Decimal{
			mantissa: *big.NewInt(m.compact),
			scale:    m.scale,
		},
	)
}

func (z *Decimal) divBig(x, y, m *Decimal) (div *Decimal, moz *Decimal) {

	shift := z.Prec()

	x0 := mulBigPow10(&x.mantissa, shift)
	q, r := x0.DivMod(x0, &y.mantissa, &m.mantissa)

	sign := int64(1)
	if (x.mantissa.Sign() < 0) && (y.mantissa.Sign() < 0) {
		sign = -1
	}

	z.mantissa = *q
	m.mantissa = *r

	odd := new(big.Int).And(q, oneInt).Cmp(zeroInt) != 0
	if r.Cmp(zeroInt) != 0 && z.needsIncBig(&y.mantissa, r, sign > 0, odd) {
		z.mantissa.Add(&z.mantissa, big.NewInt(sign))
	}

	z.scale = safeScale2(x.scale, x.scale-y.scale+shift)

	z.compact = overflown
	m.compact = overflown

	// I'm only comfortable calling shrink here because division
	// has a tendency to blow up numbers real big and then
	// shrink them back down.
	return z.Shrink().SetPrec(shift), m.Shrink()
}

// Equals returns true if z == x.
func (z *Decimal) Equals(x *Decimal) bool {
	return z.Cmp(x) == 0
}

// The following are some internal optimizations when we need to compare a
// Decimal to zero since d's comparison methods aren't optimized for 'zero'.

// ez returns true if z == 0.
func (z *Decimal) ez() bool {
	return z.Sign() == 0
}

// ltz returns true if z < 0
func (z *Decimal) ltz() bool {
	return z.Sign() < 0
}

// ltez returns true if z <= 0
func (z *Decimal) ltez() bool {
	return z.Sign() <= 0
}

// gtz returns true if z > 0
func (z *Decimal) gtz() bool {
	return z.Sign() > 0
}

// gtez returns true if z >= 0
func (z *Decimal) gtez() bool {
	return z.Sign() >= 0
}

// Exp sets z to x**y mod |m| and returns z.
//
// If m == nil or m == 0, z == z**y.
// If y <= the result is 1 mod |m|.
//
// Special cases are (in order):
//
// Exp(x, ±0) = 1 for any x
// Exp(1, y) = 1 for any y
// Exp(x, 1) = x for any x
// Exp(NaN, y) = NaN
// Exp(x, NaN) = NaN
// Exp(±0, y) = ±Inf for y an odd integer < 0
// Exp(±0, -Inf) = +Inf
// Exp(±0, +Inf) = +0
// Exp(±0, y) = +Inf for finite y < 0 and not an odd integer
// Exp(±0, y) = ±0 for y an odd integer > 0
// Exp(±0, y) = +0 for finite y > 0 and not an odd integer
// Exp(-1, ±Inf) = 1
// Exp(x, +Inf) = +Inf for |x| > 1
// Exp(x, -Inf) = +0 for |x| > 1
// Exp(x, +Inf) = +0 for |x| < 1
// Exp(x, -Inf) = +Inf for |x| < 1
// Exp(+Inf, y) = +Inf for y > 0
// Exp(+Inf, y) = +0 for y < 0
// Exp(-Inf, y) = Exp(-0, -y)
// Exp(x, y) = NaN for finite x < 0 and finite non-integer y
func (z *Decimal) Exp(x, y, m *Decimal) *Decimal {
	if m == nil {
		m = new(Decimal)
	}

	switch {
	case y.ez() || x.Equals(one):
		return z.SetInt64(1)
	case y.Equals(one):
		return z.Set(x)
	case y.Equals(ptFive) && m.ez():
		return z.Sqrt(x)
	}

	xbig := x.compact == overflown
	ybig := y.compact == overflown
	mbig := m.compact == overflown

	if !(xbig || ybig || mbig) {
		scale := prod(x.scale, y.compact)
		if scale == overflown {
			return NaN()
		}

		// If y is an int compute it by squaring (O log n).
		// Otherwise, use exp(log(x) * y).
		if y.IsInt() {
			z.pow(x, y)
		} else {
			x0 := new(Decimal).Set(x)
			neg := x0.ltz()
			if neg {
				x0.Neg(x0)
			}
			x0.Log(x0).Mul(x0, y)
			z.exp(x0)
			if neg {
				z.Neg(z)
			}
		}

		if !m.ez() {
			m.Mod(z, m)
		}
		return z
	}
	// Slow path. Should optimize this.

	x0 := &x.mantissa
	y0 := &y.mantissa
	m0 := &m.mantissa

	// If we have any compact Decimals assign those values.
	if !xbig {
		x0 = new(big.Int).SetInt64(x.compact)
	}
	if !ybig {
		y0 = new(big.Int).SetInt64(y.compact)
	}
	if !mbig {
		m0 = new(big.Int).SetInt64(m.compact)
	}

	// If y can't fit into an int64 then we'll overflow.
	if y0.Cmp(maxInt64) > 0 {
		return NaN()
	}
	// y <= 9223372036854775807

	intY := y0.Int64()

	// Check for scale overflow.
	scale := prod(x.scale, intY)
	if scale == overflown {
		return nil
	}
	z.mantissa.Exp(x0, y0, m0)
	z.scale = scale
	z.compact = overflown
	return z
}

// FizzBuzz literally prints out FizzBuzz from 1 up to x.
// No, really. Try it. It works.
// Yes, this is a completely useless function.
func FizzBuzz(x *Decimal) {
	var tmp Decimal
	fifteen, five, three := New(15, 0), New(5, 0), New(3, 0)
	for x0 := New(1, 0); x0.LessThan(x); x0.Add(x0, one) {
		switch {
		case tmp.Rem(x0, fifteen).ez():
			fmt.Println("FizzBuzz")
		case tmp.Rem(x0, five).ez():
			fmt.Println("Fizz")
		case tmp.Rem(x0, three).ez():
			fmt.Println("Buzz")
		default:
			fmt.Println(x0)
		}
	}
}

// Fib sets z to the Fibonacci number x and returns z.
func (z *Decimal) Fib(x *Decimal) *Decimal {
	if x.LessThan(two) {
		return z.Set(x)
	}
	a, b, x0 := New(0, 0), New(1, 0), new(Decimal).Set(x)
	for x0.Sub(x0, one); x0.gtz(); x0.Sub(x0, one) {
		a.Add(a, b)
		a, b = b, a
	}
	*z = *b
	return z
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
// For more details, see the documentation for big.Rat.Float64
func (z *Decimal) Float64() (f float64, exact bool) {
	return z.Rat(nil).Float64()
}

// Floor sets z to the nearest integer value less than or equal to x
// and returns z.
func (z *Decimal) Floor(x *Decimal) *Decimal {
	if x.compact != overflown {
		if x.compact == 0 {
			z.compact = 0
			z.scale = 0
			return z
		}

		if x.compact < 0 {
			dec, frac := modi(-x.compact, x.scale)
			if frac != 0 {
				dec++
			}
			if dec-1 != overflown {
				z.compact = -dec
				z.scale = 0
				return z
			}
		} else {
			dec, _ := modi(x.compact, x.scale)
			if dec != overflown {
				z.compact = dec
				z.scale = 0
				return z
			}
		}

		// If we reach here then we can't find the floor without using
		// big.Ints to do the math for us.
		d0 := new(Decimal).Set(x)
		d0.mantissa = *big.NewInt(x.compact)
		d0.compact = overflown
		return z.Floor(d0)
	}

	if cmp := x.mantissa.Cmp(zeroInt); cmp == 0 {
		z.mantissa.Set(zeroInt)
	} else if cmp < 0 {
		neg := new(big.Int).Neg(&x.mantissa)
		dec, frac := modbig(neg, x.scale)
		if frac.Cmp(zeroInt) != 0 {
			dec.Add(dec, oneInt)
		}
		z.mantissa.Set(dec.Neg(dec))
	} else {
		dec, _ := modbig(&x.mantissa, x.scale)
		z.mantissa.Set(dec)
	}
	z.compact = overflown
	z.scale = 0
	return z
}

// GreaterThan returns true if d is greater than x.
func (z *Decimal) GreaterThan(x *Decimal) bool {
	return z.Cmp(x) > 0
}

// GreaterThanEQ returns true if d is greater than or equal to x.
func (z *Decimal) GreaterThanEQ(x *Decimal) bool {
	return z.Cmp(x) >= 0
}

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func (z *Decimal) Hypot(p, q *Decimal) *Decimal {
	p0 := new(Decimal).Set(p)
	q0 := new(Decimal).Set(q)

	if p0.ltz() {
		p0.Neg(p0)
	}

	if q0.ltz() {
		q0.Neg(q0)
	}

	if p0.ez() {
		return New(0, 0)
	}
	p0.Mul(p0, p0)
	q0.Mul(q0, q0)
	z.Sqrt(p0.Add(p0, q0))
	return z.SetPrec(z.Prec())
}

// Int returns the integer component of z as a big.Int.
func (z *Decimal) Int() *big.Int {
	var x big.Int

	if z.scale <= 0 {
		if z.compact != overflown {
			x.SetInt64(z.compact)
		} else {
			x = z.mantissa
		}
		return mulBigPow10(&x, -z.scale)
	}

	if z.compact != overflown {
		x.SetInt64(z.compact)
	} else {
		x.Set(&z.mantissa)
	}
	b := bigPow10(z.scale)
	return x.Div(&x, &b)
}

// Int64 returns the integer component of z as an int64.
// If the integer component cannot fit into an int64 the result is undefined.
func (z *Decimal) Int64() int64 {
	var x int64
	if z.compact != overflown {
		x = z.compact
	} else {
		x = z.mantissa.Int64()
	}

	if z.scale <= 0 {
		return mulPow10(x, -z.scale)
	}

	pow := pow10int64(z.scale)
	if pow == overflown {
		return overflown
	}
	return x / pow
}

// IsBig returns true if d is too large to fit into its
// integer member and can only be represented by a big.Int.
// (This means that a call to Int will result in undefined
// behavior.)
func (z *Decimal) IsBig() bool {
	return z.compact == overflown
}

// IsInt returns true if d can be represented exactly as an integer.
// (This means that the fractional part of the number is zero or the scale
// is <= 0.)
func (z *Decimal) IsInt() bool {
	if z.scale <= 0 {
		return true
	}
	_, frac := Modf(z)
	return frac.ez()
}

func IsNan(x *Decimal) bool {
	return x.scale == overflown
}

// Jacobi returns the Jacobi symbol (x/y), either +1, -1, or 0.
// Both x and y arguments must be integers (i.e., scale <= zero).
// The y argument must be an odd integer.
func Jacobi(x, y *Decimal) int {
	if x.scale > 0 || y.scale > 0 {
		panic("decimal: invalid arguments to decimal.Jacobi: Jacobi requires integer values")
	}
	if x.compact != overflown {
		if y.compact != overflown {
			return jacobiCompact(x, y)
		}
		return jacobiHalf(x, y)
	}
	if y.compact != overflown {
		return jacobiHalf(y, x)
	}
	return big.Jacobi(&x.mantissa, &y.mantissa)
}

// formula stolen from
// https://golang.org/src/math/big/int.go?s=13815:13841#L580
func jacobiCompact(x, y *Decimal) int {
	if y.compact&1 == 0 {
		panic(fmt.Sprintf("decimal: invalid 2nd argument to decimal.Jacobi: need odd integer but got %d", y.compact))
	}

	a := abs(x.compact)
	b := abs(y.compact)
	j := 1

	if y.compact < 0 {
		if x.compact < 0 {
			j = -1
		}
	}

	for {
		if b == 1 {
			return j
		}
		if a == 0 {
			return 0
		}

		// Euclidean modulus.
		a %= b
		if a < 0 {
			if b < 0 {
				a -= b
			} else {
				a += b
			}
		}

		if a == 0 {
			return 0
		}

		s := trailingZeroBits(a)
		if s&1 != 0 {
			bmod8 := b & 7
			if bmod8 == 3 || bmod8 == 5 {
				j = -j
			}
		}
		c := a << s
		if b&3 == 3 && c&3 == 3 {
			j = -j
		}
		a, b = b, c
	}
}

func jacobiHalf(comp, nc *Decimal) int {
	return big.Jacobi(big.NewInt(comp.compact), &nc.mantissa)
}

// LessThan returns true if d is less than x.
func (z *Decimal) LessThan(x *Decimal) bool {
	return z.Cmp(x) < 0
}

// LessThanEQ returns true if d is less than or equal to x.
func (z *Decimal) LessThanEQ(x *Decimal) bool {
	return z.Cmp(x) <= 0
}

// Log10 sets z the base-10 logarithm of x and returns z.
func (z *Decimal) Log10(x *Decimal) *Decimal {
	if x.ltez() {
		panic("decimal.Log10: x <= 0")
	}
	return z.Div(z.Log(x), ln10)
}

// Ilog10 returns the base-10 integer logarithm of z
// rounded up. (i.e., the number of zigits in d)
func (z *Decimal) Ilog10() int64 {
	shift := int64(0)
	if z.scale < 0 {
		shift = -z.scale
	}
	if z.compact != overflown {
		x := z.compact
		if x < 0 {
			x = -x
		}
		if x < 10 {
			return 1
		}
		// Originally I just had a slightly different method of
		// finding the Ilog10 where I took 64-clz(x)*3/10
		// but this saves us from both a divide instruction as well as
		// loading a 64-element table.
		// (From https://graphics.stanforz.edu/~seander/bithacks.html)
		r := (((63 - clz(x) + 1) * 1233) >> 12) + shift
		if v := pow10int64(r); v == overflown || x < v {
			return r
		}
		return r + 1
	}
	// calculate with the big.Int mantissa.

	if z.mantissa.Sign() == 0 {
		return 1
	}

	// Accurate up to as high as we can report.
	r := (((int64(z.mantissa.BitLen()) + 1) * 0x268826A1) >> 31) + shift
	if cmpBigAbs(z.mantissa, bigPow10(r)) < 0 {
		return r
	}
	return r + 1
}

// Log2 sets z the base-2 logarithm of x and returns z.
func (z *Decimal) Log2(x *Decimal) *Decimal {
	if x.ltez() {
		panic("decimal.Log2: x <= 0")
	}
	return z.Div(z.Log(x), ln2)
}

func (z *Decimal) Log(x *Decimal) *Decimal {
	if x.ltez() {
		panic("decimal.Log: x <= 0")
	}

	mag := x.Ilog10() - x.scale - 1
	if mag < 3 {
		return z.logNewton(x)
	}
	root := z.integralRoot(x, mag)
	lnRoot := root.logNewton(root)
	return z.Mul(New(mag, 0), lnRoot)
}

// logNewton sets z to the natural logarithm of x
// using the Newtonian method and returns z.
func (z *Decimal) logNewton(x *Decimal) *Decimal {
	sp := z.Prec() + 1
	x0 := new(Decimal).Set(x)
	tol := New(5, sp)

	var term, etx Decimal
	term.ctx.Prec = sp
	etx.ctx.Prec = sp
	for {
		etx.exp(x0).SetPrec(sp)
		term.Sub(&etx, x)
		term.Div(&term, &etx).SetPrec(sp)
		x0.Sub(x0, &term)
		if term.LessThanEQ(tol) {
			break
		}
	}
	*z = *x0
	return z.SetPrec(sp - 1)
}

// exp sets z to e ** x and returns z.
func (z *Decimal) exp(x *Decimal) *Decimal {
	if x.ez() {
		p := z.Prec()
		z.SetInt64(1)
		z.ctx.Prec = p
		return z
	}

	if x.ltz() {
		x0 := new(Decimal).Set(x)
		// 1 / (e ** -x)
		return z.Div(New(1, 0), x0.exp(x0.Neg(x0)))
	}

	dec, frac := Modf(x)
	if dec.ez() {
		return z.taylor(x)
	}

	o := New(1, 0)
	o.ctx.Prec = z.Prec()
	o.Add(z, frac.Div(frac, dec))
	o.taylor(o)

	res := New(1, 0)
	for dec.GreaterThanEQ(dmaxInt64) {
		res.Mul(res, new(Decimal).powInt(o, math.MaxInt64))
		dec.Sub(dec, dmaxInt64)
	}
	return z.Mul(res, o.powInt(o, dec.Int64()))
}

// taylor sets z to e ** x using the Taylor series and returns z.
func (z *Decimal) taylor(x *Decimal) *Decimal {
	sum := new(Decimal).Add(New(1, 0), x).SetPrec(z.Prec())
	xp := new(Decimal).Set(x).SetPrec(z.Prec())
	fac := New(1, 0)

	var sp, term Decimal
	term.ctx.Prec = z.Prec()
	for i := int64(2); ; i++ {
		xp.Mul(xp, x).SetPrec(z.Prec())
		fac.Mul(fac, New(i, 0)).SetPrec(z.Prec())
		term.Div(xp, fac)
		sp.Set(sum)
		sum.Add(sum, &term)
		if sum.Equals(&sp) {
			break
		}
	}
	*z = *sum
	return z
}

// Max returns the larger of x or y.
func Max(x *Decimal, y *Decimal) *Decimal {
	if x.GreaterThan(y) {
		return x
	}
	return y
}

// Min returns the smaller of x or y.
func Min(x *Decimal, y *Decimal) *Decimal {
	if x.LessThan(y) {
		return x
	}
	return y
}

// Mod sets z to the modulus x%y for y != 0 and returns z.
// If y == 0, a division-by-zero run-time panic occurs.
func (z *Decimal) Mod(x, y *Decimal) *Decimal {
	var q Decimal
	_, mod := q.DivMod(x, y, z)
	return mod
}

// Modf returns the decomposed integral and fractional parts of the
// value of x.
func Modf(x *Decimal) (int *Decimal, frac *Decimal) {
	int = &Decimal{ctx: Context{Prec: x.Prec()}}
	frac = &Decimal{
		scale: x.scale,
		ctx:   Context{Prec: x.Prec()},
	}
	if x.compact != overflown {
		i, f := modi(x.compact, x.scale)
		int.compact = i
		frac.compact = f
	} else {
		i, f := modbig(&x.mantissa, x.scale)
		int.compact = overflown
		int.mantissa.Set(i)
		frac.compact = overflown
		frac.mantissa.Set(f)
	}
	return int, frac
}

// Mul sets z to x * y and returns z.
func (z *Decimal) Mul(x, y *Decimal) *Decimal {
	if x.compact != overflown {
		if y.compact != overflown {
			return z.mulCompact(x, y)
		}
		return z.mulHalf(x, y)
	}
	if y.compact != overflown {
		return z.mulHalf(y, x)
	}
	return z.mulBig(x, y)
}

func (z *Decimal) mulCompact(x, y *Decimal) *Decimal {
	prod := prod(x.compact, y.compact)
	if prod != overflown {
		z.compact = prod
	} else {
		z.mantissa.Mul(big.NewInt(x.compact), big.NewInt(y.compact))
		z.compact = overflown
	}
	z.scale = safeScale2(x.scale, sum(x.scale, y.scale))
	return z
}

// mulHalf multiplies a compact Decimal with a non-compact
// Decimal.
// Let the first arg be the compact and the second the non-compact.
func (z *Decimal) mulHalf(comp, nc *Decimal) *Decimal {
	if comp.compact == overflown {
		panic("decimal.Mul: (bug) comp should != overflown")
	}
	if comp.scale == nc.scale {
		z.scale = safeScale2(comp.scale, sum(comp.scale, nc.scale))
		z.mantissa.Mul(big.NewInt(comp.compact), &nc.mantissa)
		z.compact = overflown
		return z
	}
	return z.mulBig(&Decimal{
		mantissa: *big.NewInt(comp.compact),
		scale:    comp.scale,
	}, nc)
}

func (z *Decimal) mulBig(x, y *Decimal) *Decimal {
	z.scale = safeScale2(x.scale, sum(x.scale, y.scale))
	z.mantissa.Mul(&x.mantissa, &y.mantissa)
	z.compact = overflown
	return z
}

// MulRange sets z to the product of all integers in the range
// [a, b] inclusively and returns z.
// If a > b (empty range) the result is 1.
func (z *Decimal) MulRange(a, b int64) *Decimal {
	switch {
	case a > b:
		return z.SetInt64(1)
	case a <= 0 && b >= 0:
		return z.SetInt64(0)
	}
	neg := false
	if a < 0 {
		neg = (b-a)&1 == 0
		a, b = -b, -a
	}
	z.mulRange(a, b)
	if neg {
		z.Neg(z)
	}
	return z
}

func (z *Decimal) mulRange(a, b int64) *Decimal {
	switch {
	case a == 0:
		return z.SetInt64(0)
	case a > b:
		return z.SetInt64(1)
	case a == b:
		return z.SetInt64(a)
	case a+1 == b:
		return z.Mul(new(Decimal).SetInt64(a), new(Decimal).SetInt64(b))
	}
	m := (a + b) / 2
	return z.Mul(new(Decimal).mulRange(a, m), new(Decimal).mulRange(m+1, b))
}

// Neg sets z to -x and returns z.
func (z *Decimal) Neg(x *Decimal) *Decimal {
	if x.compact != overflown {
		z.compact = -x.compact
	} else {
		z.mantissa.Neg(&x.mantissa)
		z.compact = overflown
	}
	z.scale = x.scale
	return z
}

// Rat returns the rational number representation of z.
// If x is non-nil, Rat stores the result in x instead of
// allocating a new Rat.
func (z *Decimal) Rat(x *big.Rat) *big.Rat {
	// TODO(eric):
	// We use big.Ints here when technically we could use our
	// int64 with big.Rat's SetInt64 methoz.
	// I'm not sure if it'll be an optimization or not.
	var num, denom big.Int
	if z.compact != overflown {
		c := big.NewInt(z.compact)
		if z.scale >= 0 {
			num.Set(c)
			denom.Set(mulBigPow10(oneInt, z.scale))
		} else {
			num.Set(mulBigPow10(c, -z.scale))
			denom.SetInt64(1)
		}
	} else {
		if z.scale >= 0 {
			num.Set(&z.mantissa)
			denom.Set(mulBigPow10(oneInt, z.scale))
		} else {
			num.Set(mulBigPow10(&z.mantissa, -z.scale))
			denom.SetInt64(1)
		}
	}
	if x != nil {
		return x.SetFrac(&num, &denom)
	}
	return new(big.Rat).SetFrac(&num, &denom)
}

// Rem sets z to the remainder of x/y and returns z.
func (z *Decimal) Rem(x, y *Decimal) *Decimal {
	if y.ez() {
		panic("decimal.Rem: division by zero")
	}

	if x.compact != overflown {
		if y.compact != overflown {
			z.scale = safeScale(x.compact, x.scale, sub(x.scale, y.scale))
			z.compact = x.compact % y.compact
			return z
		}
		return z.remBig(&Decimal{
			mantissa: *big.NewInt(x.compact),
			scale:    x.scale,
		}, y)
	}
	if y.compact != overflown {
		return z.remBig(x, &Decimal{
			mantissa: *big.NewInt(y.compact),
			scale:    y.scale,
		})
	}
	return z.remBig(x, y)
}

func (z *Decimal) remBig(x, y *Decimal) *Decimal {
	if x.scale == y.scale {
		z.scale = x.scale
		z.mantissa.Rem(&x.mantissa, &y.mantissa)
	} else {
		z.scale = safeScale(x.compact, x.scale, sub(x.scale, y.scale))
		z.mantissa.Rem(&x.mantissa, &y.mantissa)
	}
	z.compact = overflown
	return z
}

// rsh sets to d to x >> n and returns z.
func (z *Decimal) rsh(x *Decimal, n uint64) *Decimal {
	if x.compact != overflown {
		z.compact = x.compact >> n
	} else {
		z.mantissa.Rsh(&x.mantissa, uint(n))
		z.compact = overflown
	}
	return z
}

// Precision returns d's precision.
// This method does not necessarily return the value inside
// inside d's Context member.
func (z *Decimal) Prec() int64 {
	// Zero (default value) is assumed to mean
	// the default precision. This allows us to
	// still use new(Decimal).Div(x, y) and still
	// keep a reasonable amount of precision.
	//
	// Values < 0 are inferred to be zero. The only
	// value we manually set to be less than zero is
	// "overflown", which is set when somebody calls
	// SetPrec(0).
	switch p := z.ctx.Prec; {
	case p > 0:
		return p
	case p < 0:
		return 0
	default:
		return DefaultPrecision
	}
}

// Scale returns the scale component of z.
func (z *Decimal) Scale() int64 {
	return z.scale
}

// Set sets z to exactly x and returns z.
func (z *Decimal) Set(x *Decimal) *Decimal {
	if z != x {
		*z = Decimal{
			compact: x.compact,
			scale:   x.scale,
			ctx:     x.ctx,
		}
		if x.compact == overflown {
			z.mantissa.Set(&x.mantissa)
		}
	}
	return z
}

// SetBytes interprets buf as the bytes of a big-endian, usigned
// integer, sets z to that value, and returns z.
func (z *Decimal) SetBytes(buf []byte) *Decimal {
	z.mantissa.SetBytes(buf)
	z.compact = overflown
	return z
}

// SetContext sets z's context to ctx.
func (z *Decimal) SetContext(ctx Context) *Decimal {
	z.ctx = ctx
	return z
}

// SetInt64 sets z to x and returns z.
// As with SetInt64, it sets z's scale to 0.
func (z *Decimal) SetInt(x *big.Int) *Decimal {
	z.scale = 0
	z.compact = overflown
	z.mantissa.Set(x)
	return z
}

// SetInt64 sets z to x and returns z.
// It sets z's scale to 0.
func (z *Decimal) SetInt64(x int64) *Decimal {
	z.scale = 0
	z.compact = x
	z.mantissa.SetBits(nil)
	return z
}

// SetPrec sets z's precision to prec and returns a possibly
// rounded z.
//
// If you only want to set the Context's Prec member you should do that
// manually.
//
// A negative prec or calling this method on a Decimal with
// a negative scale is undefined behavior.
func (z *Decimal) SetPrec(prec int64) *Decimal {
	// Our undefined behavior.
	if z.scale < 0 || prec < 0 {
		return z
	}

	if prec == 0 {
		z.ctx.Prec = overflown
	} else {
		z.ctx.Prec = prec
	}

	switch {
	case z.scale == prec:
		return z
	case z.ez():
		z.SetInt64(0)
		z.scale = prec
		return z
	}

	shift := prec - z.scale
	if z.compact != overflown {

		res := z.compact
		// prec > z.scale
		if shift > 0 {
			z.compact = mulPow10(z.compact, shift)
		} else {
			z.compact /= pow10int64(-shift)
		}

		if z.compact != overflown {
			z.scale += shift
			return z
		}

		z.mantissa.SetInt64(res)
	}

	if shift > 0 {
		z.mantissa = *mulBigPow10(&z.mantissa, shift)
	} else {
		b := bigPow10(-shift)
		z.mantissa.Div(&z.mantissa, &b)
	}
	z.scale += shift
	return z
}

// ShiftRadix shifts d's radix (decimal point) n places, with a negative value meaning to
// the right and a positive value meaning to the left.
func (z *Decimal) ShiftRadix(n int64) *Decimal {
	scale := safeScale2(z.scale, sum(z.scale, n))
	if scale == overflown {
		panic("decimal.ShiftRadix: shift overflows scale")
	}
	z.scale = scale
	if z.scale < 0 {
		return z.SetPrec(0)
	}
	return z
}

// Shrink shrinks d from a big.Int into its integer member
// if possible and returns z.
func (z *Decimal) Shrink() *Decimal {
	if z.compact == overflown &&
		(z.mantissa.Sign() > 0 && z.mantissa.Cmp(maxInt64) < 0 ||
			z.mantissa.Sign() < 0 && z.mantissa.Cmp(minInt64) > 0) {
		z.compact = z.mantissa.Int64()
		z.mantissa.SetBits(nil)
	}
	return z
}

// Sign returns:
//
//	-1 if z <  0
//	 0 if z == 0
//	+1 if z >  0
//
func (z *Decimal) Sign() int {
	if z.compact != overflown {
		if z.compact == 0 {
			return 0
		}
		if z.compact < 0 {
			return -1
		}
		return +1
	}
	return z.mantissa.Sign()
}

// Signbit returns true if d is negative.
func (z *Decimal) Signbit() bool {
	return z.Sign() < 0
}

// String returns the string representation of the decimal
// with the fixed point.
//
// Example:
//
//     d := New(-12345, 3)
//     println(z.String())
//
// Output:
//
//     -12.345
//
func (z *Decimal) String() string {
	switch {
	case z == nil:
		return "<nil>"
	case IsNan(z):
		return "NaN"
	default:
		return z.toString(trimZeros | plain)
	}
}

// Scientific returns the scientific notation of z
func (z *Decimal) Scientific() string {
	switch {
	case z == nil:
		return "<nil>"
	case IsNan(z):
		return "NaN"
	default:
		return z.toString(trimZeros | scientific)
	}
}

// StringFixed returns a rounded fixed-point string with places digits after
// the decimal point.
//
// Example:
//
// 	   NewFromFloat(0).StringFixed(2) // output: "0.00"
// 	   NewFromFloat(0).StringFixed(0) // output: "0"
// 	   NewFromFloat(5.45).StringFixed(0) // output: "5"
// 	   NewFromFloat(5.45).StringFixed(1) // output: "5.5"
// 	   NewFromFloat(5.45).StringFixed(2) // output: "5.45"
// 	   NewFromFloat(5.45).StringFixed(3) // output: "5.450"
// 	   NewFromFloat(545).StringFixed(-1) // output: "550"
//
func (z *Decimal) StringFixed(places int64) string {
	return z.SetPrec(places).toString(trimZeros | plain)
}

// Sqrt sets z to the square root of x and returns z.
// The precision of Sqrt is determined by d's Context.
// Sqrt will panic on negative values since Decimal cannot
// represent imaginary numbers.
func (z *Decimal) Sqrt(x *Decimal) *Decimal {

	if x.Sign() < 0 {
		panic("decimal.Sqrt: cannot take square root of negative number")
	}

	// Check if x is a perfect square. If it is, we can avoid having to
	// inflate x and can possibly use can use the hardware SQRT.
	// Note that we can only catch perfect squares that aren't big.Ints.
	if sq, ok := x.perfectSquare(); ok {
		z.compact = sq
		z.scale = 0
		z.ctx = x.ctx
		return z
	}
	// x isn't a perfect square or x is a big.Int

	var a, n, ix, p big.Int

	// We just need to shift the radix so don't make
	// a deep copy with the Set methoz.
	x0 := *x
	n.Set(x0.ShiftRadix(-(z.Prec() << 1)).Int())
	ix.Rsh(&n, uint((n.BitLen()+1)>>1))

	for {
		p.Set(&ix)
		ix.Add(&ix, a.Div(&n, &ix)).Rsh(&ix, 1)
		if ix.Cmp(&p) == 0 {
			z.mantissa = ix
			z.scale = z.Prec()
			z.compact = overflown
			return z.Shrink()
		}
	}
}

// perfectSquare algorithm slightly partially borrowed from
// http://stackoverflow.com/a/18686659/2967113
func (z *Decimal) perfectSquare() (square int64, ok bool) {
	if z.IsBig() {
		return
	}
	x0 := z.compact
	if x0 == 0 {
		return 0, true
	}
	tlz := uint64(ctz(x0))
	x0 >>= tlz
	if tlz&1 == 0 && (x0&7 != 1 || x0 == 0) {

		// "Show that floating point sqrt(x*x) >= x for all long x."
		// http://math.stackexchange.com/a/238885/153292
		tst := int64(math.Sqrt(float64(z.compact)))
		return tst, tst*tst == z.compact
	}
	return
}

// Sub sets z to x - y and returns z.
func (z *Decimal) Sub(x, y *Decimal) *Decimal {
	return z.Add(x, new(Decimal).Neg(y))
}

// Trunc sets z to x truncated down to n digits of precision
// (assuming x.scale > 0) and returns d
//
// Precision is the last digit that will not be truncated.
// If x's scale < 0, n < 0, or n > x's scale d will be set to x and no
// truncation will occur.
//
// Example:
//
//     NewFromString("123.456").Trunc(2).String() // output: "123.45"
//     NewFromString("123.456").Trun(0).String() // output: "123"
func (z *Decimal) Trunc(x *Decimal, n int64) *Decimal {
	if x.scale > 0 && n > 0 && n < x.scale {
		if x.compact != overflown {
			z.compact = x.compact / pow10int64(z.scale-n)
		} else {
			b := bigPow10(z.scale - n)
			z.mantissa.Div(&x.mantissa, &b)
		}
		z.scale = n
	}
	return z
}

// strOpts are ORd together.
type strOpts uint8

const (
	trimZeros strOpts = 1 << iota
	plain
	scientific
)

func (z *Decimal) toString(opts strOpts) string {
	// Fast path: return our value as-is.
	if z.scale == 0 {
		if z.compact == overflown {
			return z.mantissa.String()
		}
		return strconv.FormatInt(z.compact, 10)
	}

	var (
		str string
		neg bool
		b   bytes.Buffer
	)

	if z.compact == overflown {
		str = new(big.Int).Abs(&z.mantissa).String()
		neg = z.mantissa.Sign() < 0
	} else {
		abs := uint64(abs(z.compact))
		str = strconv.FormatUint(abs, 10)
		neg = z.compact < 0
	}

	if neg {
		b.WriteByte('-')
	}

	if z.scale < 0 {
		if z.ez() {
			return "0"
		}
		b.WriteString(str)
		b.Write(bytes.Repeat([]byte{'0'}, -int(z.scale)))
		return b.String()
	}

	switch p := int64(len(str)) - z.scale; {
	case p == 0:
		b.Write([]byte{'0', '.'})
		b.WriteString(str)
	case p > 0:
		b.WriteString(str[:p])
		b.WriteByte('.')
		b.WriteString(str[p:])
	default:
		b.Write([]byte{'0', '.'})
		b.Write(bytes.Repeat([]byte{'0'}, -int(p)))
		b.WriteString(str)
	}

	if opts&trimZeros != 0 {
		buf := b.Bytes()
		i := b.Len() - 1
		for ; i >= 0 && buf[i] == '0'; i-- {
		}
		if buf[i] == '.' {
			i--
		}
		b.Truncate(i + 1)
	}
	return b.String()
}
