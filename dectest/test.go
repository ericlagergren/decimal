package dectest

import (
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

func Test(t *testing.T, file string) {
	r := open(file)
	defer r.Close()
	s := NewScanner(r)
	for s.Scan() {
		c := s.Case()
		if !isSupported(c) {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			execute(t, c)
		})
	}
	if err := s.Err(); err != nil {
		t.Error(err)
	}
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

func execute(t *testing.T, c *Case) {
	if c.MaxScale > decimal.MaxScale {
		t.Fatalf("invalid max scale: %d", c.MaxScale)
	}

	if c.MinScale < decimal.MinScale {
		t.Fatalf("invalid min scale: %d", c.MinScale)
	}

	if c.MinScale < decimal.MinScale {
		t.Fatalf("invalid min scale: %d", c.MinScale)
	}

	if c.Prec < decimal.MinPrecision || c.Prec > decimal.MaxPrecision {
		t.Fatalf("invalid precision: %d", c.Prec)
	}

	if _, ok := skip[c.ID]; ok {
		t.Skipf("skipped dectest")
	}

	flags, ok := convertConditions(c.Conditions)
	if !ok {
		t.Fatalf("invalid condition(s): %s", c.Conditions)
	}

	mode, ok := decRoundingModes[c.Mode]
	if !ok {
		t.Fatalf("invalid rounding mode: %s", c.Mode)
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
		check(t, nfn(x), r, c, flags)
	} else if ufn, ok := unary[c.Op]; ok {
		check(t, ufn(z, x), r, c, flags)
	} else if bfn, ok := binary[c.Op]; ok {
		check(t, bfn(z, x, y), r, c, flags)
	} else if tfn, ok := ternary[c.Op]; ok {
		check(t, tfn(z, x, y, u), r, c, flags)
	} else {
		switch c.Op {
		case Class:
			assert(t, c, x.Class(), r)
		case Compare:
			rv := x.Cmp(y)
			r, _, snan := cmp(t, c)
			assert(t, c, rv, r)
			assert(t, c, snan, x.Context.Conditions&decimal.InvalidOperation != 0)
		case CompareTotal:
			rv := misc.CmpTotal(x, y)
			r, _, snan := cmp(t, c)
			assert(t, c, rv, r)
			assert(t, c, snan, x.Context.Conditions&decimal.InvalidOperation != 0)
		case CompareTotMag:
			rv := misc.CmpTotalAbs(x, y)
			r, _, snan := cmp(t, c)
			assert(t, c, rv, r)
			assert(t, c, snan, x.Context.Conditions&decimal.InvalidOperation != 0)
		case Max:
			check(t, z.Set(misc.Max(x, y)), r, c, flags)
		case Min:
			check(t, z.Set(misc.Min(x, y)), r, c, flags)
		case Quantize:
			v, _ := y.Int64()
			check(t, x.Quantize(int(v)), r, c, flags)
		case SameQuantum:
			rv := misc.SameQuantum(x, y)
			assert(t, c, rv, c.Output == Data("1"))
		default:
			t.Fatalf("unknown op: " + c.Op.String())
		}
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
		case Class, Compare, CompareTotal, CompareTotMag, Max, Min, Quantize, SameQuantum:
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

func assert(t *testing.T, c *Case, a, b interface{}) {
	helper(t)()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf(`%s
wanted: %v
got   : %v
`, c.ShortString(22), b, a)
	}
}

func check(t *testing.T, z, r *decimal.Big, c *Case, flags decimal.Condition) {
	helper(t)()
	if !equal(z, r) {
		str := fmt.Sprintf(`%s
wanted: %q (%s:%d)
got   : %q (%s:%d)
`,
			c.ShortString(10000),
			r, flags, -r.Scale(),
			z, z.Context.Conditions, -z.Scale(),
		)
		t.Error(str)
	}
}

func cmp(t *testing.T, c *Case) (int, bool, bool) {
	qnan, snan := Data(c.Output).IsNaN()
	if qnan || snan {
		return 0, qnan, snan
	}
	r, err := strconv.Atoi(string(c.Output))
	if err != nil {
		helper(t)()
		t.Fatal(err)
	}
	return r, false, false
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
