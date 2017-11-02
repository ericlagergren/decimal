package decimal

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ericlagergren/decimal/suite"
)

func getTests(t *testing.T, name string) (s *bufio.Scanner, close func()) {
	t.Helper()

	fpath := filepath.Join("_testdata", fmt.Sprintf("%s-tables.gz", name))
	file, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	gzr, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	return bufio.NewScanner(gzr), func() { gzr.Close(); file.Close() }
}

func didPanic(f func()) (ok bool) {
	defer func() { ok = recover() != nil }()
	f()
	return ok
}

func newbig(t *testing.T, s string) *Big {
	x, ok := new(Big).SetString(s)
	if !ok {
		if t == nil {
			panic("wanted true got false during set")
		}
		t.Fatal("wanted true got false during set")
	}
	testFormZero(t, x, "newbig")
	return x
}

var bigZero = new(Big)

// testFormZero verifies that if z == 0, z.form == zero.
func testFormZero(t *testing.T, z *Big, name string) {
	iszero := z.Cmp(bigZero) == 0
	if iszero && z.form > nzero {
		t.Fatalf("%s: z == 0, but form not marked zero: %v", name, z.form)
	}
	if !iszero && z.form == zero {
		t.Fatalf("%s: z != 0, but form marked zero", name)
	}
}

func TestBig_Abs(t *testing.T) {
	s, close := getTests(t, "abs")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		z := new(Big)
		z.Context.SetPrecision(c.Prec)
		z.Context.OperatingMode = GDA
		z.Context.RoundingMode = RoundingMode(c.Mode)

		x, _ := new(Big).SetString(string(c.Inputs[0]))
		r, _ := new(Big).SetString(string(c.Output))

		z.Abs(x)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: %s
wanted: %q
got   : %q
`, i+1, c, r, z)
		}
	}
}

func TestBig_Add(t *testing.T) {
	s, close := getTests(t, "addition")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		z := new(Big)
		z.Context.SetPrecision(c.Prec)
		z.Context.OperatingMode = GDA
		z.Context.RoundingMode = RoundingMode(c.Mode)
		x, _ := new(Big).SetString(string(c.Inputs[0]))
		y, _ := new(Big).SetString(string(c.Inputs[1]))
		r, _ := new(Big).SetString(string(c.Output))

		z.Add(x, y)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: %s
wanted: %q
got   : %q
`, i+1, c, r, z)
		}
	}
}

func TestBig_Cmp(t *testing.T) {
	s, close := getTests(t, "comparison")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		x := new(Big)
		x.Context.SetPrecision(c.Prec)
		x.Context.OperatingMode = GDA
		x.Context.RoundingMode = RoundingMode(c.Mode)
		x.SetString(string(c.Inputs[0]))

		y, _ := new(Big).SetString(string(c.Inputs[1]))

		var r int
		isnan, sig := c.Output.IsNaN()
		if !isnan {
			r, err = strconv.Atoi(string(c.Output))
			if err != nil {
				t.Fatal(err)
			}
		}

		rv := x.Cmp(y)
		bad := rv != r || (sig && (x.Context.Err == nil ||
			x.Context.Conditions&InvalidOperation == 0))
		if bad {
			t.Fatalf(`#%d: %s
wanted: %d (%s)
got   : %d (%v:%s)
`, i+1, c, r, c.Output, rv, x.Context.Err, x.Context.Conditions)
		}
	}
}

func TestBig_Float(t *testing.T) {
	for i, test := range [...]string{
		"42", "3.14156", "23423141234", ".44444", "1e+1222", "12e-444", "0",
	} {
		flt, ok := new(big.Float).SetString(test)
		if !ok {
			t.Fatal("!ok")
		}
		fv := new(big.Float).SetPrec(flt.Prec())
		xf := new(Big).SetFloat(flt).Float(fv)
		if xf.String() != flt.String() {
			t.Fatalf("#%d: wanted %f, got %f", i, flt, xf)
		}
	}
}

func TestBig_IsBig(t *testing.T) {
	for i, test := range [...]struct {
		a   *Big
		big bool
	}{
		0: {newbig(t, "100"), false},
		1: {newbig(t, "-100"), false},
		2: {newbig(t, "5000"), false},
		3: {newbig(t, "-5000"), false},
		4: {newbig(t, "9999999999999999999999999999"), true},
		5: {newbig(t, "1000.5000E+500"), true},
		6: {newbig(t, "1000.5000E-500"), true},
		7: {newbig(t, "+Inf"), false},
		8: {newbig(t, "0"), false},
	} {
		if ib := test.a.IsBig(); ib != test.big {
			t.Fatalf("#%d: wanted %t, got %t", i, test.big, ib)
		}
	}
}

func TestBig_Int(t *testing.T) {
	for i, test := range [...]string{
		"1.234", "4.567", "11111111111111111111111111111111111.2",
		"1234234.2321", "121111111111", "44444444.241", "1241.1",
		"4", "5123", "1.2345123134123414123123213", "0.11", ".1",
	} {
		a, ok := new(Big).SetString(test)
		if !ok {
			t.Fatalf("#%d: !ok", i)
		}
		iv := test
		switch x := strings.IndexByte(test, '.'); {
		case x > 0:
			iv = test[:x]
		case x == 0:
			iv = "0"
		}
		n := a.Int(nil)
		if n.String() != iv {
			t.Fatalf("#%d: wanted %q, got %q", i, iv, n.String())
		}
	}
}

