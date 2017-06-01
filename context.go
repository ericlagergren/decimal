package decimal

import (
	"math"
	"math/big"
)

// Precision and scale limits.
const (
	MaxScale = math.MaxInt32 // largest allowed scale.
	MinScale = math.MinInt32 // smallest allowed scale.

	MaxPrec = math.MaxInt32 // largest allowed context precision.
	MinPrec = 0             // smallest allowed context precision.
)

// DefaultPrecision is the default precision used for decimals created as
// literals or using new.
const DefaultPrecision = 16

const noPrecision = -1

// DefaultTraps is the default traps used for decimals created as literals or
// using new.
const DefaultTraps = ^(Inexact | Rounded | Subnormal)

const noTraps Condition = 1

// Context is a per-decimal contextual object that governs specific operations
// such as how lossy operations (e.g. division) round.
type Context struct {
	// OperatingMode which dictates how the
	// decimal operates. For example, the Decimal Arithmetic Specification
	// (version 1.70) requires that if a decimal is an infinity, the "String"
	// (or equivalent) method must return the string "Infinity." This, however,
	// differs from other Go types like float64 and big.Float that return
	// "+Inf" or "-Inf." To compensate, Context provides multiple modes that so
	// the client can choose a preferred mode.
	OperatingMode OperatingMode

	precision  int32
	traps      Condition
	conditions Condition
	err        error

	// RoundingMode instructs how an infinite (repeating) decimal expansion
	// (digits following the radix) should be rounded. This can occur during
	// "lossy" operations like division.
	RoundingMode RoundingMode
}

// Precision returns the Context's precision.
func (c Context) Precision() int32 {
	switch c.precision {
	case 0:
		return DefaultPrecision
	case noPrecision:
		return 0
	default:
		return c.precision
	}
}

// SetPrecision sets c's precision.
func (c *Context) SetPrecision(prec int32) {
	if prec == 0 {
		c.precision = noPrecision
	} else {
		c.precision = prec
	}
}

// SetTraps sets c's traps conditions.
func (c *Context) SetTraps(t Condition) {
	if t == 0 {
		c.traps = noTraps
	} else {
		c.traps = t
	}
}

// Traps returns the Context's traps.
func (c Context) Traps() Condition {
	switch c.traps {
	case 0:
		return DefaultTraps
	case noTraps:
		return 0
	default:
		return c.traps
	}
}

// The following are called ContextXX instead of DecimalXX
// to reserve the DecimalXX namespace for future decimal types.

// The following Contexts are based on IEEE 754R. Context is exported for this
// documentation but is not expected to be used itself. Each Context's
// RoundingMode is ToNearestEven, OperatingMode is GDA, and traps are set to
// DefaultTraps.
var (
	// Context32 is the IEEE 754R Decimal32 format.
	Context32 = Context{
		precision:     7,
		RoundingMode:  ToNearestEven,
		OperatingMode: GDA,
		traps:         DefaultTraps,
	}

	// Context64 is the IEEE 754R Decimal64 format.
	Context64 = Context{
		precision:     16,
		RoundingMode:  ToNearestEven,
		OperatingMode: GDA,
		traps:         DefaultTraps,
	}

	// Context128 is the IEEE 754R Decimal128 format.
	Context128 = Context{
		precision:     34,
		RoundingMode:  ToNearestEven,
		OperatingMode: GDA,
		traps:         DefaultTraps,
	}
)

// RoundingMode determines how a decimal will be rounded if the exact result
// cannot accurately be represented.
type RoundingMode uint8

// The following rounding modes are supported.
const (
	ToNearestEven RoundingMode = iota // == IEEE 754-2008 roundTiesToEven
	ToNearestAway                     // == IEEE 754-2008 roundTiesToAway
	ToZero                            // == IEEE 754-2008 roundTowardZero
	AwayFromZero                      // no IEEE 754-2008 equivalent
	ToNegativeInf                     // == IEEE 754-2008 roundTowardNegative
	ToPositiveInf                     // == IEEE 754-2008 roundTowardPositive

	// Unneeded means finite decimal expansion. Lossy routines will panic if
	// this RoundingMode is provided and the lossy operation does not have a
	// finite decimal expansion.
	Unneeded
)

//go:generate stringer -type RoundingMode

// OperatingMode dictates how the decimal approaches specific non-numeric
// operations like conversions to strings and panicking on NaNs. See Context's
// documentation for further information.
type OperatingMode uint8

const (
	// Go adheres to typical Go idioms
	Go OperatingMode = iota
	// GDA strictly adheres to the General Decimal Arithmetic Specification
	// Version 1.70
	GDA
)

//go:generate stringer -type OperatingMode

// Condition is a bitmask value raised after or during specific operations. For
// example, dividing by zero is undefined so a DivisionByZero Condition flag
// will be set in the decimal's Context.
type Condition uint32

const (
	// Clamped occurs if the scale has been modified to fit the constraints of
	// the decimal representation.
	Clamped Condition = 1 << (1 + iota)
	// ConversionSyntax occurs when a string is converted to a decimal and does
	// not have a valid syntax.
	ConversionSyntax
	// DivisionByZero occurs when division is attempted with a finite,
	// non-zero dividend and a divisor with a value of zero.
	DivisionByZero
	// DivisionImpossible occurs when the result of integer division would
	// contain too many digits (i.e. be longer than the specified precision).
	DivisionImpossible
	// DivisionUndefined occurs when division is attempted with in which both
	// the divided and divisor are zero.
	DivisionUndefined
	// Inexact occurs when the result of an operation (e.g. division) is not
	// exact, or when the Overflow/Underflow Conditions occur.
	Inexact
	// InsufficientStorage occurs when the system doesn't have enough storage
	// (i.e. memory) to store the decimal.
	InsufficientStorage
	// InvalidContext occurs when an invalid context was detected during an
	// operation. This might occur if, for example, an invalid RoundingMode was
	// passed to a Context.
	InvalidContext
	// InvalidOperation occurs when:
	//
	// 	- an operand to an operation is a signaling NaN
	// 	- an attempt is made to add or subtract infinities of opposite signs
	// 	- an attempt is made to multiply zero by an infinity of either sign
	// 	- an attempt is made to divide an infinity by an infinity
	// 	- the divisor for a remainder operation is zero
	// 	- the dividend for a remainder operation is an infinity
	// 	- either operand of the quantize operation is an infinity, or the
	// 	  result of a quantize operation would require greater precision than
	// 	  is available
	// 	- the operand of the ln or the log10 operation is less than zero
	// 	- the operand of the square-root operation has a sign of 1 and a
	// 	  non-zero coefficient
	// 	- both operands of the power operation are zero, or if the left-hand
	// 	  operand is less than zero and the right-hand operand does not have an
	// 	  integral value or is an infinity
	//
	InvalidOperation
	// Overflow occurs when the adjusted scale, after rounding, would be
	// greater than MaxScale. (Inexact and Rounded will also be raised.)
	Overflow
	// Rounded occurs when the result of an operation is rounded, or if an
	// Overflow/Underflow occurs.
	Rounded
	// Subnormal ocurs when the result of a conversion or operation is
	// subnormal (i.e. the adjusted scale is less than MinScale before any
	// rounding).
	Subnormal
	// Underflow occurs when the result is inexact and the adjusted scale would
	// be smaller (more negative) than MinScale.
	Underflow
)

//go:generate stringer -type Condition

var (
	one = New(1, 0)

	oneInt = big.NewInt(1)
	twoInt = big.NewInt(2)
)
