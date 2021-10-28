package dectest

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"

	. "github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
)

func Test(t *testing.T, file string) {
	r := open(file)
	t.Cleanup(func() { r.Close() })

	s := NewScanner(r)
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
		execute(t, c)
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
}

var nilary = map[Op]func(ctx Context, z *Big) *Big{
	OpReduce:      (Context).Reduce,
	OpToIntegralX: (Context).RoundToInt,
}

var unary = map[Op]func(ctx Context, z, x *Big) *Big{
	OpApply: (Context).Set,
	OpAbs:   Context.Abs,
	OpCanonical: func(_ Context, z, x *Big) *Big {
		return z.Canonical(x)
	},
	OpCopy: func(_ Context, z, x *Big) *Big {
		return z.Copy(x)
	},
	OpCopyAbs: func(_ Context, z, x *Big) *Big {
		return z.CopyAbs(x)
	},
	OpCopyNegate: func(_ Context, z, x *Big) *Big {
		return z.CopyNeg(x)
	},
	OpExp:        Context.Exp,
	OpLogB:       Context.Log10,
	OpLog10:      Context.Log10,
	OpLn:         Context.Log,
	OpMinus:      Context.Neg,
	OpNextMinus:  Context.NextMinus,
	OpNextPlus:   Context.NextPlus,
	OpSquareRoot: Context.Sqrt,
}

var binary = map[Op]func(ctx Context, z, x, y *Big) *Big{
	OpAdd: Context.Add,
	OpCopySign: func(_ Context, z, x, y *Big) *Big {
		return z.CopySign(x, y)
	},
	OpDivide:    Context.Quo,
	OpDivideInt: Context.QuoInt,
	OpMax: func(_ Context, _, x, y *Big) *Big {
		return Max(x, y)
	},
	OpMaxMag: func(_ Context, _, x, y *Big) *Big {
		return MaxAbs(x, y)
	},
	OpMin: func(_ Context, _, x, y *Big) *Big {
		return Min(x, y)
	},
	OpMinMag: func(_ Context, _, x, y *Big) *Big {
		return MinAbs(x, y)
	},
	OpMultiply:  Context.Mul,
	OpPower:     Context.Pow,
	OpRemainder: Context.Rem,
	OpSubtract:  Context.Sub,
}

var ternary = map[Op]func(ctx Context, z, x, y, u *Big) *Big{
	OpFMA: Context.FMA,
}

var decConditions = map[Condition]Condition{
	Clamped:             Clamped,
	ConversionSyntax:    ConversionSyntax,
	DivisionByZero:      DivisionByZero,
	DivisionImpossible:  DivisionImpossible,
	DivisionUndefined:   DivisionUndefined,
	Inexact:             Inexact,
	InsufficientStorage: InsufficientStorage,
	InvalidContext:      InvalidContext,
	InvalidOperation:    InvalidOperation,
	Overflow:            Overflow,
	Rounded:             Rounded,
	Subnormal:           Subnormal,
	Underflow:           Underflow,
}

