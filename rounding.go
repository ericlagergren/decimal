package decimal

// DefaultPrecision is the default precision used if the Context's
// 'Prec' member is 0.
const DefaultPrecision = 16

// Context tells the arithmetic operations how to do their jobs.
//
// Prec is the maximum number of digits that should trail
// the radix during a potentially lossy (e.g., division) operation.
// The Decimal's precision will only be less than Prec if
// the operation has a finite expansion less than Prec.
//
// Mode instructs lossy operations how to round.
type Context struct {
	Prec int64
	Mode RoundingMode
}

var (
	// Decimal32 is the IEEE 754R Decimal32 format.
	Decimal32 = Context{Prec: 7, Mode: HalfEven}

	// Decimal64 is the IEEE 754R Decimal64 format.
	Decimal64 = Context{Prec: 16, Mode: HalfEven}

	// Decimal128 is the IEEE 754R Decimal128 format.
	Decimal128 = Context{Prec: 34, Mode: HalfEven}
)

// RoundingMode determines how a Decimal will be rounded
// if the exact result cannot accurately be represented.
type RoundingMode byte

// The following rounding modes are supported.
const (
	HalfEven     RoundingMode = iota // Bankers'/Gaussian rounding.
	AwayFromZero                     // Round away from zero.
	ToZero                           // Round towards zero.
	PositiveInf                      // Round to +Inf.
	NegativeInf                      // Round to -Inf.

	// Nearest neighbor rounding. AwayFromZero if equidistant.
	HalfUp

	// Nearest neighbor rounding. ToZero if equidistant.
	HalfDown

	// Finite decimal expansion. Will panic if this RoundingMode
	// is provided and the lossy operation does not have a finite
	// decimal expansion.
	Unneeded
)

//go:generate stringer -type RoundingMode

func (r RoundingMode) needsInc(c int, pos, odd bool) bool {
	switch r {
	case Unneeded:
		panic("decimal: rounding is necessary")
	case AwayFromZero:
		return true
	case ToZero:
		return false
	case PositiveInf:
		return pos
	case NegativeInf:
		return !pos
	case HalfEven, HalfUp, HalfDown:
		switch {
		case c < 0:
			return false
		case c > 0:
			return true
		case c == 0:
			switch r {
			case HalfEven:
				return odd
			case HalfUp:
				return true
			case HalfDown:
				return false
			}
		}
	}
	panic("unreachable")
}
