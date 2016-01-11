package decimal

import (
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
	compact int64
	scale   int64
	ctx     Context

	// Don't shrink big.Int -> int64 even if it can fit.
	//
	// I'm concerned about swapping back and forth repeatedly
	// which could degrade performance. This could happen if
	// repeated mul/div/add/sub routines are called.
	//
	// I'm sure we could use some sort of internal counter to
	// determine if this is happening and act accordingly.
	mantissa big.Int
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
//     d, err := NewFromString("-123.45")
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

	d := Decimal{
		compact:  overflown,
		scale:    int64(exp),
		mantissa: *dValue,
	}
	return d.Shrink(), nil
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
// Approximately 2.3% of Decimals created from floats will have a rounding
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

	d := Decimal{scale: scale}

	// Given float64(math.MaxInt64) == math.MaxInt64
	if value <= math.MaxInt64 {
		// TODO(eric):
		// Should we put an integer that's so close to overflowing inside
		// the compact member?
		d.compact = int64(value)
	} else {
		// Given float64(math.MaxUint64) == math.MaxUint64
		if value <= math.MaxUint64 {
			d.mantissa.SetUint64(uint64(value))
		} else {
			d.mantissa.Set(bigIntFromFloat(value))
		}
		d.compact = overflown
	}
	return &d
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

	a := new(big.Int)
	a.SetUint64(mantissa)
	return a.Lsh(a, uint(shift))
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

// Abs sets d to the absolute value of x and returns d.
func (d *Decimal) Abs(x *Decimal) *Decimal {
	if x.compact != overflown {
		d.compact = abs(x.compact)
	} else {
		d.mantissa.Abs(&x.mantissa)
		d.compact = overflown
	}
	d.scale = x.scale
	return d
}

// Add sets d to x + y and returns d.
func (d *Decimal) Add(x, y *Decimal) *Decimal {
	// The Mul method follows the same steps as Add, so I'll detail the
	// formula in the various add methods.
	if x.compact != overflown {
		if y.compact != overflown {
			return d.addCompact(x, y)
		}
		return d.addHalf(x, y)
	}
	if y.compact != overflown {
		return d.addHalf(y, x)
	}
	return d.addBig(x, y)
}

// addCompact set d to the sum of x and y and returns d.
// Each case depends on the scales.
func (d *Decimal) addCompact(x, y *Decimal) *Decimal {
	// Fast path: we don't need to adjust anything.
	// Just check for overflows (if so, use a big.Int)
	// and return the result.
	if x.scale == y.scale {
		d.scale = x.scale
		sum := sum(x.compact, y.compact)
		if sum != overflown {
			d.compact = sum
		} else {
			d.mantissa.Add(big.NewInt(x.compact), big.NewInt(y.compact))
			d.compact = overflown
		}
		return d
	}

	// Guess the high and low scale. If we guess wrong, swap.
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	// Find which power of 10 we have to multiply our low value by in order
	// to equalize their scales.
	inc := safeScale(lo.compact, hi.scale, sub(hi.scale, lo.scale))

	d.scale = hi.scale

	// Expand the low value (checking for overflows) and
	// find the sum (checking for overflows).
	//
	// If we overflow at all use a big.Int to calculate the sum.
	scaledLo := mulPow10(lo.compact, inc)
	if scaledLo != overflown {
		sum := sum(hi.compact, scaledLo)
		if sum != overflown {
			d.compact = sum
			return d
		}
	}

	scaled := mulBigPow10(big.NewInt(lo.compact), inc)
	d.mantissa.Add(scaled, big.NewInt(hi.compact))
	d.compact = overflown
	return d
}

// addHalf adds a compact Decimal with a non-compact
// Decimal.
// Let the first arg be the compact and the second the non-compact.
func (d *Decimal) addHalf(comp, nc *Decimal) *Decimal {
	if comp.compact == overflown {
		panic("decimal.Add: (bug) comp should != overflown")
	}
	if comp.scale == nc.scale {
		d.mantissa.Add(big.NewInt(comp.compact), &nc.mantissa)
		d.scale = comp.scale
		d.compact = overflown
		return d
	}
	// Since we have to rescale we need to add two big.Ints
	// together because big.Int doesn't have an API for
	// increasing its value by an integer.
	return d.addBig(&Decimal{
		mantissa: *big.NewInt(comp.compact),
		scale:    comp.scale,
	}, nc)
}

func (d *Decimal) addBig(x, y *Decimal) *Decimal {
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	inc := safeScale(lo.compact, hi.scale, sub(hi.scale, lo.scale))
	scaled := mulBigPow10(&lo.mantissa, inc)
	d.mantissa.Add(&hi.mantissa, scaled)
	d.compact = overflown
	d.scale = hi.scale
	return d
}

// and sets d to to x & n and returns d.
func (d *Decimal) and(x *Decimal, n int64) *Decimal {
	if x.compact != overflown {
		d.compact = x.compact & n
	} else {
		d.mantissa.And(&x.mantissa, big.NewInt(n))
		d.compact = overflown
	}
	// Save an assignment if d is x.
	if d != x {
		d.scale = x.scale
		d.ctx = x.ctx
	}
	return d
}

// Binomial sets d to the binomial coefficient of (n, k) and returns d.
func (d *Decimal) Binomial(n, k int64) *Decimal {
	if n/2 < k && k <= n {
		k = n - k
	}
	var a, b Decimal
	a.MulRange(n-k+1, n)
	b.MulRange(1, k)
	return d.IntDiv(&a, &b)
}

// BitLen returns the absolute value of d in bits.
func (d *Decimal) BitLen() int64 {
	if d.compact != overflown {
		x := d.compact
		if d.scale < 0 {
			x = mulPow10(x, -d.scale)
		}
		if x != overflown {
			return (64 - clz(x))
		}
	}
	x := &d.mantissa
	if d.scale < 0 {
		// Double check because we fall through if
		// mulPow10(x, -d.scale) returns overflown.
		if d.compact != overflown {
			x = mulBigPow10(big.NewInt(d.compact), -d.scale)
		} else {
			x = mulBigPow10(x, -d.scale)
		}
	}
	return int64(x.BitLen())
}

// Bytes returns the absolute value of d as a big-endian
// byte slice.
func (d *Decimal) Bytes() []byte {
	if d.compact != overflown {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(abs(d.compact)))
		return b[:]
	}
	return d.mantissa.Bytes()
}

