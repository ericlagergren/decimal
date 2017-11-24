package decimal

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"github.com/ericlagergren/decimal/internal/c"
)

func (z *Big) scan(r io.ByteScanner) error {
	if debug {
		defer func() { z.validate() }()
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

	// Sign
	neg, err := scanSign(r)
	if err != nil {
		return err
	}

	z.form, err = z.scanForm(r)
	if err != nil {
		if err == strconv.ErrSyntax {
			z.Context.Conditions |= ConversionSyntax
		}
		return err
	}

	if z.form&special != 0 {
		if neg {
			z.form |= signbit
		}
		return nil
	}

	// Mantissa (as a unsigned integer)
	if err := z.scanMant(r); err != nil {
		switch err {
		case io.EOF:
			z.form = qnan
			return io.ErrUnexpectedEOF
		case strconv.ErrSyntax:
			z.form = qnan
			z.Context.Conditions |= ConversionSyntax
		// Can only overflow
		case Overflow:
			z.xflow(true, neg)
		}
		return nil
	}

	// Exponent
	if err := z.scanExponent(r); err != nil && err != io.EOF {
		switch err {
		case Underflow:
			z.xflow(false, neg)
		case Overflow:
			z.xflow(true, neg)
		case strconv.ErrSyntax:
			z.form = qnan
			z.Context.Conditions |= ConversionSyntax
		default:
			return err
		}
		return nil
	}

	// Adjust for negative values.
	if neg {
		z.form |= signbit
	}
	return nil
}

func scanSign(r io.ByteScanner) (bool, error) {
	ch, err := r.ReadByte()
	if err != nil {
		return false, err
	}
	switch ch {
	case '+':
		return false, nil
	case '-':
		return true, nil
	default:
		return false, r.UnreadByte()
	}
}

func (z *Big) scanForm(r io.ByteScanner) (form, error) {
	ch, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	if (ch >= '0' && ch <= '9') || ch == '.' {
		return finite, r.UnreadByte()
	}

	// Likely infinity.
	if ch == 'i' || ch == 'I' {
		const (
			inf1 = "infinity"
			inf2 = "INFINITY"
		)
		for i := 1; i < len(inf1); i++ {
			ch, err = r.ReadByte()
			if err != nil {
				if err == io.EOF {
					// "inf"
					if i == len("inf") {
						return inf, nil
					}
					return 0, io.ErrUnexpectedEOF
				}
				return 0, err
			}
			if ch != inf1[i] && ch != inf2[i] {
				return 0, strconv.ErrSyntax
			}
		}
		return inf, nil
	}

	i := 0
	signal := false
	switch ch {
	case 'q', 'Q':
		// OK
	case 's', 'S':
		signal = true
	case 'n', 'N':
		i = 1 // or r.UnreadByte() and don't use i.
	default:
		return 0, strconv.ErrSyntax
	}

	const (
		nan1 = "nan"
		nan2 = "NAN"
	)
	for ; i < len(nan1); i++ {
		ch, err = r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return 0, io.ErrUnexpectedEOF
			}
			return 0, err
		}
		if ch != nan1[i] && ch != nan2[i] {
			return 0, strconv.ErrSyntax
		}
	}

	// Parse payload
	var buf [20]byte
	for i = 0; i < 20; i++ {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		if ch >= '0' && ch <= '9' {
			buf[i] = ch
		}
	}
	if i > 0 {
		z.compact, err = strconv.ParseUint(string(buf[:i]), 10, 64)
		if err != nil {
			return 0, err
		}
	}

	if signal {
		return snan, nil
	}
	return qnan, nil
}

// fakeState is implements fmt.ScanState so we can call big.Int.Scan.
type fakeState struct {
	length int
	scale  int
	i      int    // index into small
	small  []byte // buffer of first 20 or so characters.
	r      io.ByteScanner
}

func (f *fakeState) ReadRune() (rune, int, error) {
	// small is guaranteed to be a valid numeric character.
	if f.i < len(f.small) {
		r := rune(f.small[f.i])
		f.i++
		f.length++
		return r, 1, nil
	}

	ch, err := f.r.ReadByte()
	if err != nil {
		return 0, 0, err
	}

	if ch >= '0' && ch <= '9' {
		f.length++
		return rune(ch), 1, nil
	}
	if ch == '.' {
		const noScale = -1
		if f.scale > noScale {
			return 0, 0, strconv.ErrSyntax
		}
		f.scale = f.length
		return f.ReadRune() // skip to next raracter
	}
	if ch == 'e' || ch == 'E' {
		// Can simply UnreadByte here since we're not using small anymore.
		if err := f.r.UnreadByte(); err != nil {
			return 0, 0, err
		}
		return 0, 0, io.EOF // end of mantissa
	}
	return 0, 0, strconv.ErrSyntax
}

func (f *fakeState) UnreadRune() error {
	if f.i < len(f.small) {
		if f.i == 0 {
			return errors.New("attempted to read before start of buffer")
		}
		f.i--
		return nil
	}
	return f.r.UnreadByte()
}

