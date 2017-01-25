package decimal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ericlagergren/decimal/suite"
)

func TestSuiteCases(t *testing.T) {
	testdir := filepath.Join("suite", "tests")
	dir, err := os.Open(testdir)
	if err != nil {
		t.Fatal(err)
	}
	names, err := dir.Readdirnames(-1)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(names)

	sets := make(map[string][]suite.Case)
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		file, err := os.Open(filepath.Join(testdir, name))
		if err != nil {
			t.Fatal(err)
		}

		var c []suite.Case
		err = json.NewDecoder(file).Decode(&c)
		if err != nil {
			file.Close()
			t.Fatal(err)
		}
		sets[name] = c
		file.Close()
	}

	// Loop over names instead of sets so our tests run in the same order.
	// Makes debugging easier.
	for _, name := range names {
		t.Log(name)
		for i, cs := range sets[name] {
			// +1 makes debugging easier since file lines are 1-indexed.
			testCase(name, i+1, cs, t)
		}
	}
}

func badTest(c suite.Case) bool {
	for _, v := range c.Inputs {
		switch v {
		// Don't test quiet and signaling NaNs since we do not allow
		// decimals to be created as NaN.
		case "Q", "S":
			return true
		}
	}
	switch c.Output {
	// This means an exception probably happened.
	// TODO: handle this later.
	case "#":
		return true
	}
	return false
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

func testCase(fname string, i int, c suite.Case, t *testing.T) {
	if badTest(c) {
		return
	}

	z := new(Big)
	z.SetMode(RoundingMode(c.Mode))
	z.SetPrecision(50)

	// We need to expect an exception if:
	//
	// 1. The output from the test is NaN (Q, S)
	// 2. The test expects to receive a DivByZero or Invalid exception
	//
	nan, _ := c.Output.IsNaN()
	except := nan ||
		(c.Excep == suite.DivByZero || c.Excep == suite.Invalid)

	// Some tests trap an exception.
	trap := c.Trap == suite.DivByZero || c.Trap == suite.Invalid

	compare := func() {
		// If nan is true then the test output Q or S. Skip it since our
		// results are undefined in those cases.
		if nan {
			return
		}

		want, _ := new(Big).SetString(string(c.Output))
		want.Round(precision(c.Output))
		z.Round(int32(want.Prec()))
		if want.Cmp(z) != 0 {
			if testing.Verbose() {
				t.Log(precision(c.Output), ":", c.Output)
				t.Logf(`%s#%d: %s
wanted: %s
got   : %s
`, fname, i, c, want, z)
			}
		}
	}

	if except || trap {
		defer func() {
			_, ok := recover().(error)
			// We panicked but did not want to. Sometimes the test
			// files will trap an exception. We don't honor traps,
			// meaning we'll always panic.
			if ok && !except && !trap {
				t.Fatal("shouldn't have panicked")
			}
			// We didn't panic but wanted to.
			if !ok && except {
				t.Fatal("wanted to panic but didn't")
			}
			// We correctly caught a panic. Still, compare the output since
			// we need to correctly set the receiver's form.
			compare()
		}()
	}

	switch c.Op {
	case suite.Add:
		binaryBig(z.Add, c.Inputs[0], c.Inputs[1])
	case suite.Sub:
		binaryBig(z.Sub, c.Inputs[0], c.Inputs[1])
	case suite.Mul:
		binaryBig(z.Mul, c.Inputs[0], c.Inputs[1])
	case suite.Div:
		binaryBig(z.Quo, c.Inputs[0], c.Inputs[1])
	case suite.Neg:
	}

	if !except {
		compare()
	}
}

func binaryBig(fn func(x, y *Big) *Big, arg1, arg2 suite.Data) {
	x, _ := new(Big).SetString(string(arg1))
	y, _ := new(Big).SetString(string(arg2))
	fn(x, y)
}
