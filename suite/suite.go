// Package suite provides a simple API for parsing and using IBM Labs'
// "Floating-Point Test-Suite for IEEE"
package suite

//go:generate go run getcases.go
//go:generate go run makejson.go

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
)

// ParseCases returns a slice of test cases read from r.
func ParseCases(r io.Reader) (cases []Case, err error) {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)

	var c Case
	var line []byte
	for s.Scan() {
		line = s.Bytes()

		c.Prefix = string(line[0])
		line = line[1:]

		for i, v := range line {
			if v < '0' || v > '9' {
				c.Prec, err = strconv.Atoi(string(line[:i]))
				if err != nil {
					return nil, err
				}
				line = line[i:]
				break
			}
		}

		i := bytes.IndexByte(line, ' ')
		if i < 0 {
			return nil, fmt.Errorf("invalid line pre-op: %s", line)
		}

		var ok bool
		c.Op, ok = valToOp[string(line[:i])]
		if !ok {
			return nil, fmt.Errorf("invalid op: %s", line[:i])
		}
		line = line[i+1:]

		i = bytes.IndexByte(line, ' ')
		if i < 0 {
			return nil, fmt.Errorf("invalid line pre-mode: %s", line)
		}

		c.Mode, ok = valToMode[string(line[:i])]
		if !ok {
			return nil, fmt.Errorf("invalid mode: %s", line[:i])
		}
		line = line[i+1:]

		i = bytes.IndexByte(line, ' ')
		if i < 0 {
			return nil, fmt.Errorf("invalid line pre-trap: %s", line)
		}

		c.Trap, ok = valToException[string(line[:i])]
		if !ok {
			c.Trap = None
		}
		line = line[i+1:]

		i = bytes.Index(line, []byte{'-', '>'})
		if i < 0 {
			return nil, fmt.Errorf("invalid line pre-inputs: %s", line)
		}

		for _, p := range bytes.Split(line[:i], []byte{' '}) {
			if len(p) == 0 {
				continue
			}
			c.Inputs = append(c.Inputs, Data(p))
		}
		line = line[i+2+1:]

		i = bytes.IndexByte(line, ' ')
		if i < 0 {
			return nil, fmt.Errorf("invalid line pre-exception: %s", line)
		}
		c.Output = Data(line[:i])
		line = line[i+1:]

		c.Excep, ok = valToException[string(line)]
		if !ok && len(line) != 0 {
			return nil, fmt.Errorf("invalid resulting exception: %s", line)
		}
		cases = append(cases, c)
		// Reset the inputs otherwise we end up with a *ton* of inputs that's
		// 1) incorrect, and 2) makes 500MB+ files.
		c.Inputs = c.Inputs[0:0]
	}
	return cases, s.Err()
}

// Test suite documentation is from:
// https://www.research.ibm.com/haifa/projects/verification/fpgen/papers/ieee-test-suite-v2.pdf

// Case represents a specific test case.
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

// Data is input or output from a test case.
type Data string

// IsNaN returns two booleans indicating whether the data is a NaN value
// and whether it's signaling or not.
func (i Data) IsNaN() (nan, signal bool) {
	return (i == "S" || i == "Q"), i == "S"
}

// Inf returns a boolean indicating whether the data is an Infinity and an
// int indicating the signedness of the Infinity.
func (i Data) Inf() (int, bool) {
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
	None Exception = 1 << iota
	Inexact
	Underflow
	Overflow
	DivByZero
	Invalid
)

const maxExceptionLen = 1

var valToException = map[string]Exception{
	"x":  Inexact,
	"u":  Underflow, // tininess and "extraordinary" error
	"v":  Underflow, // tininess and inexactness after rounding
	"w":  Underflow, // tininess and inexactness prior to rounding
	"o":  Overflow,
	"z":  DivByZero,
	"i":  Invalid,
	"xo": Inexact | Overflow,
	"xu": Inexact | Underflow,
}

const maxModeLen = 2

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
