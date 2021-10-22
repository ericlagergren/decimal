package decimal

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/ericlagergren/decimal/suite"
)

func TestPytables(t *testing.T) {
	for _, s := range []string{
		testAbs,
		testAdd,
		testCTR,
		testCTS,
		testCTS,
		testClass,
		testCmp,
		testExp,
		testFMA,
		testLog,
		testLog10,
		testMul,
		testNeg,
		testNextMinus,
		testNextPlus,
		testPow,
		testQuant,
		testQuo,
		testQuoInt,
		testReduce,
		testRem,
		testRoundToInt,
		testSign,
		testSignbit,
		testSqrt,
		testSub,
	} {
		t.Run(s, func(t *testing.T) {
			pytables(t, s)
		})
	}
}

const (
	testAbs        = "absolute-value"
	testAdd        = "addition"
	testClass      = "class"
	testCmp        = "comparison"
	testCTR        = "convert-to-rat"
	testCFS        = "convert-from-string"
	testCTS        = "convert-to-string"
	testExp        = "exponential-function"
	testFMA        = "fused-multiply-add"
	testLog10      = "common-logarithm"
	testLogb       = "base-b-logarithm"
	testLog        = "natural-logarithm"
	testMul        = "multiplication"
	testNeg        = "negation"
	testNextMinus  = "next-minus"
	testNextPlus   = "next-plus"
	testPow        = "power"
	testQuant      = "quantization"
	testQuo        = "division"
	testQuoInt     = "integer-division"
	testReduce     = "reduction"
	testRem        = "remainder"
	testRoundToInt = "round-to-integral-exact"
	testShift      = "shift"
	testSign       = "sign"
	testSignbit    = "signbit"
	testSub        = "subtraction"
	testSqrt       = "square-root"
)

func pytables(t *testing.T, name string) {
	s := open(t, name)
	for s.Next() {
		c := s.Case(t)
		c.execute(name)
	}
}

var nilaryTests = map[string]func(z *Big) *Big{
	testReduce:     (*Big).Reduce,
	testRoundToInt: (*Big).RoundToInt,
}

var unaryTests = map[string]func(ctx Context, z, x *Big) *Big{
	testAbs:       Context.Abs,
	testNeg:       Context.Neg,
	testExp:       Context.Exp,
	testLog:       Context.Log,
	testLog10:     Context.Log10,
	testNextMinus: Context.NextMinus,
	testNextPlus:  Context.NextPlus,
	testSqrt:      Context.Sqrt,
}

var binaryTests = map[string]func(ctx Context, z, x, y *Big) *Big{
	testAdd:    Context.Add,
	testMul:    Context.Mul,
	testQuo:    Context.Quo,
	testQuoInt: Context.QuoInt,
	testRem:    Context.Rem,
	testSub:    Context.Sub,
	// The Python version we test against has rounding errors of
	// 1 ULP. So test to see if we're within 1 ULP.
	// Pow:    math.Pow,
}

var ternaryTests = map[string]func(ctx Context, z, x, y, u *Big) *Big{
	testFMA: Context.FMA,
}

func (c *scase) execute(name string) {
	ctx := c.ctx

	if nfn, ok := nilaryTests[name]; ok {
		c.Check(nfn(c.x))
	} else if ufn, ok := unaryTests[name]; ok {
		c.Check(ufn(ctx, c.z, c.x))
	} else if bfn, ok := binaryTests[name]; ok {
		c.Check(bfn(ctx, c.z, c.x, c.y))
	} else if tfn, ok := ternaryTests[name]; ok {
		c.Check(tfn(ctx, c.z, c.x, c.y, c.u))
	} else {
		switch name {
		case testClass:
			c.Assert(c.x.Class(), c.r)
		case testCmp:
			rv := c.x.Cmp(c.y)
			r, _, snan := c.Cmp()
			c.Assert(rv, r)
			c.Assert(snan, c.x.Context.Conditions&InvalidOperation != 0)
		case testShift:
			c.t.Skip("TODO")
		case testQuant:
			v, _ := c.y.Int64()
			c.Check(c.x.Quantize(int(v)))
		case testCTR:
			r := new(big.Rat).SetFrac(c.x.Int(nil), c.y.Int(nil))
			c.Check(ctx.SetRat(c.z, r))
		case testSign:
			c.Assert(c.x.Sign(), c.Sign())
		case testCTS, testCFS:
			xs := c.x.String()
			if !Regexp.MatchString(xs) {
				c.t.Fatalf("should match regexp: %q", xs)
			}
			c.Assert(xs, c.r)
		case testPow:
			ctx.Pow(c.z, c.x, c.y)
			r := c.R()
			if !pytablesEqual(c.z, r) {
				diff := new(Big)
				eps := New(1, c.c.Prec)
				ctx := Context{Precision: -c.c.Prec}
				if ctx.Sub(diff, r, c.z).CmpAbs(eps) > 0 {
					c.t.Logf(`#%d: %s
wanted: %q (%s:%d)
got   : %q (%s:%d)
`,
						c.i, c.c.ShortString(22),
						r, c.flags, -r.Scale(),
						c.z, c.z.Context.Conditions, -c.z.Scale(),
					)
				}
			}
		case testSignbit:
			c.Assert(c.x.Signbit(), c.Signbit())
		default:
			panic("unknown test: " + name)
		}
	}
}

