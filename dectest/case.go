package dectest

import (
	"fmt"
	"strings"

	. "github.com/ericlagergren/decimal"
)

type Case struct {
	ID         string
	Prec       int
	Clamp      bool
	Mode       RoundingMode
	MaxScale   int
	MinScale   int
	Op         Op
	Inputs     []Data
	Output     Data
	Conditions Condition
}

func (c Case) String() string {
	return fmt.Sprintf("%s %d [%s, %s]: %s = %s %s",
		c.ID, c.Prec, c.Mode, c.Op,
		join(c.Inputs, ", ", -1), c.Output, c.Conditions)
}

// ShortString returns the same as String, except long data values are capped at
// length digits.
func (c Case) ShortString(length int) string {
	return fmt.Sprintf("%s %d [%s, %s]: %s = %s %s",
		c.ID, c.Prec, c.Mode, c.Op,
		join(c.Inputs, ", ", length), trunc(c.Output, length), c.Conditions)
}

func trunc(d Data, l int) Data {
	if l <= 0 || l >= len(d) {
		return d
	}
	return d[:l/2] + "..." + d[len(d)-(l/2):]
}

func writeTrunc(b *strings.Builder, d Data, maxLen int) {
	if maxLen <= 0 || len(d) < maxLen {
		b.WriteString(string(d))
	} else {
		b.WriteString(string(d[:maxLen/2]))
		b.WriteString("...")
		b.WriteString(string(d[len(d)-(maxLen/2):]))
	}
}

func join(a []Data, sep string, maxLen int) Data {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return trunc(a[0], maxLen)
	}
	n := len(sep) * (len(a) - 1)
	for _, s := range a {
		if len(s) > maxLen {
			n += maxLen + len("...")
		} else {
			n += len(s)
		}
	}

	var b strings.Builder
	b.Grow(n)
	writeTrunc(&b, a[0], maxLen)
	for _, s := range a[1:] {
		b.WriteString(sep)
		writeTrunc(&b, s, maxLen)
	}
	return Data(b.String())
}

// Data is input or output from a test case.
type Data string

// NoData is output when the operation throws some sort of Condition and does
// not "return" any data.
const NoData Data = "?"

func (i Data) TrimQuotes() Data {
	return Data(strings.Trim(string(i), "'"))
}

// IsNaN returns two booleans indicating whether the data is a NaN value and
// whether it's signaling or not.
func (i Data) IsNaN() (nan, signal bool) {
	i = i.TrimQuotes()
	if len(i) == 1 {
		return (i == "S" || i == "Q"), i == "S"
	}
	if i[0] == '-' {
		i = i[1:]
	}
	// trim instance number e.g. NaN1
	s := strings.TrimRight(string(i), "0123456789")
	return strings.EqualFold(s, "nan") ||
		strings.EqualFold(s, "qnan") ||
		strings.EqualFold(s, "snan"), i[0] == 's' || i[0] == 'S'
}

// IsInf returns a boolean indicating whether the data is an Infinity and an
// int indicating the signedness of the Infinity.
func (i Data) IsInf() (int, bool) {
	i = i.TrimQuotes()
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

var conditions = map[string]Condition{
	"clamped":              Clamped,
	"conversion_syntax":    ConversionSyntax,
	"division_by_zero":     DivisionByZero,
	"division_impossible":  DivisionImpossible,
	"division_undefined":   DivisionUndefined,
	"inexact":              Inexact,
	"insufficient_storage": InsufficientStorage,
	"invalid_context":      InvalidContext,
	"invalid_operation":    InvalidOperation,
	"lost_digits":          0, // TODO
	"overflow":             Overflow,
	"rounded":              Rounded,
	"subnormal":            Subnormal,
	"underflow":            Underflow,
}

// Op is a specific operation the test case must perform.
type Op uint8

const (
	OpUnknownOp Op = iota
	OpAbs
	OpAdd
	OpAnd
	OpApply
	OpCanonical
	OpClass
	OpCompare
	OpCompareSig
	OpCompareTotal
	OpCompareTotMag
	OpCopy
	OpCopyAbs
	OpCopyNegate
	OpCopySign
	OpDivide
	OpDivideInt
	OpExp
	OpFMA
	OpInvert
	OpLn
	OpLog10
	OpLogB
	OpMax
	OpMaxMag
	OpMin
	OpMinMag
	OpMinus
	OpMultiply
	OpNextMinus
	OpNextPlus
	OpNextToward
	OpOr
	OpPlus
	OpPower
	OpQuantize
	OpReduce
	OpRemainder
	OpRemainderNear
	OpRescale
	OpRotate
	OpSameQuantum
	OpScaleB
	OpShift
	OpSquareRoot
	OpSubtract
	OpToEng
	OpToIntegral
	OpToIntegralX
	OpToSci
	OpTrim
	OpXor
)

var operations = map[string]Op{
	"abs":           OpAbs,
	"add":           OpAdd,
	"and":           OpAnd,
	"apply":         OpApply,
	"canonical":     OpCanonical,
	"class":         OpClass,
	"compare":       OpCompare,
	"comparesig":    OpCompareSig,
	"comparetotal":  OpCompareTotal,
	"comparetotmag": OpCompareTotMag,
	"copy":          OpCopy,
	"copyabs":       OpCopyAbs,
	"copynegate":    OpCopyNegate,
	"copysign":      OpCopySign,
	"divide":        OpDivide,
	"divideint":     OpDivideInt,
	"exp":           OpExp,
	"fma":           OpFMA,
	"invert":        OpInvert,
	"ln":            OpLn,
	"log10":         OpLog10,
	"logb":          OpLogB,
	"max":           OpMax,
	"maxmag":        OpMaxMag,
	"min":           OpMin,
	"minmag":        OpMinMag,
	"minus":         OpMinus,
	"multiply":      OpMultiply,
	"nextminus":     OpNextMinus,
	"nextplus":      OpNextPlus,
	"nexttoward":    OpNextToward,
	"or":            OpOr,
	"plus":          OpPlus,
	"power":         OpPower,
	"quantize":      OpQuantize,
	"reduce":        OpReduce,
	"remainder":     OpRemainder,
	"remaindernear": OpRemainderNear,
	"rescale":       OpRescale,
	"rotate":        OpRotate,
	"samequantum":   OpSameQuantum,
	"scaleb":        OpScaleB,
	"shift":         OpShift,
	"squareroot":    OpSquareRoot,
	"subtract":      OpSubtract,
	"toeng":         OpToEng,
	"tointegral":    OpToIntegral,
	"tointegralx":   OpToIntegralX,
	"tosci":         OpToSci,
	"trim":          OpTrim,
	"xor":           OpXor,
}

var roundingModes = map[string]RoundingMode{
	"ceiling":   ToPositiveInf,
	"down":      ToZero,
	"floor":     ToNegativeInf,
	"half_down": ToNearestTowardZero,
	"half_even": ToNearestEven,
	"half_up":   ToNearestAway,
	"up":        0,
	"05up":      0,
}

//go:generate stringer -type=Op
