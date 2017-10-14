package decimal

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ericlagergren/decimal/suite"
)

func TestSuiteCases(t *testing.T) {
	if testing.Short() {
		return
	}

	file, err := os.Open(filepath.Join("suite", "_testdata", "json.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	sets := make(map[string][]suite.Case)
	var names []string
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		var c []suite.Case
		if err := json.NewDecoder(tr).Decode(&c); err != nil {
			t.Fatal(err)
		}
		sets[h.Name] = c
		names = append(names, h.Name)
	}

	for _, mode := range [...]OperatingMode{GDA, Go} {
		// Loop over names instead of sets so our tests run in the same order.
		// Makes debugging easier.
		for _, name := range names {
			for i, cs := range sets[name] {
				// +1 makes debugging easier since file lines are 1-indexed.
				testCase(name, i+1, cs, mode, t)
			}
		}
	}
}

func precision(s suite.Data) (p int32) {
	j := strings.IndexAny(string(s), "eE")
	if j < 0 {
		j = len(s)
	}
	for i := 0; i < j; i++ {
		if s[i] >= '0' && s[i] < '9' {
			p++
		}
	}
	return p
}

var convs = [...]struct {
	e suite.Exception
	c Condition
}{
	{suite.Inexact, Inexact},
	{suite.Underflow, Underflow},
	{suite.Overflow, Overflow},
	{suite.DivByZero, DivisionByZero},
	{suite.Invalid, InvalidOperation},
}

func convException(e suite.Exception) (c Condition) {
	for _, pair := range convs {
		if e&pair.e != 0 {
			c |= pair.c
		}
	}
	return c
}

func testCase(fname string, i int, c suite.Case, mode OperatingMode, t *testing.T) {
	switch fname {
	case "Underflow.json":
		return
	default:
	}

	z := new(Big)
	z.Context.RoundingMode = RoundingMode(c.Mode)
	z.Context.SetPrecision(50)
	z.Context.OperatingMode = mode
	z.Context.Traps = convException(c.Trap)

	var (
		cond Condition
		err  error
		args = make([]*Big, len(c.Inputs))
	)
	for i, data := range c.Inputs {
		args[i] = dataToBig(data, z.Context)
	}

	func() {
		defer func() {
			if e, ok := recover().(error); ok {
				err = e
			} else {
				err = z.Context.Err
			}
			cond = z.Context.Conditions
		}()
		switch c.Op {
		case suite.Add:
			z.Add(args[0], args[1])
		case suite.Sub:
			z.Sub(args[0], args[1])
		case suite.Mul:
			z.Mul(args[0], args[1])
		case suite.Div:
			z.Quo(args[0], args[1])
		case suite.Neg:
			z.Neg(args[0])
		}
	}()

	if testing.Verbose() {
		t.Logf("%s: %s => [%e, %q, %v]", mode, c, z, cond, err)
	}

	want := dataToBig(c.Output, z.Context)
	if want != nil {
		z.Round(int32(want.Precision()))
	}

	// fpgen doesn't test for Rounded.
	cond &= ^Rounded

	wantConds := convException(c.Excep)
	if wantConds != cond {
		// Since we can accept decimals of arbitrary size, we can handle larger
		// decimals than the fpgen test suite. These need to be manually checked
		// if they're division. Arbitrary precision decimals aren't lossy for
		// add, sub, etc.
		msg := fmt.Sprintf("%s#%d: wanted %q, got %q", fname, i, wantConds, cond)
		if (Inexact|Overflow)&wantConds != 0 {
			if c.Op == suite.Div {
				t.Logf("CHECK: %s", msg)
			}
		} else if mode != Go {
			t.Fatalf(msg)
		}
	}

	if werr, zerr := want.Context.Err, z.Context.Err; werr != zerr {
		t.Fatalf("%s#%d: wanted %v, got %v", fname, i, werr, zerr)
	}

	badNaN := want.IsNaN(+1) != z.IsNaN(+1) || want.IsNaN(-1) != z.IsNaN(-1)
	nancmp := want.IsNaN(0) || z.IsNaN(0)

	if badNaN || (!nancmp && want.Cmp(z) != 0) {
		msg := fmt.Sprintf(`%s#%d: %s
wanted: "%e"
got   : "%e"
`, fname, i, c, want, z)

		badInexact := Inexact&wantConds != 0
		if badInexact {
			if _, badInexact = c.Output.IsInf(); !badInexact {
				badInexact = want.Cmp(testZero) == 0
			}
		}

		if want.Signbit() == z.Signbit() &&
			(badInexact || wantConds&(Overflow|Underflow) != 0) {
			t.Logf("CHECK: %s", msg)
		} else {
			t.Fatal(msg)
		}
	}
}

var testZero = New(0, 0)

func makeNaN(signal bool, ctx Context) *Big {
	z := new(Big)
	z.Context = ctx
	if signal {
		z.SetString("snan")
	} else {
		z.SetString("qnan")
	}
	return z
}

func dataToBig(s suite.Data, ctx Context) *Big {
	switch s {
	case "Q", "S", suite.NoData:
		return makeNaN(s == "S", ctx)
	default:
		x := new(Big)
		x.Context = ctx
		if _, ok := x.SetString(string(s)); !ok {
			panic(fmt.Sprintf("couldn't SetString(%q)", s))
		}
		return x
	}
}
