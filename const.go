package decimal

import "math"

// Precision and scale limits.
const (
	MaxScale = math.MaxInt32 // smallest allowed scale.
	MinScale = math.MinInt32 // largest allowed scale.

	MinPrec = 0             // smallest allowed Context precision.
	MaxPrec = math.MaxInt32 // largest allowed Context precision.
)

// DefaultPrec is the default precision used if the Context's
// 'Prec' member is 0.
const DefaultPrec = 16

// Context tells the arithmetic operations how to do their jobs.
//
// Prec is the maximum number of digits that should trail
// the radix during a potentially lossy (e.g., division) operation.
// The Decimal's precision will only be less than Prec if
// the operation has a finite expansion less than Prec.
//
// Mode instructs lossy operations how to round.
type Context struct {
	Prec int32
	Mode RoundingMode
}

// The following are called ContextXX instead of DecimalXX
// to reserve the DecimalXX namespace for future decimal types.

var (
	// Contex32 is the IEEE 754R Decimal32 format.
	Contex32 = Context{Prec: 7, Mode: ToNearestEven}

	// Context64 is the IEEE 754R Decimal64 format.
	Context64 = Context{Prec: 16, Mode: ToNearestEven}

	// Context128 is the IEEE 754R Decimal128 format.
	Context128 = Context{Prec: 34, Mode: ToNearestEven}
)

// RoundingMode determines how a Decimal will be rounded
// if the exact result cannot accurately be represented.
type RoundingMode byte

// The following rounding modes are supported.
const (
	ToNearestEven RoundingMode = iota // == IEEE 754-2008 roundTiesToEven
	ToNearestAway                     // == IEEE 754-2008 roundTiesToAway
	ToZero                            // == IEEE 754-2008 roundTowardZero
	AwayFromZero                      // no IEEE 754-2008 equivalent
	ToNegativeInf                     // == IEEE 754-2008 roundTowardNegative
	ToPositiveInf                     // == IEEE 754-2008 roundTowardPositive

	// Unneeded means finite decimal expansion. Will panic if this
	// RoundingMode is provided and the lossy operation does not have a finite
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
	case ToPositiveInf:
		return pos
	case ToNegativeInf:
		return !pos
	case ToNearestEven, ToNearestAway:
		switch {
		case c < 0:
			return false
		case c > 0:
			return true
		case c == 0:
			if r == ToNearestEven {
				return odd
			}
			return true
		}
	}
	panic("unknown RoundingMode")
}
