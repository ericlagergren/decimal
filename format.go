package decimal

import (
	"fmt"
	"io"
	"math/big"
	"strconv"
)

// allZeros returns true if every character in b is '0'.
func allZeros(b []byte) bool {
	for _, c := range b {
		if c != '0' {
			return false
		}
	}
	return true
}

// roundString rounds the plain numeric string (e.g., "1234") b.
func roundString(b []byte, mode RoundingMode, pos bool, prec int) []byte {
	if prec >= len(b) {
		return b
	}

	// Trim zeros until prec. This is useful when we can round exactly by
	// simply chopping zeros off the end of the number.
	if allZeros(b[prec:]) {
		return b[:prec]
	}

	b = b[:prec+1]
	i := prec - 1

	// Blindly increment b[i] and handle possible carries later.
	switch mode {
	case AwayFromZero:
		b[i]++
	case ToZero:
		// OK
	case ToPositiveInf:
		if pos {
			b[i]++
		}
	case ToNegativeInf:
		if !pos {
			b[i]++
		}
	case ToNearestEven:
		if b[i+1] > '5' || b[i+1] == '5' && b[i]%2 != 0 {
			b[i]++
		}
	case ToNearestAway:
		if b[i+1] >= '5' {
			b[i]++
		}
	}

	if b[i] != '9'+1 {
		return b[:prec]
	}

	// We had to carry.
	b[i] = '0'

	for i--; i >= 0; i-- {
		if b[i] != '9' {
			b[i]++
			break
		}
		b[i] = '0'
	}

	// Carried all the way over to the first column, so slide the buffer down
	// and instead of reallocating.
	if b[0] == '0' {
		copy(b[1:], b)
		b[0] = '1'
		// We might end up with an extra digit of precision. E.g., given the
		// decimal 9.9 with a requested precision of 1, we'd convert 99 -> 10.
		// Let the calling code handle that case.
		prec++
	}
	return b[:prec]
}

// formatCompact formats the compact decimal, x, as an unsigned integer.
func formatCompact(x int64) []byte {
	if x < 0 {
		x = -x
	}
	var b [20]byte
	return strconv.AppendUint(b[0:0], uint64(x), 10)
}

// formatUnscaled formats the unscaled (non-compact) decimal, unscaled, as an
// unsigned integer.
func formatUnscaled(unscaled *big.Int) []byte {
	// math/big.MarshalText never returns an error, only nil, so there's no
	// need to check for an error. Use MarshalText instead of Append because it
	// limits us to one allocation.
	b, _ := unscaled.MarshalText()
	if b[0] == '-' {
		b = b[1:]
	}
	return b
}

const (
	// noWidth indicates the width of a formatted number wasn't set.
	noWidth = -1
	// noPrec indicates the precision of a formatted number wasn't set.
	noPrec = -1
)

const (
	normal = iota // either sci or plain, depending on x
	plain         // forced plain
	sci           // forced sci
)

type formatter struct {
	x *Big
	w interface {
		io.Writer
		io.ByteWriter
		WriteString(string) (int, error)
	}
	sign  byte  // leading '+' or ' ' flag
	prec  int   // total precision
	width int   // min width
	n     int64 // cumulative number of bytes written to w
}

func (f *formatter) WriteByte(c byte) error {
	f.n++
	return f.w.WriteByte(c)
}

func (f *formatter) WriteString(s string) (int, error) {
	m, err := f.w.WriteString(s)
	f.n += int64(m)
	return m, err
}

func (f *formatter) Write(p []byte) (n int, err error) {
	n, err = f.w.Write(p)
	f.n += int64(n)
	return n, err
}

var infs = [...]struct{ pinf, ninf string }{
	Go:  {pinf: "+Inf", ninf: "-Inf"},
	GDA: {pinf: "Infinity", ninf: "-Infinity"},
}

