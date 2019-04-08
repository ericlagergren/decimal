// Package suite provides a simple API for parsing and using IBM Labs'
// "Floating-Point Test-Suite for IEEE"
//
// This package is deprecated and will be removed in the next major version.
package suite

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"math/bits"
	"strings"
)

// ParseCases returns a slice of test cases in .fptest form read from r.
func ParseCases(r io.Reader) (cases []Case, err error) {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)

	for s.Scan() {
		p := s.Bytes()
		// Skip empty lines and comments.
		if len(p) == 0 || p[0] == '#' {
			continue
		}

		c, err := ParseCase(p)
		if err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, s.Err()
}

// Test suite documentation is from:
// https://www.research.ibm.com/haifa/projects/verification/fpgen/papers/ieee-test-suite-v2.pdf

// Case represents a specific test case.
//
// Here's a nice ascii diagram:
//
//     prec   trap             excep
//     |      |                |
//     |\     |                |
//     vv     v                v
//    d64+ =0 i 100 200 -> 300 i
//    ^  ^ ^^   ^^^ ^^^  ^ ^^^
//    |  | \|   \|/ \|/  | \|/
//    |  |  |    |   |   |  |
//    |  |  mode  \ /    |  output
//    |  op        |     output delim
//    prefix       inputs
//
type Case struct {
	Prefix string
	Prec   int
	Op     Op
	Mode   big.RoundingMode
	Trap   Condition
	Inputs []Data
	Output Data
	Excep  Condition
}

// TODO(eric): String should print the same format as the input

func (c Case) String() string {
	return fmt.Sprintf("%s%d [%s, %s]: %s(%s) = %s %s",
		c.Prefix, c.Prec, c.Trap, c.Mode, c.Op,
		join(c.Inputs, ", ", -1), c.Output, c.Excep)
}

// ShortString returns the same as String, except long data values are capped at
// length digits.
func (c Case) ShortString(length int) string {
	return fmt.Sprintf("%s%d [%s, %s]: %s(%s) = %s %s",
		c.Prefix, c.Prec, c.Trap, c.Mode, c.Op,
		join(c.Inputs, ", ", length), trunc(c.Output, length), c.Excep)
}

func trunc(d Data, l int) Data {
	if l <= 0 || l >= len(d) {
		return d
	}
	return d[:l/2] + "..." + d[len(d)-(l/2):]
}

func join(a []Data, sep Data, l int) Data {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return trunc(a[0], l)
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(trunc(a[i], l))
	}

	b := make([]byte, n)
	bp := copy(b, trunc(a[0], l))
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], trunc(s, l))
	}
	return Data(b)
}

// Data is input or output from a test case.
type Data string

// NoData is output when the operation throws some sort of Condition and does
// not "return" any data.
const NoData Data = "#"

// IsNaN returns two booleans indicating whether the data is a NaN value and
// whether it's signaling or not.
func (i Data) IsNaN() (nan, signal bool) {
	if len(i) == 1 {
		return (i == "S" || i == "Q"), i == "S"
	}
	if i[0] == '-' {
		i = i[1:]
	}
	return strings.EqualFold(string(i), "nan") ||
		strings.EqualFold(string(i), "qnan") ||
		strings.EqualFold(string(i), "snan"), i[0] == 's' || i[0] == 'S'
}

// IsInf returns a boolean indicating whether the data is an Infinity and an
// int indicating the signedness of the Infinity.
func (i Data) IsInf() (int, bool) {
	if len(i) != 4 {
		return 0, false
	}
	if strings.EqualFold(string(i), "-Inf") {
		return -1, true
	}
	if strings.EqualFold(string(i), "+Inf") {
		return +1, true
	}
	return 0, false
}

// Condition is a bitmask value raised after or during specific operations.
type Condition uint32

const (
	Clamped Condition = 1 << iota
	ConversionSyntax
	DivisionByZero
	DivisionImpossible
	DivisionUndefined
	Inexact
	InsufficientStorage
	InvalidContext
	InvalidOperation
	Overflow
	Rounded
	Subnormal
	Underflow
)

func (c Condition) String() string {
	if c == 0 {
		return "NoConditions"
	}

	// Each condition is one bit, so this saves some allocations.
	a := make([]string, 0, bits.OnesCount32(uint32(c)))
	for i := Condition(1); c != 0; i <<= 1 {
		if c&i == 0 {
			continue
		}
		switch c ^= i; i {
		case Clamped:
			a = append(a, "clamped")
		case ConversionSyntax:
			a = append(a, "conversion syntax")
		case DivisionByZero:
			a = append(a, "division by zero")
		case DivisionImpossible:
			a = append(a, "division impossible")
		case Inexact:
			a = append(a, "inexact")
		case InsufficientStorage:
			a = append(a, "insufficient storage")
		case InvalidContext:
			a = append(a, "invalid context")
		case InvalidOperation:
			a = append(a, "invalid operation")
		case Overflow:
			a = append(a, "overflow")
		case Rounded:
			a = append(a, "rounded")
		case Subnormal:
			a = append(a, "subnormal")
		case Underflow:
			a = append(a, "underflow")
		default:
			a = append(a, fmt.Sprintf("unknown(%d)", i))
		}
	}
	return strings.Join(a, ", ")
}

