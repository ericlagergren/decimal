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
		for i, cs := range sets[name] {
			// +1 makes debugging easier since file lines are 1-indexed.
			testCase(name, i+1, cs, t)
		}
	}
}

func badTest(c suite.Case) bool {
	for _, v := range c.Inputs {
		switch v {
		// Don't test quiet and signaling NaNs since we panic on any NaN
		// values.
		case "Q", "S":
			return true
		}
	}
	switch c.Output {
	// This means an exception happened.
	// TODO: handle this later.
	case "#":
		return true
	}
	return false
}

func testCase(fname string, i int, c suite.Case, t *testing.T) {
	if badTest(c) {
		return
	}
	t.Logf("(%s) #%d: %s\n", fname, i, c)
	z := new(Big)
	z.SetMode(RoundingMode(c.Mode))

	except := c.Excep == suite.DivByZero || c.Excep == suite.Invalid
	trap := c.Trap == suite.DivByZero || c.Trap == suite.Invalid
	if except || trap {
		defer func() {
			_, ok := recover().(error)
			// We panicked but did not want to. Sometimes the test files will
			// trap an exception. We don't honor traps, meaning we'll always
			// panic.
			if ok && !except && !trap {
				t.Fatal("shouldn't have panicked")
			}
			// We didn't panic but wanted to.
			if !ok && except {
				t.Fatal("wanted to panic but didn't")
			}
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
		want, _ := new(Big).SetString(string(c.Output))
		if want.Cmp(z) != 0 {
			t.Logf("wanted: %s\ngot   : %s\n", want, z)
		}
	}
}

func binaryBig(fn func(x, y *Big) *Big, arg1, arg2 suite.Data) {
	x, _ := new(Big).SetString(string(arg1))
	y, _ := new(Big).SetString(string(arg2))
	fn(x, y)
}