func execute(t *testing.T, c *Case) {
	if c.MaxScale > MaxScale {
		t.Fatalf("invalid max scale: %d", c.MaxScale)
	}

	if c.MinScale < MinScale {
		t.Fatalf("invalid min scale: %d", c.MinScale)
	}

	if c.MinScale < MinScale {
		t.Fatalf("invalid min scale: %d", c.MinScale)
	}

	if c.Prec < MinPrecision || c.Prec > MaxPrecision {
		t.Fatalf("invalid precision: %d", c.Prec)
	}

	if _, ok := skip[c.ID]; ok {
		// Can't use t.Skip since it'll fail the entire category,
		// which we do not want.
		t.Logf("skipped test %s", c.ID)
		return
	}

	flags, ok := convertConditions(c.Conditions)
	if !ok {
		t.Fatalf("invalid condition(s): %s", c.Conditions)
	}

	ctx := Context{
		Precision:     c.Prec,
		OperatingMode: GDA,
		RoundingMode:  c.Mode,
		MinScale:      c.MinScale,
		MaxScale:      c.MaxScale,
	}

	z, x, y, u := parseInputs(ctx, c)
	r := parseOutput(ctx, c, flags)

	if nfn, ok := nilary[c.Op]; ok {
		check(t, nfn(ctx, x), r, c, flags)
	} else if ufn, ok := unary[c.Op]; ok {
		check(t, ufn(ctx, z, x), r, c, flags)
	} else if bfn, ok := binary[c.Op]; ok {
		check(t, bfn(ctx, z, x, y), r, c, flags)
	} else if tfn, ok := ternary[c.Op]; ok {
		check(t, tfn(ctx, z, x, y, u), r, c, flags)
	} else {
		switch c.Op {
		case OpClass:
			assert(t, c, x.Class(), r)
		case OpCompare:
			rv := x.Cmp(y)
			r, _, snan := cmp(t, c)
			assert(t, c, rv, r)
			assert(t, c, snan, x.Context.Conditions&InvalidOperation != 0)
		case OpCompareTotal:
			rv := x.CmpTotal(y)
			r, _, snan := cmp(t, c)
			assert(t, c, rv, r)
			assert(t, c, snan, x.Context.Conditions&InvalidOperation != 0)
		case OpCompareTotMag:
			rv := x.CmpTotalAbs(y)
			r, _, snan := cmp(t, c)
			assert(t, c, rv, r)
			assert(t, c, snan, x.Context.Conditions&InvalidOperation != 0)
		case OpMax:
			check(t, z.Set(Max(x, y)), r, c, flags)
		case OpMin:
			check(t, z.Set(Min(x, y)), r, c, flags)
		case OpQuantize:
			v, _ := y.Int64()
			if v > arith.MaxInt {
				t.Logf("%s: int out of range: %d", c.ID, v)
				return
			}
			check(t, x.Quantize(int(v)), r, c, flags)
		case OpSameQuantum:
			rv := x.SameQuantum(y)
			assert(t, c, rv, c.Output == Data("1"))
		case OpToSci:
			rv := fmt.Sprintf("%E", x)
			assert(t, c, rv, string(c.Output))
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
		case OpClass, OpCompare, OpCompareTotal, OpCompareTotMag,
			OpMax, OpMin, OpQuantize, OpSameQuantum, OpToSci:
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

func parseOutput(ctx Context, c *Case, f Condition) *Big {
	r := c.Output.toBig(ctx)
	r.Context.Conditions = f
	return r
}

func (d Data) toBig(ctx Context) *Big {
	if d == NoData {
		return New(0, 0).SetInf(false)
	}
	var z Big
	_, ok := z.SetString(string(d.TrimQuotes()))
	if !ok {
		z.SetNaN(true)
	}
	return &z
}

func parseInputs(ctx Context, c *Case) (z, x, y, u *Big) {
	z = new(Big)
	switch len(c.Inputs) {
	case 3:
		u = c.Inputs[2].toBig(ctx)
		fallthrough
	case 2:
		y = c.Inputs[1].toBig(ctx)
		fallthrough
	case 1:
		x = c.Inputs[0].toBig(ctx)
	case 0:
		break
	default:
		panic(fmt.Errorf("%s: invalid number of inputs (%d)", c.ID, len(c.Inputs)))
	}
	return
}

func assert(t *testing.T, c *Case, a, b interface{}) {
	t.Helper()

	if !reflect.DeepEqual(a, b) {
		t.Fatalf(`%s
wanted: %v
got   : %v
`, c.ShortString(22), b, a)
	}
}

func check(t *testing.T, z, r *Big, c *Case, flags Condition) {
	t.Helper()

	if !equal(z, r) {
		str := fmt.Sprintf(`%s
wanted: %q (%s:%d)
got   : %q (%s:%d)
`,
			c.ShortString(10000),
			r, flags, -r.Scale(),
			z, z.Context.Conditions, -z.Scale(),
		)
		t.Fatal(str)
	}
}

func cmp(t *testing.T, c *Case) (int, bool, bool) {
	t.Helper()

	qnan, snan := Data(c.Output).IsNaN()
	if qnan || snan {
		return 0, qnan, snan
	}
	r, err := strconv.Atoi(string(c.Output))
	if err != nil {
		t.Fatal(err)
	}
	return r, false, false
}

func equal(x, y *Big) bool {
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

func convertConditions(c Condition) (Condition, bool) {
	var r Condition
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