// Ceil sets d to the nearest integer value greater than or equal to x
// and returns d.
func (d *Decimal) Ceil(x *Decimal) *Decimal {
	d.Floor(d.Neg(x))
	return d.Neg(d)
}

// Cmp compares d and x and returns:
//
//   -1 if d <  x
//    0 if d == x
//   +1 if d >  x
//
// It does not modify d or x.
func (d *Decimal) Cmp(x *Decimal) int {
	// Check for same pointers.
	if d == x {
		return 0
	}

	// Same scales means we can compare straight across.
	if d.scale == x.scale &&
		d.compact != overflown && x.compact != overflown {
		if d.compact > x.compact {
			return +1
		}
		if d.compact < x.compact {
			return -1
		}
		return 0
	}

	// Different scales -- check signs and/or if they're
	// both zero.

	ds := d.Sign()
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
	dl := d.Ilog10() - d.scale
	xl := x.Ilog10() - x.scale
	if dl > xl {
		return +1
	}
	if dl < xl {
		return -1
	}

	// We need to inflate one of the numbers.

	dc := d.compact // hi
	xc := x.compact // lo

	var swap bool

	hi, lo := d, x
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
				return mulBigPow10(&d.mantissa, diff).
					Cmp(&x.mantissa)
			}
			// x is lo
			return d.mantissa.Cmp(mulBigPow10(&x.mantissa, diff))
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
		return d.mantissa.Cmp(big.NewInt(xc))
	}
	return d.mantissa.Cmp(&x.mantissa)
}

// Dim sets d to the maximum of x - y or 0 and returns d.
func (d *Decimal) Dim(x, y *Decimal) *Decimal {
	x0 := new(Decimal).Sub(x, y)
	return Max(x0, New(0, 0))
}

// DivMod sets d to the quotient x div y and m to the modulus x mod y and
// returns the pair (z, m) for y != 0. If y == 0, a division-by-zero run-time panic occurs
func (d *Decimal) DivMod(x, y, m *Decimal) (div *Decimal, mod *Decimal) {
	if y.ez() {
		panic("decimal.DivMod: division by zero")
	}

	if x.ez() {
		d.compact = 0
		d.scale = safeScale2(x.scale, sub(x.scale, y.scale))
		return d, m.SetInt64(0)
	}

	if x.compact != overflown {
		if y.compact != overflown {
			if m.compact != overflown {
				return d.divCompact(x, y, m)
			}
			return d.divBig(x, y, &Decimal{
				mantissa: *big.NewInt(m.compact),
				scale:    m.scale,
			})
		}
		return d.divBig(&Decimal{
			mantissa: *big.NewInt(x.compact),
			scale:    x.scale,
		}, y, m)
	}
	if y.compact != overflown {
		return d.divBig(x, &Decimal{
			mantissa: *big.NewInt(y.compact),
			scale:    y.scale,
		}, m)
	}
	return d.divBig(x, y, m)
}

// Div sets d to the quotient x/y for y != 0 and returns d. If y == 0, a
// division-by-zero run-time panic occurs.
func (d *Decimal) Div(x, y *Decimal) *Decimal {
	var r Decimal
	div, _ := d.DivMod(x, y, &r)
	return div
}

func (d *Decimal) needsInc(x, r int64, pos, odd bool) bool {
	m := 1
	if r > math.MinInt64/2 || r <= math.MaxInt64/2 {
		m = cmpAbs(r<<1, x)
	}
	return d.ctx.Mode.needsInc(m, pos, odd)
}

func (d *Decimal) needsIncBig(x, r *big.Int, pos, odd bool) bool {
	var x0 big.Int
	m := cmpBigAbs(*x0.Mul(r, twoInt), *x)
	return d.ctx.Mode.needsInc(m, pos, odd)
}

