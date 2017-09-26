package decimal

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/ericlagergren/decimal/internal/c"
)

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
	type inp struct {
		a   string
		b   string
		res string
		nan bool
	}

	inputs := [...]inp{
		0:  {a: "2", b: "3", res: "5"},
		1:  {a: "2454495034", b: "3451204593", res: "5905699627"},
		2:  {a: "24544.95034", b: ".3451204593", res: "24545.2954604593"},
		3:  {a: ".1", b: ".1", res: "0.2"},
		4:  {a: ".1", b: "-.1", res: "0"},
		5:  {a: "0", b: "1.001", res: "1.001"},
		6:  {a: "123456789123456789.12345", b: "123456789123456789.12345", res: "246913578246913578.2469"},
		7:  {a: ".999999999", b: ".00000000000000000000000000000001", res: "0.99999999900000000000000000000001"},
		8:  {"+Inf", "-Inf", "NaN", true},
		9:  {"+Inf", "+Inf", "+Inf", false},
		10: {"0", "0", "0", false},
	}

	for i, inp := range inputs {
		a, ok := new(Big).SetString(inp.a)
		if !ok {
			t.FailNow()
		}
		b, ok := new(Big).SetString(inp.b)
		if !ok {
			t.FailNow()
		}
		panicked := didPanic(func() { a.Add(a, b) })
		if panicked != inp.nan {
			t.Fatalf("#%d: wanted panicked == %t, got %t", i, inp.nan, panicked)
		}
		if as := a.String(); as != inp.res {
			t.Fatalf("#%d: wanted %s, got %s", i, inp.res, as)
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
		n := a.Int()
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
		"0.5",
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
			t.Fatalf("#%d: %s.IsInt() == %t", i, s, got)
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
	for i, v := range mulTestTable {
		d1 := newbig(t, v.d1)
		d2 := newbig(t, v.d2)
		panicked := didPanic(func() { d1.Mul(d1, d2) })
		if panicked != v.nan {
			t.Fatalf("#%d: wanted panicked: %t, got %t", i, v.nan, panicked)
		}
		if s := d1.String(); s != v.res {
			t.Fatalf("#%d: wanted %s got %s", i, v.res, s)
		}
	}
}

func TestBig_Prec(t *testing.T) {
	// confirmed to work inside internal/arith/intlen_test.go
}

func TestBig_Quo(t *testing.T) {
	huge1, ok := new(Big).SetString("12345678901234567890.1234")
	if !ok {
		t.Fatal("invalid")
	}
	huge2, ok := new(Big).SetString("239482394823948239843298432984.4324324234324234324")
	if !ok {
		t.Fatal("invalid")
	}

	huge3, ok := new(Big).SetString("10000000000000000000000000000000000000000")
	if !ok {
		t.Fatal("invalid")
	}
	huge4, ok := new(Big).SetString("10000000000000000000000000000000000000000")
	if !ok {
		t.Fatal("invalid")
	}

	tests := [...]struct {
		a *Big
		b *Big
		p int32
		r string
	}{
		0:  {a: New(10, 0), b: New(2, 0), r: "5"},
		1:  {a: New(1234, 3), b: New(-2, 0), r: "-0.617"},
		2:  {a: New(10, 0), b: New(3, 0), r: "3.333333333333333"},
		3:  {a: New(100, 0), b: New(3, 0), p: 4, r: "33.33"},
		4:  {a: New(-405, 1), b: New(1257, 2), r: "-3.221957040572792"},
		5:  {a: New(-991242141244124, 7), b: New(235325235323, 3), r: "-0.4212222033406559"},
		6:  {a: huge1, b: huge2, r: "5.155150928864855e-11"},
		7:  {a: New(1000, 0), b: New(20, 0), r: "50"},
		8:  {a: huge3, b: huge4, r: "1"},
		9:  {a: New(100, 0), b: New(1, 0), r: "100"},
		10: {a: New(10, 0), b: New(1, 0), r: "10"},
		11: {a: New(1, 0), b: New(10, 0), r: "0.1"},
	}
	for i, v := range tests {
		if v.p != 0 {
			v.a.Context.SetPrecision(v.p)
		} else {
			v.a.Context.SetPrecision(DefaultPrecision)
		}
		q := v.a.Quo(v.a, v.b)
		if qs := q.String(); qs != v.r {
			t.Fatalf("#%d: wanted %q, got %q", i, v.r, qs)
		}
	}
}

func TestBig_Round(t *testing.T) {
	for i, test := range [...]struct {
		v   string
		to  int32
		res string
	}{
		{"5.5", 1, "6"},
		{"1.234", 2, "1.2"},
		{"1", 1, "1"},
		{"9.876", 0, "9.876"},
		{"5.65", 2, "5.6"},
		{"5.0002", 2, "5"},
		{"0.000158674", 6, "0.000158674"},
	} {
		bd := newbig(t, test.v)
		if rs := bd.Round(test.to).String(); rs != test.res {
			t.Fatalf("#%d: wanted %s, got %s", i, test.res, rs)
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
	inputs := [...]struct {
		a   string
		b   string
		r   string
		nan bool
	}{
		0:  {a: "2", b: "3", r: "-1"},
		1:  {a: "12", b: "3", r: "9"},
		2:  {a: "-2", b: "9", r: "-11"},
		3:  {a: "2454495034", b: "3451204593", r: "-996709559"},
		4:  {a: "24544.95034", b: ".3451204593", r: "24544.6052195407"},
		5:  {a: ".1", b: "-.1", r: "0.2"},
		6:  {a: ".1", b: ".1", r: "0"},
		7:  {a: "0", b: "1.001", r: "-1.001"},
		8:  {a: "1.001", b: "0", r: "1.001"},
		9:  {a: "2.3", b: ".3", r: "2"},
		10: {"+Inf", "-Inf", "+Inf", false},
		11: {"+Inf", "+Inf", "NaN", true},
		12: {"0", "0", "0", false},
	}

	for i, inp := range inputs {
		a, ok := new(Big).SetString(inp.a)
		if !ok {
			t.FailNow()
		}
		b, ok := new(Big).SetString(inp.b)
		if !ok {
			t.FailNow()
		}
		panicked := didPanic(func() { a.Sub(a, b) })
		if panicked != inp.nan {
			t.Fatalf("#%d: wanted panicked == %t, got %t", i, inp.nan, panicked)
		}
		if as := a.String(); as != inp.r {
			t.Fatalf("#%d: wanted %s, got %s", i, inp.r, as)
		}
	}
}