func ConditionFromString(s string) (r Condition) {
	for i := range s {
		r |= valToCondition[s[i]]
	}
	return r
}

var valToCondition = map[byte]Condition{
	'x': Inexact,
	'u': Underflow, // tininess and 'extraordinary' error
	'v': Underflow, // tininess and inexactness after rounding
	'w': Underflow, // tininess and inexactness prior to rounding
	'o': Overflow,
	'z': DivisionByZero,
	'i': InvalidOperation,

	// custom
	'c': Clamped,
	'r': Rounded,
	'y': ConversionSyntax,
	'm': DivisionImpossible,
	'n': DivisionUndefined,
	't': InsufficientStorage,
	'?': InvalidContext,
	's': Subnormal,
}

var valToMode = map[string]big.RoundingMode{
	">":  big.ToPositiveInf,
	"<":  big.ToNegativeInf,
	"0":  big.ToZero,
	"=0": big.ToNearestEven,
	"=^": big.ToNearestAway,
	"^":  big.AwayFromZero,
}

// Op is a specific operation the test case must perform.
type Op uint8

const (
	Add         Op = iota // add
	Sub                   // subtract
	Mul                   // multiply
	Div                   // divide
	FMA                   // fused multiply-add
	Sqrt                  // square root
	Rem                   // remainder
	RFI                   // round float to int
	CFF                   // convert between floating point formats
	CFI                   // convert float to integer
	CIF                   // convert integer to float
	CFD                   // convert to string
	CDF                   // convert string to float
	QuietCmp              // quiet comparison
	SigCmp                // signaling comparison
	Copy                  // copy
	Neg                   // negate
	Abs                   // absolute value
	CopySign              // copy sign
	Scalb                 // scalb
	Logb                  // logb
	NextAfter             // next after
	Class                 // class
	IsSigned              // is signed
	IsNormal              // is norm
	IsInf                 // is inf
	IsZero                // is zero
	IsSubNormal           // is subnormal
	IsNaN                 // is nan
	IsSignaling           // is signaling
	IsFinite              // is finite
	MinNum                // minnum
	MaxNum                // maxnum
	MinNumMag             // minnummag
	MaxNumMag             // maxnummag
	SameQuantum           // same quantum
	Quantize              // quantize
	NextUp                // next up
	NextDown              // next down
	Equiv                 // equivalent

	// Custom
	SetRat
	Sign
	Signbit
	Exp
	Log
	Log10
	Pow
	IntDiv
	Normalize
	RoundToInt
	Shift
)

var valToOp = map[string]Op{
	"+":      Add,
	"-":      Sub,
	"*":      Mul,
	"/":      Div,
	"*-":     FMA,
	"V":      Sqrt,
	"%":      Rem,
	"rfi":    RFI,
	"cff":    CFF,
	"cfi":    CFI,
	"cif":    CIF,
	"cfd":    CFD,
	"cdf":    CDF,
	"qC":     QuietCmp,
	"sC":     SigCmp,
	"cp":     Copy,
	"~":      Neg,
	"A":      Abs,
	"@":      CopySign,
	"S":      Scalb,
	"L":      Logb,
	"Na":     NextAfter,
	"?":      Class,
	"?-":     IsSigned,
	"?n":     IsNormal,
	"?f":     IsFinite,
	"?0":     IsZero,
	"?s":     IsSubNormal,
	"?i":     IsInf,
	"?N":     IsNaN,
	"?sN":    IsSignaling,
	"<C":     MinNum,
	">C":     MaxNum,
	"<A":     MinNumMag,
	">A":     MaxNumMag,
	"=quant": SameQuantum,
	"quant":  Quantize,
	"Nu":     NextUp,
	"Nd":     NextDown,
	"eq":     Equiv,

	// Custom
	"rat":     SetRat,
	"sign":    Sign,
	"signbit": Signbit,
	"exp":     Exp,
	"log":     Log,
	"log10":   Log10,
	"pow":     Pow,
	"//":      IntDiv,
	"norm":    Normalize,
	"rtie":    RoundToInt,
	"shift":   Shift,
}

//go:generate stringer -type=Op
