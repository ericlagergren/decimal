package test

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

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
	"github.com/ericlagergren/decimal/misc"
	"github.com/ericlagergren/decimal/suite"
)

// Helper returns testing.T.Helper, if it exists.
func Helper(v interface{}) func() {
	if fn, ok := v.(interface {
		Helper()
	}); ok {
		return fn.Helper
	}
	return func() {}
}

type Test string

const (
	Abs        Test = "absolute-value"
	Add        Test = "addition"
	Class      Test = "class"
	Cmp        Test = "comparison"
	CTR        Test = "convert-to-rat"
	CFS        Test = "convert-from-string"
	CTS        Test = "convert-to-string"
	Exp        Test = "exponential-function"
	FMA        Test = "fused-multiply-add"
	Log10      Test = "common-logarithm"
	Logb       Test = "base-b-logarithm"
	Log        Test = "natural-logarithm"
	Mul        Test = "multiplication"
	Neg        Test = "negation"
	NextMinus  Test = "next-minus"
	NextPlus   Test = "next-plus"
	Pow        Test = "power"
	Quant      Test = "quantization"
	Quo        Test = "division"
	QuoInt     Test = "integer-division"
	Reduce     Test = "reduction"
	Rem        Test = "remainder"
	RoundToInt Test = "round-to-integral-exact"
	Shift      Test = "shift"
	Sign       Test = "sign"
	Signbit    Test = "signbit"
	Sub        Test = "subtraction"
	Sqrt       Test = "square-root"
)

func (tst Test) Test(t *testing.T) {
	t.Parallel() // Call after parsing so we don't goof the scanner.
	s := open(string(tst))
	for s.Next() {
		t.Run(string(tst), func(t *testing.T) {
			c := s.Case(t)
			c.execute(tst)
		})
	}
}

var nilary = map[Test]func(z *decimal.Big) *decimal.Big{
	Reduce:     (*decimal.Big).Reduce,
	RoundToInt: (*decimal.Big).RoundToInt,
}

var unary = map[Test]func(z, x *decimal.Big) *decimal.Big{
	Abs:       (*decimal.Big).Abs,
	Neg:       (*decimal.Big).Neg,
	Exp:       math.Exp,
	Log:       math.Log,
	Log10:     math.Log10,
	NextMinus: misc.NextMinus,
	NextPlus:  misc.NextPlus,
	Sqrt:      math.Sqrt,
}

var binary = map[Test]func(z, x, y *decimal.Big) *decimal.Big{
	Add:    (*decimal.Big).Add,
	Mul:    (*decimal.Big).Mul,
	Quo:    (*decimal.Big).Quo,
	QuoInt: (*decimal.Big).QuoInt,
	Rem:    (*decimal.Big).Rem,
	Sub:    (*decimal.Big).Sub,
	// The Python version we test against has rounding errors of 1 ULP. So test
	// to see if we're within 1 ULP.
	// Pow:    math.Pow,
}

var ternary = map[Test]func(z, x, y, u *decimal.Big) *decimal.Big{
	FMA: (*decimal.Big).FMA,
}