func (f *formatter) format(format, e byte) {
	// Special cases.
	if f.x == nil {
		f.WriteString("<nil>")
		return
	}

	x := f.x
	if x.form != finite {
		switch x.form {
		case zero:
			if f.width == noWidth {
				f.WriteByte('0')
			} else {
				f.WriteString("0.")
				io.CopyN(f, zeroReader{}, int64(f.width))
			}
		case pinf:
			f.WriteString(infs[x.Context.OperatingMode].pinf)
		case ninf:
			f.WriteString(infs[x.Context.OperatingMode].ninf)
		case snan:
			f.WriteString("sNaN")
		case qnan:
			f.WriteString("NaN")
		}
		return
	}

	neg := x.Signbit()
	if neg {
		f.WriteByte('-')
	} else if f.sign != 0 {
		f.WriteByte(f.sign)
	}

	var b []byte
	if x.isInflated() {
		b = formatUnscaled(&x.unscaled)
	} else {
		b = formatCompact(x.compact)
	}

	if f.prec > 0 {
		b = roundString(b, x.Context.RoundingMode, !neg, f.prec)
	}

	if format == plain {
		f.formatPlain(b)
		return
	}

	// "Next, the adjusted exponent is calculated; this is the exponent, plus
	// the number of characters in the converted coefficient, less one. That
	// is, exponent+(clength-1), where clength is the length of the coefficient
	// in decimal digits.
	adj := -int(x.scale) + (len(b) - 1)

	if format == normal {
		if x.scale <= 0 {
			f.Write(b)
			if x.scale < 0 {
				io.CopyN(f, zeroReader{}, -int64(x.scale))
			}
			return
		}

		// "If the exponent is less than or equal to zero and the
		// adjusted exponent is greater than or equal to -6...
		if x.scale >= 0 && adj >= -6 {
			// "...the number will be converted to a character
			// form without using exponential notation."
			//
			// - http://speleotrove.com/decimal/daconvs.html#reftostr
			f.formatPlain(b)
			return
		}
	}
	f.formatSci(b, adj, e)
}

// formatSci returns the scientific version of b.
func (f *formatter) formatSci(b []byte, adj int, e byte) {
	f.WriteByte(b[0])
	if len(b) > 1 {
		f.WriteByte('.')

		b = b[1:]
		if f.prec > len(b) {
			f.prec = len(b)
		}
		f.Write(b)
	}
	if adj != 0 {
		f.WriteByte(e)

		// If negative the following strconv.Append call will add the minus
		// sign for us.
		if adj > 0 {
			f.WriteByte('+')
		}
		f.WriteString(strconv.Itoa(adj))
	}
}

// formatPlain returns the plain string version of b.
func (f *formatter) formatPlain(b []byte) {
	const zeroRadix = "0."

	scale := int(f.x.scale)
	switch radix := len(b) - scale; {
	// log10(b) == scale, so immediately before b.
	case radix == 0:
		f.WriteString(zeroRadix)
		if i := trimIndex(b); i > 0 {
			b = b[:i]
		}
		f.Write(b)

	// log10(b) > scale, so somewhere inside b.
	case radix > 0:
		f.Write(b[:radix])
		if i := trimIndex(b[radix:]); i > 0 {
			f.WriteByte('.')
			f.Write(b[radix : radix+i])
		}

	// log10(b) < scale, so before p "0s" and before b.
	default:
		f.WriteString(zeroRadix)
		io.CopyN(f, zeroReader{}, -int64(radix))

		end := len(b)
		if f.prec >= 0 && f.prec < end {
			end = f.prec
		}
		f.Write(b[:end])
	}
}

// TODO: can we merge zeroReader and spaceReader into a "singleReader" or
// something and still maintain the same performance?

// zeroReader is an io.Reader that, when read from, only provides the character
// '0'.
type zeroReader struct{}

// Read implements io.Reader.
func (z zeroReader) Read(p []byte) (n int, err error) {
	// zeroLiterals is 16 '0' bytes. It's used to speed up zeroReader's Read
	// method.
	const zeroLiterals = "0000000000000000"
	for n < len(p) {
		m := copy(p[n:], zeroLiterals)
		if m == 0 {
			break
		}
		n += m
	}
	return n, nil
}

// spaceReader is an io.Reader that, when read from, only provides the
// character ' '.
type spaceReader struct{}

// Read implements io.Reader.
func (s spaceReader) Read(p []byte) (n int, err error) {
	// spaceLiterals is 16 ' ' bytes. It's used to speed up spaceReader's Read
	// method.
	const spaceLiterals = "                "
	for n < len(p) {
		m := copy(p[n:], spaceLiterals)
		if m == 0 {
			break
		}
		n += m
	}
	return n, nil
}

// trimIndex returns the index in b where b should be trimmed to remove
// trailing '0's.
func trimIndex(b []byte) int {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != '0' {
			return i + 1
		}
	}
	return -1
}

// stateWrapper is a wrapper around an io.Writer to add WriteByte and
// WriteString methods.
type stateWrapper struct{ fmt.State }

func (w stateWrapper) WriteByte(c byte) error {
	_, err := w.Write([]byte{c})
	return err
}

func (w stateWrapper) WriteString(s string) (int, error) {
	return io.WriteString(w.State, s)
}