func TestBig_Int64(t *testing.T) {
	for i, test := range [...]string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
		"100", "200", "300", "400", "500", "600", "700", "800", "900",
		"1000", "2000", "4000", "5000", "6000", "7000", "8000", "9000",
		"1000000", "2000000", "-12", "-500", "-13123213", "12.000000",
	} {
		a, ok := new(Big).SetString(test)
		if !ok {
			t.Fatalf("#%d: !ok", i)
		}
		iv := test
		switch x := strings.IndexByte(test, '.'); {
		case x > 0:
			iv = test[:x]
		case x == 0:
			iv = "0"
		}
		n := a.Int64()
		if ns := strconv.FormatInt(n, 10); ns != iv {
			t.Fatalf("#%d: wanted %q, got %q", i, iv, ns)
		}
	}
}

func TestBig_IsInt(t *testing.T) {
	for i, test := range [...]string{
		"0 int",
		"-0 int",
		"1 int",
		"-1 int",
		"0.0120",
		"444.000 int",
		"10.000 int",
		"1.0001e+33333 int",
		"0.5",
		"0.011",
		"1.23",
		"1.23e1",
		"1.23e2 int",
		"0.000000001e+8",
		"0.000000001e+9 int",
		"1.2345e200 int",
		"Inf",
		"+Inf",
		"-Inf",
		"-inf",
	} {
		s := strings.TrimSuffix(test, " int")
		x, ok := new(Big).SetString(s)
		if !ok {
			t.Fatal("TestBig_IsInt !ok")
		}
		want := s != test
		if got := x.IsInt(); got != want {
			t.Fatalf("#%d: (%q).IsInt() == %t", i, s, got)
		}
	}
}

// func TestBig_Format(t *testing.T) {
// 	tests := [...]struct {
// 		format string
// 		a      string
// 		b      string
// 	}{
// 		0: {format: "%e", a: "1.234", b: "1.234"},
// 		1: {format: "%s", a: "1.2134124124", b: "1.2134124124"},
// 		2: {format: "%e", a: "1.00003e-12", b: "1.00003e-12"},
// 		3: {format: "%E", a: "1.000003E-12", b: "1.000003E-12"},
// 	}
// 	for i, v := range tests {
// 		x, ok := new(Big).SetString(v.a)
// 		if !ok {
// 			t.Fatal("invalid SetString")
// 		}
// 		if fs := fmt.Sprintf(v.format, x); fs != v.b {
// 			t.Fatalf("#%d: wanted %q, got %q:", i, v.b, fs)
// 		}
// 	}
// }

func TestBig_Neg(t *testing.T) {
	tests := [...]struct {
		a, b *Big
	}{
		0: {a: New(1, 0), b: New(-1, 0)},
		1: {a: New(999999999999, -1000), b: New(-999999999999, -1000)},
		2: {a: New(-512, 2), b: New(512, 2)},
	}
	var b Big
	for i, v := range tests {
		b.Neg(v.a)

		bs := v.b.String()
		if gs := b.String(); gs != bs {
			t.Fatalf("#%d: wanted %q, got %q", i, bs, gs)
		}
	}
}

func TestBig_Mul(t *testing.T) {
	s, close := getTests(t, "multiplication")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		z := new(Big)
		z.Context.SetPrecision(c.Prec)
		z.Context.OperatingMode = GDA
		z.Context.RoundingMode = RoundingMode(c.Mode)
		x, _ := new(Big).SetString(string(c.Inputs[0]))
		y, _ := new(Big).SetString(string(c.Inputs[1]))
		r, _ := new(Big).SetString(string(c.Output))

		z.Mul(x, y)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: %s
wanted: %q
got   : %q
`, i+1, c, r, z)
		}
	}
}

func TestBig_Prec(t *testing.T) {
	// confirmed to work inside internal/arith/intlen_test.go
}

func TestBig_Quantize(t *testing.T) {
	s, close := getTests(t, "quantize")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		x, _ := new(Big).SetString(string(c.Inputs[0]))
		x.Context.SetPrecision(c.Prec)
		x.Context.OperatingMode = GDA
		x.Context.RoundingMode = RoundingMode(c.Mode)

		y, err := strconv.ParseInt(string(c.Inputs[1]), 10, 32)
		if err != nil {
			t.Fatal(err)
		}

		r, _ := new(Big).SetString(string(c.Output))

		x.Quantize(int32(y))
		if x.Cmp(r) != 0 {
			t.Fatalf(`#%d: %s
wanted: "%g"
got   : "%g"
`, i+1, c, r, x)
		}
	}
}

func TestBig_Quo(t *testing.T) {
	s, close := getTests(t, "division")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		z := new(Big)
		z.Context.SetPrecision(c.Prec)
		z.Context.OperatingMode = GDA
		z.Context.RoundingMode = RoundingMode(c.Mode)
		x, _ := new(Big).SetString(string(c.Inputs[0]))
		y, _ := new(Big).SetString(string(c.Inputs[1]))
		r, _ := new(Big).SetString(string(c.Output))

		z.Quo(x, y)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: %s
wanted: %q
got   : %q
`, i+1, c, r, z)
		}
	}
}

