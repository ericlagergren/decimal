package dectest

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
	"github.com/ericlagergren/decimal/misc"
)

func Test(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := NewScanner(f)
	for s.Scan() {
		c := s.Case()
		if !isSupported(c) {
			continue
		}
		if testing.Short() {
			if _, ok := longRunning[c.ID]; ok {
				continue
			}
		}
		err := execute(c)
		if err != nil {
			return err
		}
	}
	return s.Err()
}

var nilary = map[Op]func(z *decimal.Big) *decimal.Big{
	Reduce:      (*decimal.Big).Reduce,
	ToIntegralX: (*decimal.Big).RoundToInt,
}

var unary = map[Op]func(z, x *decimal.Big) *decimal.Big{
	Apply:      (*decimal.Big).Set,
	Abs:        (*decimal.Big).Abs,
	Canonical:  misc.Canonical,
	Copy:       (*decimal.Big).Copy,
	CopyAbs:    misc.CopyAbs,
	CopyNegate: misc.CopyNeg,
	Exp:        math.Exp,
	LogB:       math.Log10,
	Log10:      math.Log10,
	Ln:         math.Log,
	Minus:      (*decimal.Big).Neg,
	NextMinus:  misc.NextMinus,
	NextPlus:   misc.NextPlus,
	SquareRoot: math.Sqrt,
}

var binary = map[Op]func(z, x, y *decimal.Big) *decimal.Big{
	Add:       (*decimal.Big).Add,
	CopySign:  (*decimal.Big).CopySign,
	Divide:    (*decimal.Big).Quo,
	DivideInt: (*decimal.Big).QuoInt,
	Max:       func(z, x, y *decimal.Big) *decimal.Big { return misc.Max(x, y) },
	MaxMag:    func(z, x, y *decimal.Big) *decimal.Big { return misc.MaxAbs(x, y) },
	Min:       func(z, x, y *decimal.Big) *decimal.Big { return misc.Min(x, y) },
	MinMag:    func(z, x, y *decimal.Big) *decimal.Big { return misc.MinAbs(x, y) },
	Multiply:  (*decimal.Big).Mul,
	Power:     math.Pow,
	Remainder: (*decimal.Big).Rem,
	Subtract:  (*decimal.Big).Sub,
}

var ternary = map[Op]func(z, x, y, u *decimal.Big) *decimal.Big{
	FMA: (*decimal.Big).FMA,
}

var decRoundingModes = map[RoundingMode]decimal.RoundingMode{
	Ceiling:  decimal.ToPositiveInf,
	Down:     decimal.ToZero,
	Floor:    decimal.ToNegativeInf,
	HalfEven: decimal.ToNearestEven,
	HalfUp:   decimal.ToNearestAway,
	HalfDown: decimal.ToNearestTowardZero,
}

var decConditions = map[Condition]decimal.Condition{
	Clamped:             decimal.Clamped,
	ConversionSyntax:    decimal.ConversionSyntax,
	DivisionByZero:      decimal.DivisionByZero,
	DivisionImpossible:  decimal.DivisionImpossible,
	DivisionUndefined:   decimal.DivisionUndefined,
	Inexact:             decimal.Inexact,
	InsufficientStorage: decimal.InsufficientStorage,
	InvalidContext:      decimal.InvalidContext,
	InvalidOperation:    decimal.InvalidOperation,
	Overflow:            decimal.Overflow,
	Rounded:             decimal.Rounded,
	Subnormal:           decimal.Subnormal,
	Underflow:           decimal.Underflow,
}

var ErrSkipTest = errors.New("skipped test")

func execute(c *Case) error {
	if c.MaxScale > decimal.MaxScale {
		return fmt.Errorf("invalid max scale: %d", c.MaxScale)
	}

	if c.MinScale < decimal.MinScale {
		return fmt.Errorf("invalid min scale: %d", c.MinScale)
	}

	if c.MinScale < decimal.MinScale {
		return fmt.Errorf("invalid min scale: %d", c.MinScale)
	}

	if c.Prec < decimal.MinPrecision || c.Prec > decimal.MaxPrecision {
		return fmt.Errorf("invalid precision: %d", c.Prec)
	}

	if _, ok := skip[c.ID]; ok {
		return ErrSkipTest
	}

	flags, ok := convertConditions(c.Conditions)
	if !ok {
		return fmt.Errorf("invalid condition(s): %s", c.Conditions)
	}

	mode, ok := decRoundingModes[c.Mode]
	if !ok {
		return fmt.Errorf("invalid rounding mode: %s", c.Mode)
	}

	ctx := decimal.Context{
		Precision:     c.Prec,
		OperatingMode: decimal.GDA,
		RoundingMode:  mode,
		MinScale:      c.MinScale,
		MaxScale:      c.MaxScale,
	}

	z, x, y, u := parseInputs(ctx, c)
	r := parseOutput(ctx, c, flags)

	if nfn, ok := nilary[c.Op]; ok {
		return check(nfn(x), r, c, flags)
	}
	if ufn, ok := unary[c.Op]; ok {
		return check(ufn(z, x), r, c, flags)
	}
	if bfn, ok := binary[c.Op]; ok {
		return check(bfn(z, x, y), r, c, flags)
	}
	if tfn, ok := ternary[c.Op]; ok {
		return check(tfn(z, x, y, u), r, c, flags)
	}

	switch c.Op {
	case Class:
		return assert(c, x.Class(), r)
	case Compare:
		rv := x.Cmp(y)
		r, _, snan, err := cmp(c)
		if err != nil {
			return err
		}
		err = assert(c, rv, r)
		if err != nil {
			return err
		}
		return assert(c, snan, x.Context.Conditions&decimal.InvalidOperation != 0)
	case CompareTotal:
		rv := misc.CmpTotal(x, y)
		r, _, snan, err := cmp(c)
		if err != nil {
			return err
		}
		err = assert(c, rv, r)
		if err != nil {
			return err
		}
		return assert(c, snan, x.Context.Conditions&decimal.InvalidOperation != 0)
	case CompareTotMag:
		rv := misc.CmpTotalAbs(x, y)
		r, _, snan, err := cmp(c)
		if err != nil {
			return err
		}
		err = assert(c, rv, r)
		if err != nil {
			return err
		}
		return assert(c, snan, x.Context.Conditions&decimal.InvalidOperation != 0)
	case Max:
		return check(z.Set(misc.Max(x, y)), r, c, flags)
	case Min:
		return check(z.Set(misc.Min(x, y)), r, c, flags)
	case Quantize:
		v, _ := y.Int64()
		return check(x.Quantize(int(v)), r, c, flags)
	case SameQuantum:
		rv := misc.SameQuantum(x, y)
		return assert(c, rv, c.Output == Data("1"))
	case ToSci:
		rv := fmt.Sprintf("%E", x)
		return assert(c, rv, string(c.Output))
	default:
		return fmt.Errorf("unknown op: %s", c.Op.String())
	}
}

