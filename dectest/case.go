package dectest

import (
	"fmt"
	"math/bits"
	"strings"
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
	LostDigits
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
	"lost_digits":          LostDigits,
	"overflow":             Overflow,
	"rounded":              Rounded,
	"subnormal":            Subnormal,
	"underflow":            Underflow,
}

// Op is a specific operation the test case must perform.
type Op uint8

const (
	UnknownOp Op = iota
	Abs
	Add
	And
	Apply
	Canonical
	Class
	Compare
	CompareSig
	CompareTotal
	CompareTotMag
	Copy
	CopyAbs
	CopyNegate
	CopySign
	Divide
	DivideInt
	Exp
	FMA
	Invert
	Ln
	Log10
	LogB
	Max
	MaxMag
	Min
	MinMag
	Minus
	Multiply
	NextMinus
	NextPlus
	NextToward
	Or
	Plus
	Power
	Quantize
	Reduce
	Remainder
	RemainderNear
	Rescale
	Rotate
	SameQuantum
	ScaleB
	Shift
	SquareRoot
	Subtract
	ToEng
	ToIntegral
	ToIntegralX
	ToSci
	Trim
	Xor
)

var operations = map[string]Op{
	"abs":           Abs,
	"add":           Add,
	"and":           And,
	"apply":         Apply,
	"canonical":     Canonical,
	"class":         Class,
	"compare":       Compare,
	"comparesig":    CompareSig,
	"comparetotal":  CompareTotal,
	"comparetotmag": CompareTotMag,
	"copy":          Copy,
	"copyabs":       CopyAbs,
	"copynegate":    CopyNegate,
	"copysign":      CopySign,
	"divide":        Divide,
	"divideint":     DivideInt,
	"exp":           Exp,
	"fma":           FMA,
	"invert":        Invert,
	"ln":            Ln,
	"log10":         Log10,
	"logb":          LogB,
	"max":           Max,
	"maxmag":        MaxMag,
	"min":           Min,
	"minmag":        MinMag,
	"minus":         Minus,
	"multiply":      Multiply,
	"nextminus":     NextMinus,
	"nextplus":      NextPlus,
	"nexttoward":    NextToward,
	"or":            Or,
	"plus":          Plus,
	"power":         Power,
	"quantize":      Quantize,
	"reduce":        Reduce,
	"remainder":     Remainder,
	"remaindernear": RemainderNear,
	"rescale":       Rescale,
	"rotate":        Rotate,
	"samequantum":   SameQuantum,
	"scaleb":        ScaleB,
	"shift":         Shift,
	"squareroot":    SquareRoot,
	"subtract":      Subtract,
	"toeng":         ToEng,
	"tointegral":    ToIntegral,
	"tointegralx":   ToIntegralX,
	"tosci":         ToSci,
	"trim":          Trim,
	"xor":           Xor,
}

type RoundingMode int

const (
	Ceiling RoundingMode = iota
	Down
	Floor
	HalfDown
	HalfEven
	HalfUp
	Up
	ZeroFiveUp
)

var roundingModes = map[string]RoundingMode{
	"ceiling":   Ceiling,
	"down":      Down,
	"floor":     Floor,
	"half_down": HalfDown,
	"half_even": HalfEven,
	"half_up":   HalfUp,
	"up":        Up,
	"05up":      ZeroFiveUp,
}

//go:generate stringer -type=Op
//go:generate stringer -type=RoundingMode