func TestBig_Rat(t *testing.T) {
	s, close := getTests(t, "convert-to-rat")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		var n, d big.Int
		n.SetString(string(c.Inputs[0]), 10)
		d.SetString(string(c.Inputs[1]), 10)
		r := new(big.Rat).SetFrac(&n, &d)

		x := new(Big)
		x.Context.SetPrecision(c.Prec)
		x.Context.OperatingMode = GDA
		x.Context.RoundingMode = RoundingMode(c.Mode)
		xr := x.SetRat(r).Rat(nil)

		if xr.Cmp(r) != 0 {
			t.Fatalf(`#%d: %s
wanted: %q
got   : %q
`, i+1, c, r, xr)
		}
	}
}

func TestBig_Round(t *testing.T) {
	for i, test := range [...]struct {
		v   string
		to  int
		res string
	}{
		0: {"5.5", 1, "6"},
		1: {"1.234", 2, "1.2"},
		2: {"1", 1, "1"},
		3: {"9.876", 0, "9.876"},
		4: {"5.65", 2, "5.6"},
		5: {"5.0002", 2, "5"},
		6: {"0.000158674", 6, "0.000158674"},
		7: {"1.58089722856961873690377135139876745465351534188711107066818e+12288", 50, "1.5808972285696187369037713513987674546535153418871e+12288"},
	} {
		bd := newbig(t, test.v)
		r := newbig(t, test.res)
		if bd.Round(test.to).Cmp(r) != 0 {
			t.Fatalf(`#%d:
wanted: %q
got   : %q
`, i, test.res, bd)
		}
	}
}

func TestBig_Scan(t *testing.T) {
	// TODO(eric): this
}

func TestBig_SetFloat64(t *testing.T) {
	z := new(Big)
	z.Context.OperatingMode = GDA
	z.Context.SetPrecision(25)

	var start, end uint32
	if testing.Short() {
		start = math.MaxUint32 / 4
		end = start * 3
	}
	for x := start; x <= end; x++ {
		f := float64(math.Float32frombits(x))
		z.SetFloat64(f)
		zf := z.Float64()
		if zf != f && (!math.IsNaN(f) && !math.IsNaN(zf)) {
			t.Fatalf(`#%d:
wanted: %g
got   : %g
`, x, f, zf)
		}
	}
}

func TestBig_SetString(t *testing.T) {
	// Tested in TestBig_String
}

func TestBig_Sign(t *testing.T) {
	s, close := getTests(t, "sign")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		x := new(Big)
		x.Context.SetPrecision(c.Prec)
		x.Context.OperatingMode = GDA
		x.Context.RoundingMode = RoundingMode(c.Mode)
		x.SetString(string(c.Inputs[0]))

		r, err := strconv.Atoi(string(c.Output))
		if err != nil {
			t.Fatal(err)
		}

		s := x.Sign()
		if s != r {
			t.Fatalf(`#%d: %s
wanted: %d
got   : %d
`, i+1, c, r, s)
		}
	}
}

func TestBig_SignBit(t *testing.T) {
	s, close := getTests(t, "signbit")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		x := new(Big)
		x.Context.SetPrecision(c.Prec)
		x.Context.OperatingMode = GDA
		x.Context.RoundingMode = RoundingMode(c.Mode)
		x.SetString(string(c.Inputs[0]))

		r, err := strconv.ParseBool(string(c.Output))
		if err != nil {
			t.Fatal(err)
		}

		sb := x.Signbit()
		if sb != r {
			t.Fatalf(`#%d: %s
wanted: %t
got   : %t
`, i+1, c, r, sb)
		}
	}
}

func TestBig_String(t *testing.T) {
	s, close := getTests(t, "convert-to-string")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		x := new(Big)
		x.Context.SetPrecision(c.Prec)
		x.Context.OperatingMode = GDA
		x.Context.RoundingMode = RoundingMode(c.Mode)
		x.SetString(string(c.Inputs[0]))

		r := string(c.Output)

		xs := x.String()
		if xs != r {
			t.Fatalf(`#%d: %s
wanted: %q
got   : %q
`, i+1, c, r, xs)
		}
	}
}

func TestBig_Sub(t *testing.T) {
	s, close := getTests(t, "subtraction")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		z := new(Big)
		z.Context.SetPrecision(c.Prec)
		z.Context.OperatingMode = GDA
		z.Context.RoundingMode = RoundingMode(c.Mode)
		x, _ := new(Big).SetString(string(c.Inputs[0]))
		y, _ := new(Big).SetString(string(c.Inputs[1]))
		r, _ := new(Big).SetString(string(c.Output))

		z.Sub(x, y)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: %s
wanted: %q
got   : %q
`, i+1, c, r, z)
		}
	}
}
