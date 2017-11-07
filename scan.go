package decimal

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"github.com/ericlagergren/decimal/internal/arith/checked"
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

	// Special values (inf, nan, ...)
	form, err := scanSpecial(r)
	if err != nil {
		return err
	}
	if form != 0 { // explicit 0, but yes, != zero would work
		if neg {
			form |= sign
		}
		z.form = form
		return nil
	}

	// Assume we're finite at this point unless scanMant tells us otherwise.
	z.form = finite

	// Mantissa (as a unsigned integer)
	if err := z.scanMant(r); err != nil {
		if err == strconv.ErrSyntax {
			z.form = qnan
			z.Signal(ConversionSyntax, err)
			return err
		}
		// Can only overflow
		if err == errOverflow {
			z.xflow(true, neg)
			return err
		}
		return err
	}

	// Exponent
	ch, err := r.ReadByte()
	if err == nil {
		switch ch {
		case 'e', 'E':
			switch z.scanExponent(r) {
			case nil:
				// OK
			case errUnderflow:
				z.xflow(false, neg)
				return err
			case errOverflow:
				z.xflow(true, neg)
				return err
			}
		default:
			z.form = qnan
			z.Signal(ConversionSyntax, strconv.ErrSyntax)
			return strconv.ErrSyntax
		}
	} else if err != io.EOF {
		return err
	}

	// Adjust for negative values.
	if neg {
		if z.IsFinite() {
			if z.isCompact() {
				z.compact = -z.compact
			} else {
				z.unscaled.Neg(&z.unscaled)
			}
		} else {
			z.form |= sign
		}
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

func scanSpecial(r io.ByteScanner) (form, error) {
	ch, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	if ch >= '0' && ch <= '9' {
		return 0, r.UnreadByte()
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
	// Ignore trailing diagnostic digits, if any.
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
	small  []byte // buffer of first 19 or 20 characters.
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

	r := rune(ch)
	if r >= '0' && r <= '9' {
		f.length++
		return r, 1, nil
	}
	if r == '.' {
		const noScale = -1
		if f.scale > noScale {
			return 0, 0, strconv.ErrSyntax
		}
		f.scale = f.length
		return f.ReadRune() // skip to next raracter
	}
	if r == 'e' || r == 'E' {
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

	// Scan the first 20 or fewer bytes into a buffer. Should we hit io.EOF
	// sooner, we know to try to parse it as an int64. Otherwise, we read from
	// small—followed by r—into z.uncaled.
	var (
		small  [20]byte
		scale  int = noScale
		length int
		i      int
	)

	for ; i < len(small); i++ {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if ch >= '0' && ch <= '9' {
			small[i] = ch
			length++
		} else if ch == '.' {
			if scale > noScale {
				return strconv.ErrSyntax
			}
			scale = i
			i--
		} else if ch == 'e' || ch == 'E' {
			if err := r.UnreadByte(); err != nil {
				return err
			}
			break
		} else {
			return strconv.ErrSyntax
		}
	}

	// We can tentatively fit into an int64 if we didn't fill the buffer.
	if i < len(small) {
		z.compact, err = strconv.ParseInt(string(small[:i]), 10, 64)
		if err != nil {
			err = err.(*strconv.NumError).Err
			if err == strconv.ErrSyntax {
				return err
			}
			// strconv.ErrRange
		} else if z.compact == 0 {
			z.form = zero
		}
	}

	// Either we filled the buffer or we hit the edge case where len(s) == 19
	// but it's too large to fit into an int64.
	if i >= len(small) || (err == strconv.ErrRange && i == len(small)-1) {
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
			z.form = zero
		}

		if scale == noScale {
			scale = fs.scale
		}
		length = fs.length
		err = nil
	}

	if scale > noScale {
		z.scale = int32(length - scale)
	}

	// Ideally we'd handle this manually, _but_ we run into an issue where
	// leading zeros cause our mantissa to appear to be larger than the cutoff
	// of 19 digits. In reality, when we convert the string to an integer the
	// leading zeros are ignored. But the bookkeeping to skip the leading zeros
	// is too much effort and has a non-negligible overhead.
	z.norm()
	return err
}

func (z *Big) scanExponent(r io.ByteScanner) error {
	var buf [11]byte // max length of a signed 32-bit int, including sign.
	var i int
	for i = range buf {
		ch, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		buf[i] = ch
	}

	exp, err := strconv.ParseInt(string(buf[:i]), 10, 32)
	if err != nil {
		err = err.(*strconv.NumError).Err
		if err == strconv.ErrRange {
			// exp is set to the max int if it overflowed 32 bits, negative if
			// it underflowed.
			if exp > 0 {
				return errOverflow
			}
			return errUnderflow
		}
		return err
	}

	scale, ok := checked.Sub32(z.scale, int32(exp))
	if !ok {
		// x + -y ∈ [-1<<31, 1<<31-1]
		if z.scale > 0 {
			return errOverflow
		}
		return errUnderflow
	}
	z.scale = scale
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
