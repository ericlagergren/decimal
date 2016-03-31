package decimal

import (
	"bytes"
	"math/big"
	"strconv"
	"strings"

	"github.com/EricLagergren/decimal/internal/c"
	"github.com/EricLagergren/decimal/internal/checked"
)

// Big represents a fixed-point, multi-precision
// decimal number.
//
// A Big decimal is an arbitrary-precision number and a
// scale; the latter representing the number of digits to the
// right of the radix.
//
// A negative scale indicates the lack of a radix (typically a
// very large number).
type Big struct {
	// If |v| <= math.MaxInt64 then the mantissa will be stored
	// in this field...
	compact int64
	scale   int32
	ctx     *int // TODO
	form    form

	// ...otherwise, it's held here.
	mantissa big.Int
}

type form byte

const (
	norm form = iota
	inf
	nan
)

func (z *Big) isInflated() bool {
	return z.compact == c.Inflated
}

func (z *Big) isCompact() bool {
	return z.compact != c.Inflated
}

// New creates a new Big decimal with the given value and scale.
func New(value int64, scale int32) *Big {
	return &Big{
		compact: value,
		scale:   scale,
	}
}

// Add sets z to x + y and returns z.
func (z *Big) Add(x, y *Big) *Big {
	if x.isCompact() {
		if x.isCompact() {
			return z.addCompact(x, y)
		}
		return z.addHalf(x, y)
	}
	if y.isCompact() {
		return z.addHalf(y, x)
	}
	return z.addBig(x, y)
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
		} else {
			z.mantissa.Add(big.NewInt(x.compact), big.NewInt(y.compact))
			z.compact = c.Inflated
		}
		return z
	}

	// Guess the scales. We need to inflate lo.
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	// Power of 10 we need to multiple our lo value by in order
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
	z.mantissa.Add(scaled, big.NewInt(hi.compact))
	z.compact = c.Inflated
	return z
}

// addHalf adds a compact Big with a non-compact Big.
// addHalf will panic if the first argument is not compact.
func (z *Big) addHalf(comp, non *Big) *Big {
	if comp.isInflated() {
		panic("decimal.Add (bug) comp.isInflated() == true")
	}
	if comp.scale == non.scale {
		z.mantissa.Add(big.NewInt(comp.compact), &non.mantissa)
		z.scale = comp.scale
		z.compact = c.Inflated
		return z
	}
	// Since we have to rescale we need to add two big.Ints
	// together because big.Int doesn't have an API for
	// increasing its value by an integer.
	return z.addBig(&Big{
		mantissa: *big.NewInt(comp.compact),
		scale:    comp.scale,
	}, non)
}

func (z *Big) addBig(x, y *Big) *Big {
	hi, lo := x, y
	if hi.scale < lo.scale {
		hi, lo = lo, hi
	}

	inc := hi.scale - lo.scale
	scaled := checked.MulBigPow10(&lo.mantissa, inc)
	z.mantissa.Add(&hi.mantissa, scaled)
	z.compact = c.Inflated
	z.scale = hi.scale
	return z
}

// SetString sets z to the value of s, returning z and a bool
// indicating success. s must be a decimal number of the same format
// accepted by Parse, with base argument 0.
func (z *Big) SetString(s string) (*Big, bool) {
	var scale int32

	i := strings.IndexAny(s, "Ee")
	if i != -1 {
		eint, err := strconv.ParseInt(s[i+1:], 10, 32)
		if err != nil {
			return nil, false
		}
		if eint == c.BadScale {
			panic("decimal.NewFromString: scale is too small")
		}
		s = s[:i]
		scale = -int32(eint)
	}

	str := s
	parts := strings.Split(s, ".")
	if pl := len(parts); pl == 2 {
		str = parts[0] + parts[1]
		scale += int32(len(parts[1]))
	} else if pl != 1 {
		return nil, false
	}

	// TODO: Do we need the < 0 check?
	if scale < 0 || scale > c.BadScale {
		return nil, false
	}

	var val big.Int
	_, ok := val.SetString(str, 10)
	if !ok {
		return nil, false
	}
	z.compact = c.Inflated
	z.scale = scale
	z.mantissa = val
	return z.Shrink(), true
}

// Neg sets z to -x and returns z.
func (z *Big) Neg(x *Big) *Big {
	if x.isCompact() {
		z.compact = -x.compact
	} else {
		z.mantissa.Neg(&x.mantissa)
		z.compact = c.Inflated
	}
	z.scale = x.scale
	return z
}

// Shrink shrinks d from a big.Int into an int64 if possible
// and returns z.
func (z *Big) Shrink() *Big {
	if z.isInflated() {
		sign := z.mantissa.Sign()
		if (sign > 0 && z.mantissa.Cmp(c.MaxInt64) < 0) ||
			(sign < 0 && z.mantissa.Cmp(c.MinInt64) > 0) ||
			sign == 0 {
			z.compact = z.mantissa.Int64()
			z.mantissa.SetBits(nil)
		}
	}
	return z
}

// String returns the string representation of z.
// For special cases, if z == nil returns "<nil>"
// and if IsNaN(z) returns "NaN"
func (z *Big) String() string {
	if z == nil {
		return "<nil>"
	}
	// If IsNaN(z) {
	// 	return "NaN"
	// }
	return z.toString(trimZeros | plain)
}

// strOpts are ORd together.
type strOpts uint8

const (
	trimZeros strOpts = 1 << iota
	plain
	scientific
)

func (z *Big) toString(opts strOpts) string {
	// Fast path: return our value as-is.
	if z.scale == 0 {
		if z.isInflated() {
			return z.mantissa.String()
		}
		return strconv.FormatInt(z.compact, 10)
	}

	// TODO: ez method
	// We check for z.scale < 0 && z.ez above because it saves
	// us an allocation of a bytes.Buffer
	if z.scale < 0 && opts&trimZeros != 0 { //&& z.ez() {
		return "0"
	}

	var (
		str string
		neg bool
		b   bytes.Buffer
	)

	if z.isInflated() {
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

	// Just mantissa + z.scale "0"s -- no radix.
	if z.scale < 0 {
		b.WriteString(str)
		b.Write(bytes.Repeat([]byte{'0'}, -int(z.scale)))
		return b.String()
	}

	// Determine where to place the radix.
	switch p := int32(len(str)) - z.scale; {

	// log10(mantissa) == scale, so immediately before str.
	case p == 0:
		b.WriteString("0.")
		b.WriteString(str)

	// log10(mantissa) > scale, so somewhere inside str.
	case p > 0:
		b.WriteString(str[:p])
		b.WriteByte('.')
		b.WriteString(str[p:])

	// log10(mantissa) < scale, so before p "0s" and before str.
	default:
		b.WriteString("0.")
		b.Write(bytes.Repeat([]byte{'0'}, -int(p)))
		b.WriteString(str)
	}

	if opts&trimZeros != 0 {
		buf := b.Bytes()
		i := len(buf) - 1
		for ; i >= 0 && buf[i] == '0'; i-- {
		}
		if buf[i] == '.' {
			i--
		}
		b.Truncate(i + 1)
	}
	return b.String()
}

// Sub sets z to x - y and returns z.
func (z *Big) Sub(x, y *Big) *Big {
	return z.Add(x, new(Big).Neg(y))
}