func isSupported(c *Case) bool {
	if c.Clamp {
		return false
	}

	if _, ok := convertConditions(c.Conditions); !ok {
		return false
	}

	if _, ok := decRoundingModes[c.Mode]; !ok {
		return false
	}

	var opSupported bool
	if _, ok := nilary[c.Op]; ok {
		opSupported = true
	} else if _, ok := unary[c.Op]; ok {
		opSupported = true
	} else if _, ok := binary[c.Op]; ok {
		opSupported = true
	} else if _, ok := ternary[c.Op]; ok {
		opSupported = true
	} else {
		switch c.Op {
		case Class, Compare, CompareTotal, CompareTotMag, Max, Min, Quantize, SameQuantum, ToSci:
			opSupported = true
		}
	}

	return opSupported
}

func open(fpath string) io.ReadCloser {
	file, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	return file
}

func parseOutput(ctx decimal.Context, c *Case, f decimal.Condition) *decimal.Big {
	r := dataToBig(ctx, c.Output)
	r.Context.Conditions = f
	return r
}

func dataToBig(ctx decimal.Context, d Data) *decimal.Big {
	if d == NoData {
		return decimal.New(0, 0).SetInf(false)
	}
	b, ok := decimal.WithContext(ctx).SetString(string(d.TrimQuotes()))
	if !ok {
		b = decimal.WithContext(ctx).SetNaN(true)
	}
	return b
}

func parseInputs(ctx decimal.Context, c *Case) (z *decimal.Big, x *decimal.Big, y *decimal.Big, u *decimal.Big) {
	z = decimal.WithContext(ctx)
	switch len(c.Inputs) {
	case 3:
		u = dataToBig(ctx, c.Inputs[2])
		fallthrough
	case 2:
		y = dataToBig(ctx, c.Inputs[1])
		fallthrough
	case 1:
		x = dataToBig(ctx, c.Inputs[0])
	case 0:
		break
	default:
		panic(fmt.Errorf("%s: invalid number of inputs (%d)", c.ID, len(c.Inputs)))
	}
	return
}

func assert(c *Case, a, b interface{}) error {
	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf(`%s
wanted: %v
got   : %v
`, c.ShortString(22), b, a)
	}
	return nil
}

func check(z, r *decimal.Big, c *Case, flags decimal.Condition) error {
	if !equal(z, r) {
		return fmt.Errorf(`%s
wanted: %v (%s:%d)
got   : %v (%s:%d)
`,
			c.ShortString(10000),
			r, flags, -r.Scale(),
			z, z.Context.Conditions, -z.Scale(),
		)
	}
	return nil
}

func cmp(c *Case) (int, bool, bool, error) {
	qnan, snan := Data(c.Output).IsNaN()
	if qnan || snan {
		return 0, qnan, snan, nil
	}
	r, err := strconv.Atoi(string(c.Output))
	if err != nil {
		return 0, false, false, err
	}
	return r, false, false, nil
}

func equal(x, y *decimal.Big) bool {
	if x.Signbit() != y.Signbit() {
		return false
	}
	if x.IsFinite() != y.IsFinite() {
		return false
	}
	if !x.IsFinite() {
		return (x.IsInf(0) && y.IsInf(0)) || (x.IsNaN(0) && y.IsNaN(0))
	}
	if x.Context.Conditions != y.Context.Conditions {
		return false
	}
	cmp := x.Cmp(y) == 0
	scl := x.Scale() == y.Scale()
	prec := x.Precision() == y.Precision()
	return cmp && scl && prec
}

// helper returns testing.T.Helper, if it exists.
func helper(v interface{}) func() {
	if fn, ok := v.(interface {
		Helper()
	}); ok {
		return fn.Helper
	}
	return func() {}
}

func convertConditions(c Condition) (decimal.Condition, bool) {
	var r decimal.Condition
	x := c // check to make sure all flags are copied
	for k, v := range decConditions {
		if c&k == k {
			r |= v
			x ^= k
		}
	}
	if x != 0 {
		return 0, false
	}
	return r, true
}