func (f *fakeState) SkipSpace() {
	for {
		ch, err := f.r.ReadByte()
		if err != nil {
			return
		}
		if !unicode.IsSpace(rune(ch)) {
			f.r.UnreadByte()
			return
		}
	}
}

func (f *fakeState) Token(skipSpace bool, fn func(rune) bool) (token []byte, err error) {
	if skipSpace {
		f.SkipSpace()
	}
	if fn == nil {
		fn = func(r rune) bool { return !unicode.IsSpace(r) }
	}
	for {
		r, _, err := f.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if fn(r) {
			token = append(token, byte(r))
		}
	}
	return token, nil
}

func (f *fakeState) Width() (int, bool) { return 0, false }

func (f *fakeState) Read(_ []byte) (int, error) {
	return 0, errors.New("bad scanning routine")
}

func (z *Big) scanMant(r io.ByteScanner) (err error) {
	const noScale = -1

	// Scan the first 21 or fewer bytes into a buffer. Should we hit io.EOF
	// sooner, we know to try to parse it as a uint64. Otherwise, we read from
	// small—followed by our io.ByteScanner—into z.uncaled.
	var (
		small  [20 + 1]byte
		scale  int = noScale
		length int
		i      int
	)

Loop:
	for ; i < len(small); i++ {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				// Hit the end of our input: we're done here.
				break
			}
			return err
		}

		// Common case.
		if ch >= '0' && ch <= '9' {
			small[i] = ch
			length++
			continue
		}

		switch ch {
		case '.':
			if scale != noScale { // found two '.'s
				return strconv.ErrSyntax
			}
			scale = i
			i--
		case 'e', 'E':
			// Hit the exponent: we're done here.
			if err := r.UnreadByte(); err != nil {
				return err
			}
			break Loop
		default:
			return strconv.ErrSyntax
		}
	}

	// We can tentatively fit into a uint64 if we didn't fill the buffer.
	if i < len(small) {
		z.compact, err = strconv.ParseUint(string(small[:i]), 10, 64)
		if err != nil {
			err = err.(*strconv.NumError).Err
			if err == strconv.ErrSyntax {
				return err
			}
			// else strconv.ErrRange
		}
	}

	// Either we filled the buffer or we hit the edge case where len(s) == 19
	// but it's too large to fit into an int64.
	if i >= len(small) || (err == strconv.ErrRange && i == len(small)-1) {
		err = nil
		fs := fakeState{
			small:  small[:i],
			r:      r,
			scale:  scale,
			length: -1,
		}
		if err := z.unscaled.Scan(&fs, 'd'); err != nil {
			return err
		}

		z.compact = c.Inflated
		if z.unscaled.Sign() == 0 {
			z.compact = 0
		}

		if scale == noScale {
			scale = fs.scale
		}
		length = fs.length
	}

	if scale != noScale {
		z.exp = -int(length - scale)
	}

	// Ordinarily we'd set the precision here, but if we have numbers with
	// leading zeros we'll over-estimate the length.
	// z.precision = int(length)

	// Ideally we'd handle this manually, _but_ we run into an issue where
	// leading zeros cause our mantissa to appear to be larger than the cutoff
	// of 20 digits. In reality, when we convert the string to an integer the
	// leading zeros are ignored. But the bookkeeping to skip the leading zeros
	// is too much effort and has a non-negligible overhead.
	z.norm()
	return err
}

func (z *Big) scanExponent(r io.ByteScanner) error {
	ch, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch ch {
	case 'e', 'E':
		// OK
	default:
		return strconv.ErrSyntax
	}

	var buf [20]byte // max length of a signed 64-bit int, including sign.
	var i int
	for ; i < len(buf); i++ {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		buf[i] = ch
	}

	if _, err := r.ReadByte(); err != io.EOF {
		// TODO(eric): not _always_ over/underflow e.g. if the next character
		// isn't numeric.
		if buf[0] == '-' {
			return Underflow
		}
		return Overflow
	}

	exp, err := strconv.Atoi(string(buf[:i]))
	if err != nil {
		err = err.(*strconv.NumError).Err
		if err == strconv.ErrRange {
			// exp is set to the max int if it overflowed 32 bits, negative if
			// it underflowed.
			if exp > 0 {
				return Underflow
			}
			return Overflow
		}
		return err
	}

	z.exp += exp
	return nil
}

// byteReader implementation borrowed from math/big/intconv.go

// byteReader is a local wrapper around fmt.ScanState; it implements the
// io.ByteReader interface.
type byteReader struct {
	fmt.ScanState
}

func (r byteReader) ReadByte() (byte, error) {
	ch, size, err := r.ReadRune()
	if size != 1 && err == nil {
		err = fmt.Errorf("invalid rune %#U", ch)
	}
	return byte(ch), err
}

func (r byteReader) UnreadByte() error {
	return r.UnreadRune()
}
