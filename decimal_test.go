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

	"github.com/ericlagergren/decimal/internal/c"
	"github.com/ericlagergren/decimal/suite"
)

func getTests(t *testing.T, name string) (s *bufio.Scanner, close func()) {
	t.Helper()

	fpath := filepath.Join("_testdata", fmt.Sprintf("%s-tables.gzip", name))
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

// Verify that ErrNaN implements the error interface.
var _ error = ErrNaN{}

func TestBig_Abs(t *testing.T) {
	for i, test := range [...]string{"-1", "1", "50", "-50", "0", "-0"} {
		x := newbig(t, test)
		if test[0] == '-' {
			test = test[1:]
		}
		if xs := x.Abs(x).String(); xs != test {
			t.Fatalf("#%d: wanted %s, got %s", i, test, xs)
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
		z.Context.SetPrecision(int32(c.Prec))
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

func TestBig_BitLen(t *testing.T) {
	var x Big
	const maxCompact = (1<<63 - 1) - 1
	tests := [...]struct {
		a *Big
		b int
	}{
		0:  {a: New(0, 0), b: 0},
		1:  {a: New(12, 0), b: 4},
		2:  {a: New(50, 0), b: 6},
		3:  {a: New(12345, 0), b: 14},
		4:  {a: New(123456789, 0), b: 27},
		5:  {a: New(maxCompact, 0), b: 63},
		6:  {a: x.Add(New(maxCompact, 0), New(maxCompact, 0)), b: 64},
		7:  {a: New(1000, 0), b: 10},
		8:  {a: New(10, -2), b: 10},
		9:  {a: New(1e6, 0), b: 20},
		10: {a: New(10, -5), b: 20},
		11: {a: New(1e8, 0), b: 27},
		12: {a: New(10, -7), b: 27},
	}
	for i, v := range tests {
		if b := v.a.BitLen(); b != v.b {
			t.Fatalf("#%d: wanted %d, got %d", i, v.b, b)
		}
	}
}

func TestBig_Cmp(t *testing.T) {
	const (
		lesser  = -1
		equal   = 0
		greater = +1
	)

	samePtr := New(0, 0)
	large, ok := new(Big).SetString(strings.Repeat("9", 500))
	if !ok {
		t.Fatal(ok)
	}
	for i, test := range [...]struct {
		a, b *Big
		v    int
	}{
		// Simple
		0: {New(1, 0), New(0, 0), greater},
		1: {New(0, 0), New(1, 0), lesser},
		2: {New(0, 0), New(0, 0), equal},
		// Fractional
		3: {New(9876, 3), New(1234, 0), lesser},
		// zl > 0, xl < 0
		4: {New(1234, 3), New(50, 25), greater},
		// Same pointers
		5: {samePtr, samePtr, equal},
		// Large int vs large big.Int
		6: {New(99999999999, 0), large, lesser},
		7: {large, New(999999999999999999, 0), greater},
		8: {New(4, 0), New(4, 0), equal},
		9: {New(4, 0), new(Big).Quo(New(12, 0), New(3, 0)), equal},
		// z.scale < 0
		10: {large, new(Big).Copy(large), equal},
		// Differing signs
		11: {new(Big).Set(large).Neg(large), large, lesser},
		12: {new(Big).Quo(new(Big).Set(large), New(314156, 5)), large, lesser},
		13: {New(1234, 3), newbig(t, "1000000000000000000000000000000.234"), lesser},
		14: {
			newbig(t, "10000000000000000000"),
			newbig(t, "100000000000000000000").SetScale(1), equal,
		},
		// zl < 0, xl < 0
		15: {New(2, c.BadScale-1), New(2, c.BadScale-2), lesser},
		16: {New(1000000000000000, 16), New(16666666666666666, 18), greater},
	} {
		r := test.a.Cmp(test.b)
		if test.v != r {
			t.Fatalf("#%d: wanted %d, got %d", i, test.v, r)
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
		z.Context.SetPrecision(int32(c.Prec))
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

func TestBig_Quo(t *testing.T) {
	s, close := getTests(t, "division")
	defer close()

	for i := 0; s.Scan(); i++ {
		c, err := suite.ParseCase(s.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		z := new(Big)
		z.Context.SetPrecision(int32(c.Prec))
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
	for i, test := range [...]string{
		"42", "3.14156", "23423141234", ".44444", "1e+1222", "12e-444", "0",
	} {
		rat, ok := new(big.Rat).SetString(test)
		if !ok {
			t.Fatal("!ok")
		}
		xr := new(Big).SetRat(rat).Rat(nil)
		if xr.Cmp(rat) != 0 {
			t.Fatalf("#%d: wanted %q, got %q", i, rat, xr)
		}
	}
}

func TestBig_Round(t *testing.T) {
	for i, test := range [...]struct {
		v   string
		to  int32
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
		if rs := bd.Round(test.to).String(); rs != test.res {
			t.Fatalf(`#%d:
wanted: %q
got   : %q
`, i, test.res, rs)
		}
	}
}

func TestBig_SetFloat64(t *testing.T) {
	tests := map[float64]string{
		123.4:          "123.4",
		123.42:         "123.42",
		123.412345:     "123.412345",
		123.4123456:    "123.4123456",
		123.41234567:   "123.41234567",
		123.412345678:  "123.412345678",
		123.4123456789: "123.4123456789",
	}

	// add negatives
	for p, s := range tests {
		if p > 0 {
			tests[-p] = "-" + s
		}
	}

	var d Big
	for input, s := range tests {
		d.SetFloat64(input)
		if ds := d.String(); ds != s {
			t.Fatalf("wanted %s, got %s", s, ds)
		}
	}

	if !didPanic(func() { d.SetFloat64(math.NaN()) }) {
		t.Fatalf("wanted panic when creating a Big from NaN, got %s instead",
			d.String())
	}

	if testing.Short() {
		return
	}

	var err float64
	for f, s := range testTable {
		d.SetFloat64(f)
		if d.String() != s {
			err++
		}
	}

	// Some margin of error is acceptable when converting from
	// a float. On a table of roughly 9,000 entries an acceptable
	// margin of error is around 450. Using Gaussian/banker's rounding our
	// margin of error is roughly 215 per 9,000 entries, for a rate of around
	// 2.3%.
	if err >= 0.05*float64(len(testTable)) {
		t.Fatalf("wanted error rate to be < 0.05%% of table, got %.f", err)
	}
}

func TestBig_SetString(t *testing.T) {
	tests := []struct {
		dec string
		s   string
	}{
		0: {"0", "0"},
		1: {"00000000000000000000", "0"},
	}
	for i, v := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dec := newbig(t, v.dec)
			s := dec.String()
			if v.s != s {
				t.Fatalf("wanted %s, got %s", v.s, s)
			}
		})
	}
}

func TestBig_Sign(t *testing.T) {
	for i, test := range [...]struct {
		x string
		s int
	}{
		0: {"-Inf", -1},
		1: {"-1", -1},
		2: {"-0", 0},
		3: {"+0", 0},
		4: {"+1", +1},
		5: {"+Inf", +1},
		6: {"100", 1},
		7: {"-100", -1},
	} {
		x, ok := new(Big).SetString(test.x)
		if !ok {
			t.Fatal(ok)
		}
		s := x.Sign()
		if s != test.s {
			t.Fatalf("#%d: %s.Sign() = %d; want %d", i, test.x, s, test.s)
		}
	}
}

func TestBig_SignBit(t *testing.T) {
	x := New(1<<63-1, 0)
	tests := [...]struct {
		a *Big
		b bool
	}{
		0: {a: New(-1, 0), b: true},
		1: {a: New(1, 0), b: false},
		2: {a: x.Mul(x, x), b: false},
		3: {a: new(Big).Neg(x), b: true},
	}
	for i, v := range tests {
		sb := v.a.Signbit()
		if sb != v.b {
			t.Fatalf("#%d: wanted %t, got %t", i, v.b, sb)
		}
	}
}

func TestBig_String(t *testing.T) {
	x := New(1<<63-1, 0)
	tests := [...]struct {
		a *Big
		b string
	}{
		0: {a: New(10, 1), b: "1"},                  // Trim trailing zeros
		1: {a: New(12345, 3), b: "12.345"},          // Normal decimal
		2: {a: New(-9876, 2), b: "-98.76"},          // Negative
		3: {a: New(-1e5, 0), b: strconv.Itoa(-1e5)}, // Large number
		4: {a: New(0, -50), b: "0"},                 // "0"
		5: {a: x.Mul(x, x), b: "85070591730234615847396907784232501249"},
		6: {a: newbig(t, "-1.4e-52"), b: "-1.4e-52"},
		7: {a: newbig(t, "-500.444312301"), b: "-500.444312301"},
	}
	for i, s := range tests {
		str := s.a.String()
		if str != s.b {
			t.Fatalf("#%d: wanted %q, got %q", i, s.b, str)
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
		z.Context.SetPrecision(int32(c.Prec))
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
