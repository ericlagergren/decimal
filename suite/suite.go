// Package suite provides a simple API for parsing and using IBM Labs'
// "Floating-Point Test-Suite for IEEE"
package suite

//go:generate go run getcases.go
//go:generate go run makejson.go

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
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

		c, err := parseCase(p)
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
	Trap   Exception
	Inputs []Data
	Output Data
	Excep  Exception
}

func (c Case) String() string {
	return fmt.Sprintf("%s%d [%s, %s]: %s(%s) = %s %s",
		c.Prefix, c.Prec, c.Trap, c.Mode, c.Op,
		join(c.Inputs, ", "), c.Output, c.Excep)
}

func join(a []Data, sep Data) Data {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return a[0]
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n)
	bp := copy(b, a[0])
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return Data(b)
}

// Data is input or output from a test case.
type Data string

// NoData is output when the operation throws some sort of exception and does
// not "return" any data.
const NoData Data = "#"

// IsNaN returns two booleans indicating whether the data is a NaN value
// and whether it's signaling or not.
func (i Data) IsNaN() (nan, signal bool) {
	return (i == "S" || i == "Q"), i == "S"
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

// Exception is a type of exception.
type Exception uint8

// These values are a bitmask corresponding to specific exceptions. For
// example, an Exception is allowed to be both Inexact and an Overflow.
const (
	None    Exception = 0
	Inexact Exception = 1 << iota
	Underflow
	Overflow
	DivByZero
	Invalid
)

var exceptions = [...]struct {
	e Exception
	s string
}{
	{Inexact, "Inexact"},
	{Underflow, "Underflow"},
	{Overflow, "Overflow"},
	{DivByZero, "DivByZero"},
	{Invalid, "Invalid"},
}

func (e Exception) String() string {
	if e == None {
		return "None"
	}

	var res string
	for _, x := range exceptions {
		if e&x.e != 0 {
			res += x.s + " | "
		}
	}
	return strings.TrimSuffix(res, " | ")
}

var valToException = map[string]Exception{
	"x": Inexact,
	"u": Underflow, // tininess and "extraordinary" error
	"v": Underflow, // tininess and inexactness after rounding
	"w": Underflow, // tininess and inexactness prior to rounding
	"o": Overflow,
	"z": DivByZero,
	"i": Invalid,
}

var valToMode = map[string]big.RoundingMode{
	">":  big.ToPositiveInf,
	"<":  big.ToNegativeInf,
	"0":  big.ToZero,
	"=0": big.ToNearestEven,
	"=^": big.ToNearestAway,
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
)

func init() {
	if len(valToOp) != int(Equiv)+1 /* +1 since Add is 0 */ {
		panic(fmt.Sprintf("wanted %d toks, got %d", Equiv, len(valToOp)))
	}
}

const maxOpLen = 6

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
}

//go:generate stringer -type=Op
