package decimal

import (
	"encoding/binary"
	"fmt"
	"math"
)

// decomposer composes or decomposes a decimal value to and from its individual
// parts.
//
// There are four separate parts:
//
//    1. boolean negative flag
//    2. form byte with three possible states (finite=0, infinite=1, NaN=2)
//    3. base-2 big-endian integer coefficient (also known as a significand) byte
//       slice
//    4. int32 exponent
//
// These are composed into a final value where
//
//    decimal = (neg) (form=finite) coefficient * 10^exponent
//
// A coefficient with a length of zero represents zero.
//
// If the form is not finite the coefficient and scale should be ignored.
//
// The negative parameter may be set to true for any form, although
// implementations are not required to respect the negative parameter in the
// non-finite form.
//
// Implementations may choose to signal a negative zero or negative NaN, but
// implementations that do not support these may also ignore the negative zero
// or negative NaN without error.
//
// If an implementation does not support Infinity it may be converted into a NaN
// without error.
//
// If a value is set that is larger then what is supported by an implementation
// is attempted to be set, an error must be returned.
//
// Implementations must return an error if a NaN or Infinity is attempted to be
// set while neither are supported.
type decomposer interface {
	// Decompose decomposes the decimal into parts.
	//
	// If the provided buf has sufficient capacity, buf may be returned as the
	// coefficient with the value set and length set as appropriate.
	Decompose(buf []byte) (form byte, neg bool, coeff []byte, exp int32)

	// Compose composes the decimal value from its parts.
	//
	// If the value cannot be represented, then an error should be returned.
	//
	// The coefficent should not be modified. Successive calls to compose with
	// the same arguments should result in the same decimal value.
	Compose(form byte, neg bool, coeff []byte, exp int32) error
}

// Decompose decomposes x into individual parts.
//
// If the provided buf has sufficient capacity, buf may be returned as the
// coefficient with the value and length set as appropriate.
//
// It implements the hidden decimalDecompose in the database/sql/driver package.
func (x *Big) Decompose(buf []byte) (form byte, neg bool, coeff []byte, exp int32) {
	neg = x.Sign() < 0
	switch {
	case x.IsInf(0):
		form = 1
		return
	case x.IsNaN(0):
		form = 2
		return
	}
	if !x.IsFinite() {
		panic("expected number to be finite")
	}
	if x.exp > math.MaxInt32 {
		panic("exponent exceeds max size")
	}
	exp = int32(x.exp)

	if x.isCompact() {
		if cap(buf) >= 8 {
			coeff = buf[:8]
		} else {
			coeff = make([]byte, 8)
		}
		binary.BigEndian.PutUint64(coeff, x.compact)
	} else {
		coeff = x.unscaled.Bytes() // This returns a big-endian slice.
	}
	return
}

// Compose sets z to the decimal value composed from individual parts.
//
// It implements the hidden decimalCompose interface in the database/sql package.
func (z *Big) Compose(form byte, neg bool, coeff []byte, exp int32) error {
	switch form {
	case 0:
		z.unscaled.SetBytes(coeff)
		z.SetBigMantScale(&z.unscaled, -int(exp))
		if neg {
			z.Neg(z)
		}
		return nil
	case 1:
		z.SetInf(neg)
		return nil
	case 2:
		z.SetNaN(false)
		return nil
	default:
		return fmt.Errorf("unknown form: %v", form)
	}
}