func open(t *testing.T, name string) (c *scanner) {
	fpath := filepath.Join("testdata", "pytables",
		fmt.Sprintf("%s-tables.gz", name))
	f, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	gzr, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	return &scanner{
		s: bufio.NewScanner(gzr),
		close: func() {
			gzr.Close()
			f.Close()
		},
	}
}

type scanner struct {
	i     int
	s     *bufio.Scanner
	close func()
}

func (c *scanner) Next() bool {
	if !c.s.Scan() {
		c.close()
		return false
	}
	c.i++
	return true
}

func (c *scanner) Case(t *testing.T) *scase {
	cs, err := suite.ParseCase(c.s.Bytes())
	if err != nil {
		panic(err)
	}
	return parse(t, cs, c.i)
}

func ctx(c suite.Case) Context {
	return Context{
		Precision:     c.Prec,
		OperatingMode: GDA,
		RoundingMode:  RoundingMode(c.Mode),
		Traps:         Condition(c.Trap),
	}
}

func parse(t *testing.T, c suite.Case, i int) *scase {
	ctx := ctx(c)
	s := scase{
		t:     t,
		ctx:   ctx,
		i:     i,
		c:     c,
		z:     new(Big),
		r:     string(c.Output),
		flags: Condition(c.Excep),
	}
	switch len(c.Inputs) {
	case 3:
		s.u, _ = WithContext(ctx).SetString(string(c.Inputs[2]))
		fallthrough
	case 2:
		s.y, _ = WithContext(ctx).SetString(string(c.Inputs[1]))
		fallthrough
	case 1:
		s.x, _ = WithContext(ctx).SetString(string(c.Inputs[0]))
	default:
		panic(fmt.Sprintf("%s\n%d inputs", s.c, len(c.Inputs)))
	}
	return &s
}

func (c *scase) Assert(got, want interface{}) {
	c.t.Helper()

	if !reflect.DeepEqual(got, want) {
		c.t.Fatalf(`#%d: %s
wanted: %v
got   : %v
`, c.i, c.c.ShortString(22), want, got)
	}
}

func (c *scase) Check(z *Big) {
	c.t.Helper()

	r := c.R()
	if !pytablesEqual(z, r) {
		c.t.Fatalf(`#%d: %s
wanted: %q (%s:%d)
got   : %q (%s:%d)
`,
			c.i, c.c.ShortString(10000),
			r, c.flags, -r.Scale(),
			z, z.Context.Conditions, -z.Scale(),
		)
	}
}

type scase struct {
	z, x, y, u *Big
	c          suite.Case
	i          int
	r          string
	t          *testing.T
	flags      Condition
	ctx        Context
}

func (s *scase) R() *Big {
	r, _ := WithContext(s.ctx).SetString(s.r)
	r.Context.Conditions = s.flags
	return r
}

func (s *scase) Str() string { return s.r }

func (s *scase) Sign() int {
	s.t.Helper()

	r, err := strconv.Atoi(s.r)
	if err != nil {
		s.t.Fatal(err)
	}
	return r
}

func (s *scase) Cmp() (int, bool, bool) {
	s.t.Helper()

	qnan, snan := suite.Data(s.r).IsNaN()
	if qnan || snan {
		return 0, qnan, snan
	}
	r, err := strconv.Atoi(s.r)
	if err != nil {
		s.t.Fatal(err)
	}
	return r, false, false
}

func (s *scase) Signbit() bool {
	s.t.Helper()

	r, err := strconv.ParseBool(s.r)
	if err != nil {
		s.t.Fatal(err)
	}
	return r
}

func pytablesEqual(x, y *Big) bool {
	// Python doesn't have DivisionUndefined.
	x.Context.Conditions &^= DivisionUndefined
	return equal(x, y)
}