func (d *Decimal) divCompact(x, y, m *Decimal) (div *Decimal, mod *Decimal) {

	shift := d.Prec()

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
			d.compact = q
			if r != 0 && d.needsInc(y.compact, r, sign > 0, q&1 != 0) {
				d.compact += sign
			}
			d.scale = safeScale2(x.scale, x.scale-y.scale+shift)
			return d.SetPrec(shift), m.SetInt64(r)
		}
	}
	return d.divBig(
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

func (d *Decimal) divBig(x, y, m *Decimal) (div *Decimal, mod *Decimal) {

	shift := d.Prec()

	x0 := mulBigPow10(&x.mantissa, shift)
	q, r := x0.DivMod(x0, &y.mantissa, &m.mantissa)

	sign := int64(1)
	if (x.mantissa.Sign() < 0) && (y.mantissa.Sign() < 0) {
		sign = -1
	}

	d.mantissa = *q
	m.mantissa = *r

	odd := new(big.Int).And(q, oneInt).Cmp(zeroInt) != 0
	if r.Cmp(zeroInt) != 0 && d.needsIncBig(&y.mantissa, r, sign > 0, odd) {
		d.mantissa.Add(&d.mantissa, big.NewInt(sign))
	}

	d.scale = safeScale2(x.scale, x.scale-y.scale+shift)

	d.compact = overflown
	m.compact = overflown

	// I'm only comfortable calling shrink here because division
	// has a tendency to blow up numbers real big and then
	// shrink them back down.
	return d.Shrink().SetPrec(shift), m.Shrink()
}

// Equals returns true if d == x.
func (d *Decimal) Equals(x *Decimal) bool {
	return d.Cmp(x) == 0
}

// The following are some internal optimizations when we need to compare a
// Decimal to zero since d's comparison methods aren't optimized for 'zero'.

// ez returns true if d == 0.
func (d *Decimal) ez() bool {
	return d.Sign() == 0
}

// ltz returns true if d < 0
func (d *Decimal) ltz() bool {
	return d.Sign() < 0
}

// ltez returns true if d <= 0
func (d *Decimal) ltez() bool {
	return d.Sign() <= 0
}

// gtz returns true if d > 0
func (d *Decimal) gtz() bool {
	return d.Sign() > 0
}

// gtez returns true if d >= 0
func (d *Decimal) gtez() bool {
	return d.Sign() >= 0
}

// Exp sets d to x**y mod |m| and returns d and a boolean indicating
// whether or not the exponentiation was successful.
//
// If m == nil or m == 0, d == z**y.
// If y <= the result is 1 mod |m|.
//
// Note that Exp is limited in its capabilities
// because Decimals are already "big" numbers.
//
// Since a Decimal is `v * 10 ^ scale`,
// (I.e. 1.23 == 1.23 * 10 ^ 2 == 123)
// The scale will overflow iff x.scale * |y| > 9223372036854775807
//
// d will not be touched unless the returned bool is true.
func (d *Decimal) Exp(x, y, m *Decimal) (*Decimal, bool) {
	if m == nil {
		m = new(Decimal)
	}

	if y.Equals(one) {
		return New(1, 0), true
	}

	if y.Equals(ptFive) && m.ez() {
		return d.Sqrt(x), true
	}

	xbig := x.compact == overflown
	ybig := y.compact == overflown
	mbig := m.compact == overflown

	if !(xbig || ybig || mbig) {
		scale := prod(x.scale, y.compact)
		if scale == overflown {
			return nil, false
		}

		// If y is an int compute it by squaring (O log n).
		// Otherwise, use exp(log(x) * y).
		if y.IsInt() {
			d.pow(x, y)
		} else {
			x0 := new(Decimal).Set(x)
			neg := x0.ltz()
			if neg {
				x0.Neg(x0)
			}
			x0.Log(x0).Mul(x0, y)
			d.exp(x0)
			if neg {
				d.Neg(d)
			}
		}

		if !m.ez() {
			m.Mod(d, m)
		}
		return d, true
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
		return nil, false
	}
	// y <= 9223372036854775807

	intY := y0.Int64()

	// Check for scale overflow.
	scale := prod(x.scale, intY)
	if scale == overflown {
		return nil, false
	}
	d.mantissa.Exp(x0, y0, m0)
	d.scale = scale
	d.compact = overflown
	return d, true
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

// Fib sets d to the Fibonacci number x and returns d.
func (d *Decimal) Fib(x *Decimal) *Decimal {
	if x.LessThan(two) {
		return d.Set(x)
	}
	a, b, x0 := New(0, 0), New(1, 0), new(Decimal).Set(x)
	for x0.Sub(x0, one); x0.gtz(); x0.Sub(x0, one) {
		a.Add(a, b)
		a, b = b, a
	}
	*d = *b
	return d
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
// For more details, see the documentation for big.Rat.Float64
func (d *Decimal) Float64() (f float64, exact bool) {
	return d.Rat(nil).Float64()
}

// Floor sets d to the nearest integer value less than or equal to x
// and returns d.
func (d *Decimal) Floor(x *Decimal) *Decimal {
	if x.compact != overflown {
		if x.compact == 0 {
			d.compact = 0
			d.scale = 0
			return d
		}

		if x.compact < 0 {
			dec, frac := modi(-x.compact, x.scale)
			if frac != 0 {
				dec++
			}
			if dec-1 != overflown {
				d.compact = -dec
				d.scale = 0
				return d
			}
		} else {
			dec, _ := modi(x.compact, x.scale)
			if dec != overflown {
				d.compact = dec
				d.scale = 0
				return d
			}
		}

		// If we reach here then we can't find the floor without using
		// big.Ints to do the math for us.
		d0 := new(Decimal).Set(x)
		d0.mantissa = *big.NewInt(x.compact)
		d0.compact = overflown
		return d.Floor(d0)
	}

	if cmp := x.mantissa.Cmp(zeroInt); cmp == 0 {
		d.mantissa.Set(zeroInt)
	} else if cmp < 0 {
		neg := new(big.Int).Neg(&x.mantissa)
		dec, frac := modbig(neg, x.scale)
		if frac.Cmp(zeroInt) != 0 {
			dec.Add(dec, oneInt)
		}
		d.mantissa.Set(dec.Neg(dec))
	} else {
		dec, _ := modbig(&x.mantissa, x.scale)
		d.mantissa.Set(dec)
	}
	d.compact = overflown
	d.scale = 0
	return d
}

// GCD sets d to the greatest common divisor of a and b
// (both of which must be > 0) and returns d.
// If x and y are not nil, GCD sets x and y such that
// d = a*x + b*y.
// If either a or b is <= 0, GCD sets d = x = y = 0.
func (d *Decimal) GCD(x, y, a, b *Decimal) *Decimal {
	return nil
}

// GreaterThan returns true if d is greater than x.
func (d *Decimal) GreaterThan(x *Decimal) bool {
	return d.Cmp(x) > 0
}

// GreaterThanEQ returns true if d is greater than or equal to x.
func (d *Decimal) GreaterThanEQ(x *Decimal) bool {
	return d.Cmp(x) >= 0
}

// Hypot returns Sqrt(p*p + q*q).
// Its precision is the minimum precision of p * 2 and q * 2.
func Hypot(p, q *Decimal) *Decimal {
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
	p0.ctx.Prec = min(p0.Prec(), q0.Prec())
	p0.Mul(p0, p0)
	q0.Mul(q0, q0)
	return p0.Sqrt(p0.Add(p0, q0))
}

// Int returns the integer component of d as a big.Int.
func (d *Decimal) Int() *big.Int {
	var (
		x    big.Int
		same bool
	)
	if d.compact != overflown {
		x.SetInt64(d.compact)
	} else {
		x = d.mantissa
		same = true
	}

	if d.scale <= 0 {
		return mulBigPow10(&x, -d.scale)
	}
	if same {
		return x.Set(&d.mantissa)
	}
	return &x
}

// Int64 returns the integer component of d as an int64.
// If the integer component cannot fit into an int64 the result is undefined.
func (d *Decimal) Int64() int64 {
	var x int64
	if d.compact != overflown {
		x = d.compact
	} else {
		x = d.mantissa.Int64()
	}

	if d.scale <= 0 {
		return mulPow10(x, -d.scale)
	}

	pow := pow10int64(d.scale)
	if pow == overflown {
		return overflown
	}
	return x / pow
}

// IsBig returns true if d is too large to fit into its
// integer member and can only be represented by a big.Int.
// (This means that a call to Int will result in undefined
// behavior.)
func (d *Decimal) IsBig() bool {
	return d.compact == overflown
}

// IsInt returns true if d can be represented exactly as an integer.
// (This means that the fractional part of the number is zero or the scale
// is <= 0.)
func (d *Decimal) IsInt() bool {
	if d.scale <= 0 {
		return true
	}
	_, frac := Modf(d)
	return frac.ez()
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
func (d *Decimal) LessThan(x *Decimal) bool {
	return d.Cmp(x) < 0
}

// LessThanEQ returns true if d is less than or equal to x.
func (d *Decimal) LessThanEQ(x *Decimal) bool {
	return d.Cmp(x) <= 0
}

// Lsh sets to d to x << n and returns d.
func (d *Decimal) Lsh(x *Decimal, n uint64) *Decimal {
	if x.compact != overflown {
		d.compact = x.compact << n
	} else {
		d.mantissa.Lsh(&x.mantissa, uint(n))
		d.compact = overflown
	}
	if d != x {
		d.ctx = x.ctx
		d.scale = x.scale
	}
	return d
}

// Log10 sets d the base-10 logarithm of x and returns d.
func (d *Decimal) Log10(x *Decimal) *Decimal {
	if x.ltez() {
		panic("decimal.Log10: x <= 0")
	}
	return d.Div(d.Log(x), ln10)
}

// Ilog10 returns the base-10 integer logarithm of d
// rounded up. (i.e., the number of digits in d)
func (d *Decimal) Ilog10() int64 {
	shift := int64(0)
	if d.scale < 0 {
		shift = -d.scale
	}
	if d.compact != overflown {
		x := d.compact
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
		// (From https://graphics.stanford.edu/~seander/bithacks.html)
		r := (((63 - clz(x) + 1) * 1233) >> 12) + shift
		if v := pow10int64(r); v == overflown || x < v {
			return r
		}
		return r + 1
	}
	// calculate with the big.Int mantissa.

	if d.mantissa.Sign() == 0 {
		return 1
	}

	// Accurate up to as high as we can report.
	r := (((int64(d.mantissa.BitLen()) + 1) * 0x268826A1) >> 31) + shift
	if cmpBigAbs(d.mantissa, bigPow10(r)) < 0 {
		return r
	}
	return r + 1
}

// Log2 sets d the base-2 logarithm of x and returns d.
func (d *Decimal) Log2(x *Decimal) *Decimal {
	if x.ltez() {
		panic("decimal.Log2: x <= 0")
	}
	return d.Div(d.Log(x), ln2)
}

func (d *Decimal) Log(x *Decimal) *Decimal {
	if x.ltez() {
		panic("decimal.Log: x <= 0")
	}

	mag := x.Ilog10() - x.scale - 1
	if mag < 3 {
		return d.logNewton(x)
	}
	root := d.integralRoot(x, mag)
	lnRoot := root.logNewton(root)
	return d.Mul(New(mag, 0), lnRoot)
}

// logNewton sets d to the natural logarithm of x
// using the Newtonian method and returns d.
func (d *Decimal) logNewton(x *Decimal) *Decimal {
	sp := d.Prec() + 1
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
	*d = *x0
	return d.SetPrec(sp - 1)
}

// exp sets d to e ** x and returns d.
func (d *Decimal) exp(x *Decimal) *Decimal {
	if x.ez() {
		p := d.Prec()
		d.SetInt64(1)
		d.ctx.Prec = p
		return d
	}

	if x.ltz() {
		x0 := new(Decimal).Set(x)
		// 1 / (e ** -x)
		return d.Div(New(1, 0), x0.exp(x0.Neg(x0)))
	}

	dec, frac := Modf(x)
	if dec.ez() {
		return d.taylor(x)
	}

	z := New(1, 0)
	z.ctx.Prec = d.Prec()
	z.Add(z, frac.Div(frac, dec))

	t := z.taylor(z)

	res := New(1, 0)
	for dec.GreaterThanEQ(dmaxInt64) {
		res.Mul(res, new(Decimal).powInt(t, math.MaxInt64))
		dec.Sub(dec, dmaxInt64)
	}
	return d.Mul(res, t.powInt(t, dec.Int64()))
}

// taylor sets d to e ** x using the Taylor series and returns d.
func (d *Decimal) taylor(x *Decimal) *Decimal {
	sum := new(Decimal).Add(New(1, 0), x).SetPrec(d.Prec())
	xp := new(Decimal).Set(x).SetPrec(d.Prec())
	fac := New(1, 0)

	var sp, term Decimal
	term.ctx.Prec = d.Prec()
	for i := int64(2); ; i++ {
		xp.Mul(xp, x).SetPrec(d.Prec())
		fac.Mul(fac, New(i, 0)).SetPrec(d.Prec())
		term.Div(xp, fac)
		sp.Set(sum)
		sum.Add(sum, &term)
		if sum.Equals(&sp) {
			break
		}
	}
	*d = *sum
	return d
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

// Mod sets d to the modulus x%y for y != 0 and returns d.
// If y == 0, a division-by-zero run-time panic occurs.
func (d *Decimal) Mod(x, y *Decimal) *Decimal {
	var q Decimal
	_, mod := q.DivMod(x, y, d)
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

// Mul sets d to x * y and returns d.
func (d *Decimal) Mul(x, y *Decimal) *Decimal {
	if x.compact != overflown {
		if y.compact != overflown {
			return d.mulCompact(x, y)
		}
		return d.mulHalf(x, y)
	}
	if y.compact != overflown {
		return d.mulHalf(y, x)
	}
	return d.mulBig(x, y)
}

func (d *Decimal) mulCompact(x, y *Decimal) *Decimal {
	prod := prod(x.compact, y.compact)
	if prod != overflown {
		d.compact = prod
	} else {
		d.mantissa.Mul(big.NewInt(x.compact), big.NewInt(y.compact))
		d.compact = overflown
	}
	d.scale = safeScale2(x.scale, sum(x.scale, y.scale))
	return d
}

// mulHalf multiplies a compact Decimal with a non-compact
// Decimal.
// Let the first arg be the compact and the second the non-compact.
func (d *Decimal) mulHalf(comp, nc *Decimal) *Decimal {
	if comp.compact == overflown {
		panic("decimal.Mul: (bug) comp should != overflown")
	}
	if comp.scale == nc.scale {
		d.scale = safeScale2(comp.scale, sum(comp.scale, nc.scale))
		d.mantissa.Mul(big.NewInt(comp.compact), &nc.mantissa)
		d.compact = overflown
		return d
	}
	return d.mulBig(&Decimal{
		mantissa: *big.NewInt(comp.compact),
		scale:    comp.scale,
	}, nc)
}

func (d *Decimal) mulBig(x, y *Decimal) *Decimal {
	d.scale = safeScale2(x.scale, sum(x.scale, y.scale))
	d.mantissa.Mul(&x.mantissa, &y.mantissa)
	d.compact = overflown
	return d
}

// MulRange sets d to the product of all integers in the range
// [a, b] inclusively and returns d.
// If a > b (empty range) the result is 1.
func (d *Decimal) MulRange(a, b int64) *Decimal {
	switch {
	case a > b:
		return d.SetInt64(1)
	case a <= 0 && b >= 0:
		return d.SetInt64(0)
	}
	neg := false
	if a < 0 {
		neg = (b-a)&1 == 0
		a, b = -b, -a
	}
	d.mulRange(a, b)
	if neg {
		d.Neg(d)
	}
	return d
}

func (d *Decimal) mulRange(a, b int64) *Decimal {
	switch {
	case a == 0:
		return d.SetInt64(0)
	case a > b:
		return d.SetInt64(1)
	case a == b:
		return d.SetInt64(a)
	case a+1 == b:
		return d.Mul(new(Decimal).SetInt64(a), new(Decimal).SetInt64(b))
	}
	m := (a + b) / 2
	return d.Mul(new(Decimal).mulRange(a, m), new(Decimal).mulRange(m+1, b))
}

// Neg sets d to -x and returns d.
func (d *Decimal) Neg(x *Decimal) *Decimal {
	if x.compact != overflown {
		d.compact = -x.compact
	} else {
		d.mantissa.Neg(&x.mantissa)
		d.compact = overflown
	}
	d.scale = x.scale
	return d
}

// or sets d to x | n and returns d.
func (d *Decimal) or(x *Decimal, n int64) *Decimal {
	if x.compact != overflown {
		d.compact = x.compact | n
	} else {
		d.mantissa.Or(&x.mantissa, big.NewInt(n))
		d.compact = overflown
	}
	// Save an assignment if d is x.
	if d != x {
		d.scale = x.scale
		d.ctx = x.ctx
	}
	return d
}

// IntDiv sets d to the quotient x/y for y != 0 and returns d. If y == 0, a
// division-by-zero run-time panic occurs. Quo implements truncated integer division.
func (d *Decimal) IntDiv(x, y *Decimal) *Decimal {
	if y.ez() {
		panic("decimal.IntDiv: division by zero")
	}

	if x.ez() {
		return New(0, 0)
	}

	if x.compact != overflown {
		if y.compact != overflown {
			d.scale = 0
			d.compact = x.compact / y.compact
			return d
		}
		return d.intDivBig(&Decimal{
			mantissa: *big.NewInt(x.compact),
			scale:    x.scale,
		}, y)
	}
	if y.compact != overflown {
		return d.intDivBig(x, &Decimal{
			mantissa: *big.NewInt(y.compact),
			scale:    y.scale,
		})
	}
	return d.intDivBig(x, y)
}

func (d *Decimal) intDivBig(x, y *Decimal) *Decimal {
	d.scale = 0
	d.mantissa.Quo(&x.mantissa, &y.mantissa)
	d.compact = overflown
	return d
}

// Rat returns the rational number representation of d.
// If x is non-nil, Rat stores the result in x instead of
// allocating a new Rat.
func (d *Decimal) Rat(x *big.Rat) *big.Rat {
	// TODO(eric):
	// We use big.Ints here when technically we could use our
	// int64 with big.Rat's SetInt64 method.
	// I'm not sure if it'll be an optimization or not.
	var num, denom big.Int
	if d.compact != overflown {
		c := big.NewInt(d.compact)
		if d.scale >= 0 {
			num.Set(c)
			denom.Set(mulBigPow10(oneInt, d.scale))
		} else {
			num.Set(mulBigPow10(c, -d.scale))
			denom.SetInt64(1)
		}
	} else {
		if d.scale >= 0 {
			num.Set(&d.mantissa)
			denom.Set(mulBigPow10(oneInt, d.scale))
		} else {
			num.Set(mulBigPow10(&d.mantissa, -d.scale))
			denom.SetInt64(1)
		}
	}
	if x != nil {
		return x.SetFrac(&num, &denom)
	}
	return new(big.Rat).SetFrac(&num, &denom)
}

// Rem sets d to the remainder of x/y and returns d.
func (d *Decimal) Rem(x, y *Decimal) *Decimal {
	if y.ez() {
		panic("decimal.Rem: division by zero")
	}

	if x.compact != overflown {
		if y.compact != overflown {
			d.scale = safeScale(x.compact, x.scale, sub(x.scale, y.scale))
			d.compact = x.compact % y.compact
			return d
		}
		return d.remBig(&Decimal{
			mantissa: *big.NewInt(x.compact),
			scale:    x.scale,
		}, y)
	}
	if y.compact != overflown {
		return d.remBig(x, &Decimal{
			mantissa: *big.NewInt(y.compact),
			scale:    y.scale,
		})
	}
	return d.remBig(x, y)
}

func (d *Decimal) remBig(x, y *Decimal) *Decimal {
	if x.scale == y.scale {
		d.scale = x.scale
		d.mantissa.Rem(&x.mantissa, &y.mantissa)
	} else {
		d.scale = safeScale(x.compact, x.scale, sub(x.scale, y.scale))
		d.mantissa.Rem(&x.mantissa, &y.mantissa)
	}
	d.compact = overflown
	return d
}

// rsh sets to d to x >> n and returns d.
// It's a raw right shift on x's integer component.
func (d *Decimal) Rsh(x *Decimal, n uint64) *Decimal {
	if x.compact != overflown {
		d.compact = x.compact >> n
	} else {
		d.mantissa.Rsh(&x.mantissa, uint(n))
		d.compact = overflown
	}
	if d != x {
		d.ctx = x.ctx
		d.scale = x.scale
	}
	return d
}

// Precision returns d's precision.
// This method does not necessarily return the value inside
// inside d's Context member.
func (d *Decimal) Prec() int64 {
	// Zero (default value) is assumed to mean
	// the default precision. This allows us to
	// still use new(Decimal).Div(x, y) and still
	// keep a reasonable amount of precision.
	//
	// Values < 0 are inferred to be zero. The only
	// value we manually set to be less than zero is
	// "overflown", which is set when somebody calls
	// SetPrec(0).
	switch p := d.ctx.Prec; {
	case p > 0:
		return p
	case p < 0:
		return 0
	default:
		return DefaultPrecision
	}
}

// Scale returns the scale component of d.
func (d *Decimal) Scale() int64 {
	return d.scale
}

// Set sets d to exactly x and returns d.
func (d *Decimal) Set(x *Decimal) *Decimal {
	if d != x {
		*d = Decimal{
			compact: x.compact,
			scale:   x.scale,
			ctx:     x.ctx,
		}
		if x.compact == overflown {
			d.mantissa.Set(&x.mantissa)
		}
	}
	return d
}

// SetBytes interprets buf as the bytes of a big-endian, usigned
// integer, sets d to that value, and returns d.
func (d *Decimal) SetBytes(buf []byte) *Decimal {
	d.mantissa.SetBytes(buf)
	d.compact = overflown
	return d
}

// SetContext sets d's context to ctx.
func (d *Decimal) SetContext(ctx Context) *Decimal {
	d.ctx = ctx
	return d
}

// SetInt64 sets d to x and returns d.
// As with SetInt64, it sets d's scale to 0.
func (d *Decimal) SetInt(x *big.Int) *Decimal {
	d.scale = 0
	d.compact = overflown
	d.mantissa.Set(x)
	return d
}

// SetInt64 sets d to x and returns d.
// It sets d's scale to 0.
func (d *Decimal) SetInt64(x int64) *Decimal {
	d.scale = 0
	d.compact = x
	d.mantissa.SetBits(nil)
	return d
}

// SetPrec sets d's precision to prec and returns a possibly
// rounded d.
//
// If you only want to set the Context's Prec member you should do that
// manually.
//
// A negative prec or calling this method on a Decimal with
// a negative scale is undefined behavior.
func (d *Decimal) SetPrec(prec int64) *Decimal {
	// Our undefined behavior.
	if d.scale < 0 || prec < 0 {
		return d
	}

	if prec == 0 {
		d.ctx.Prec = overflown
	} else {
		d.ctx.Prec = prec
	}

	switch {
	case d.scale == prec:
		return d
	case d.ez():
		d.SetInt64(0)
		d.scale = prec
		return d
	}

	shift := prec - d.scale
	if d.compact != overflown {

		res := d.compact
		// prec > d.scale
		if shift > 0 {
			d.compact = mulPow10(d.compact, shift)
		} else {
			d.compact /= pow10int64(-shift)
		}

		if d.compact != overflown {
			d.scale += shift
			return d
		}

		d.mantissa.SetInt64(res)
	}

	if shift > 0 {
		d.mantissa = *mulBigPow10(&d.mantissa, shift)
	} else {
		b := bigPow10(-shift)
		d.mantissa.Div(&d.mantissa, &b)
	}
	d.scale += shift
	return d
}

// ShiftRadix shifts d's radix (decimal point) n places, with a negative value meaning to
// the right and a positive value meaning to the left.
func (d *Decimal) ShiftRadix(n int64) *Decimal {
	scale := safeScale2(d.scale, sum(d.scale, n))
	if scale == overflown {
		panic("decimal.ShiftRadix: shift overflows scale")
	}
	d.scale = scale
	return d
}

func (d *Decimal) mulPow10(n int64) *Decimal {
	if d.compact != overflown {
		d.compact = mulPow10(d.compact, n)
		if d.compact != overflown {
			return d
		}
	}
	// d.compact == overflown
	d.mantissa = *mulBigPow10(&d.mantissa, n)
	return d
}

// Shrink shrinks d from a big.Int into its integer member
// if possible and returns d.
func (d *Decimal) Shrink() *Decimal {
	if d.compact == overflown &&
		(d.mantissa.Sign() > 0 && d.mantissa.Cmp(maxInt64) < 0 ||
			d.mantissa.Sign() < 0 && d.mantissa.Cmp(minInt64) > 0) {
		d.compact = d.mantissa.Int64()
		d.mantissa.SetBits(nil)
	}
	return d
}

// Sign returns:
//
//	-1 if d <  0
//	 0 if d == 0
//	+1 if d >  0
//
func (d *Decimal) Sign() int {
	if d.compact != overflown {
		if d.compact == 0 {
			return 0
		}
		if d.compact < 0 {
			return -1
		}
		return +1
	}
	return d.mantissa.Sign()
}

// Signbit returns true if d is negative.
func (d *Decimal) Signbit() bool {
	return d.Sign() < 0
}

// String returns the string representation of the decimal
// with the fixed point.
//
// Example:
//
//     d := New(-12345, 3)
//     println(d.String())
//
// Output:
//
//     -12.345
//
func (d *Decimal) String() string {
	if d == nil {
		return "<nil>"
	}
	return d.toString(true)
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
func (d *Decimal) StringFixed(places int64) string {
	return d.SetPrec(places).toString(false)
}

// Sqrt sets d to the square root of x and returns d.
// The precision of Sqrt is determined by d's Context.
// Sqrt will panic on negative values since Decimal cannot
// represent imaginary numbers.
func (d *Decimal) Sqrt(x *Decimal) *Decimal {

	if x.Sign() < 0 {
		panic("decimal.Sqrt: cannot take square root of negative number")
	}

	// Check if x is a perfect square. If it is, we can avoid having to
	// inflate x and can possibly use can use the hardware SQRT.
	// Note that we can only catch perfect squares that aren't big.Ints.
	if sq, ok := x.perfectSquare(); ok {
		d.compact = sq
		d.scale = 0
		d.ctx = x.ctx
		return d
	}
	// x isn't a perfect square or x is a big.Int

	n := new(Decimal).Set(x).ShiftRadix(-(d.Prec() << 1))
	n.SetInt(n.Int())
	ix := new(Decimal).Rsh(n, uint64((n.BitLen()+1)>>1))
	fmt.Println(ix.scale)

	var a, p Decimal
	for {
		p.Set(ix)
		ix.Add(ix, a.IntDiv(n, ix)).IntDiv(ix, two)
		if ix.Cmp(&p) == 0 {
			p := d.Prec()
			d.Set(ix)
			d.scale = p
			return d.SetPrec(p)
		}
	}
}

// perfectSquare algorithm slightly partially borrowed from
// http://stackoverflow.com/a/18686659/2967113
func (d *Decimal) perfectSquare() (square int64, ok bool) {
	if d.IsBig() {
		return
	}
	x0 := d.compact
	if x0 == 0 {
		return 0, true
	}
	tlz := uint64(ctz(x0))
	x0 >>= tlz
	if tlz&1 == 0 && (x0&7 != 1 || x0 == 0) {

		// "Show that floating point sqrt(x*x) >= x for all long x."
		// http://math.stackexchange.com/a/238885/153292
		tst := int64(math.Sqrt(float64(d.compact)))
		return tst, tst*tst == d.compact
	}
	return
}

// Sub sets d to x - y and returns d.
func (d *Decimal) Sub(x, y *Decimal) *Decimal {
	return d.Add(x, new(Decimal).Neg(y))
}

// Trunc sets d to x truncated down to n digits of precision
// (assuming x.scale > 0) and returns d
//
// Precision is the last digit that will not be truncated.
// If x'scale < 0, n < 0, or n > x's scale d will be set to x and no
// truncation will occur.
//
// Example:
//
//     NewFromString("123.456").Trunc(2).String() // output: "123.45"
//     NewFromString("123.456").Trun(0).String() // output: "123"
func (d *Decimal) Trunc(x *Decimal, n int64) *Decimal {
	if x.scale > 0 && n > 0 && n < x.scale {
		if x.compact != overflown {
			d.compact = x.compact / pow10int64(d.scale-n)
		} else {
			b := bigPow10(d.scale - n)
			d.mantissa.Div(&x.mantissa, &b)
		}
		d.scale = n
	}
	return d
}

func (d *Decimal) toString(trimTrailingZeros bool) string {
	// Fast path: return our value as-is.
	if d.scale == 0 {
		if d.compact == overflown {
			return d.mantissa.String()
		}
		return strconv.FormatInt(d.compact, 10)
	}

	var (
		str string
		neg bool
	)

	if d.compact == overflown {
		str = new(big.Int).Abs(&d.mantissa).String()
		neg = d.mantissa.Sign() < 0
	} else {
		abs := uint64(abs(d.compact))
		str = strconv.FormatUint(abs, 10)
		neg = d.compact < 0
	}

	if d.scale < 0 {
		if d.ez() {
			return "0"
		}
		if neg {
			return "-" + str + strings.Repeat("0", -int(d.scale))
		}
		return str + strings.Repeat("0", -int(d.scale))
	}

	var Int string
	switch p := int64(int64(len(str)) - int64(d.scale)); {
	case p == 0:
		Int = "0." + str
	case p > 0:
		Int = str[:p] + "." + str[p:]
	default:
		Int = "0." + strings.Repeat("0", -int(p)) + str
	}

	if trimTrailingZeros {
		i := len(Int) - 1
		for ; i >= 0 && Int[i] == '0'; i-- {
		}
		if Int[i] == '.' {
			i--
		}
		Int = Int[:i+1]
	}

	if neg {
		return "-" + Int
	}
	return Int
}