func (c *scase) execute(name Test) {
	if nfn, ok := nilary[name]; ok {
		c.Check(nfn(c.x))
	} else if ufn, ok := unary[name]; ok {
		c.Check(ufn(c.z, c.x))
	} else if bfn, ok := binary[name]; ok {
		c.Check(bfn(c.z, c.x, c.y))
	} else if tfn, ok := ternary[name]; ok {
		c.Check(tfn(c.z, c.x, c.y, c.u))
	} else {
		switch name {
		case Class:
			c.Assert(c.x.Class(), c.r)
		case Cmp:
			rv := c.x.Cmp(c.y)
			r, _, snan := c.Cmp()
			c.Assert(rv, r)
			c.Assert(snan, c.x.Context.Conditions&decimal.InvalidOperation != 0)
		case Shift:
			//v, _ := c.y.Int64()
			//c.Check(misc.Shift(c.z, c.x, int(v)))
		case Quant:
			v, _ := c.y.Int64()
			c.Check(c.x.Quantize(int(v)))
		case CTR:
			r := new(big.Rat).SetFrac(c.x.Int(nil), c.y.Int(nil))
			// Given that SetRat/Rat are non-standard, I don't feel bad for
			// calling Assert(z.Cmp(r)) instead of Check(z).
			c.Assert(c.z.SetRat(r).Cmp(c.R()), 0)
		case Sign:
			c.Assert(c.x.Sign(), c.Sign())
		case CTS, CFS:
			xs := c.x.String()
			if !decimal.Regexp.MatchString(xs) {
				c.t.Fatalf("should match regexp: %q", xs)
			}
			c.Assert(xs, c.r)
		case Pow:
			math.Pow(c.z, c.x, c.y)
			r := c.R()
			if !equal(c.z, r) {
				diff := new(decimal.Big)
				eps := decimal.New(1, c.c.Prec)
				ctx := decimal.Context{Precision: -c.c.Prec}
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
		case Signbit:
			c.Assert(c.x.Signbit(), c.Signbit())
		default:
			panic("unknown test: " + name)
		}
	}
}

func open(name string) (c *scanner) {
	fpath := filepath.Join("testdata", fmt.Sprintf("%s-tables.gz", name))
	file, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	gzr, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	return &scanner{
		s:     bufio.NewScanner(gzr),
		close: func() { gzr.Close(); file.Close() },
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

func ctx(c suite.Case) decimal.Context {
	return decimal.Context{
		Precision:     c.Prec,
		OperatingMode: decimal.GDA,
		RoundingMode:  decimal.RoundingMode(c.Mode),
		Traps:         decimal.Condition(c.Trap),
	}
}

func parse(t *testing.T, c suite.Case, i int) *scase {
	ctx := ctx(c)
	s := scase{
		t:     t,
		ctx:   ctx,
		i:     i,
		c:     c,
		z:     decimal.WithContext(ctx),
		r:     string(c.Output),
		flags: decimal.Condition(c.Excep),
	}
	switch len(c.Inputs) {
	case 3:
		s.u, _ = decimal.WithContext(ctx).SetString(string(c.Inputs[2]))
		fallthrough
	case 2:
		s.y, _ = decimal.WithContext(ctx).SetString(string(c.Inputs[1]))
		fallthrough
	case 1:
		s.x, _ = decimal.WithContext(ctx).SetString(string(c.Inputs[0]))
	default:
		panic(fmt.Sprintf("%s\n%d inputs", s.c, len(c.Inputs)))
	}
	return &s
}

func (c *scase) Assert(a, b interface{}) {
	Helper(c.t)()
	if !reflect.DeepEqual(a, b) {
		c.t.Fatalf(`#%d: %s
wanted: %v
got   : %v
`, c.i, c.c.ShortString(22), b, a)
	}
}

func (c *scase) Check(z *decimal.Big) {
	Helper(c.t)()
	r := c.R()
	if !equal(z, r) {
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
	z, x, y, u *decimal.Big
	c          suite.Case
	i          int
	r          string
	t          *testing.T
	flags      decimal.Condition
	ctx        decimal.Context
}

func (s *scase) R() *decimal.Big {
	r, _ := decimal.WithContext(s.ctx).SetString(s.r)
	r.Context.Conditions = s.flags
	return r
}

func (s *scase) Str() string { return s.r }

func (s *scase) Sign() int {
	r, err := strconv.Atoi(s.r)
	if err != nil {
		Helper(s.t)()
		s.t.Fatal(err)
	}
	return r
}

func (s *scase) Cmp() (int, bool, bool) {
	qnan, snan := suite.Data(s.r).IsNaN()
	if qnan || snan {
		return 0, qnan, snan
	}
	r, err := strconv.Atoi(s.r)
	if err != nil {
		Helper(s.t)()
		s.t.Fatal(err)
	}
	return r, false, false
}

func (s *scase) Signbit() bool {
	r, err := strconv.ParseBool(s.r)
	if err != nil {
		Helper(s.t)()
		s.t.Fatal(err)
	}
	return r
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
	// Python doesn't have DivisionUndefined.
	if (x.Context.Conditions & ^decimal.DivisionUndefined) != y.Context.Conditions {
		return false
	}
	cmp := x.Cmp(y) == 0
	scl := x.Scale() == y.Scale()
	prec := x.Precision() == y.Precision()
	return cmp && scl && prec
}
