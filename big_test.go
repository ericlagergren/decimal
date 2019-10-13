package decimal_test

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/test"
)

func TestBig_Abs(t *testing.T)        { test.Abs.Test(t) }
func TestBig_Add(t *testing.T)        { test.Add.Test(t) }
func TestBig_Class(t *testing.T)      { test.Class.Test(t) }
func TestBig_Cmp(t *testing.T)        { test.Cmp.Test(t) }
func TestBig_FMA(t *testing.T)        { test.FMA.Test(t) }
func TestBig_Mul(t *testing.T)        { test.Mul.Test(t) }
func TestBig_Neg(t *testing.T)        { test.Neg.Test(t) }
func TestBig_Quantize(t *testing.T)   { test.Quant.Test(t) }
func TestBig_Quo(t *testing.T)        { test.Quo.Test(t) }
func TestBig_QuoInt(t *testing.T)     { test.QuoInt.Test(t) }
func TestBig_Rat(t *testing.T)        { test.CTR.Test(t) }
func TestBig_Reduce(t *testing.T)     { test.Reduce.Test(t) }
func TestBig_Rem(t *testing.T)        { test.Rem.Test(t) }
func TestBig_RoundToInt(t *testing.T) { test.RoundToInt.Test(t) }
func TestBig_SetString(t *testing.T)  { test.CTS.Test(t) /* Same as CFS */ }
func TestBig_Sign(t *testing.T)       { test.Sign.Test(t) }
func TestBig_SignBit(t *testing.T)    { test.Signbit.Test(t) }
func TestBig_String(t *testing.T)     { test.CTS.Test(t) }
func TestBig_Sub(t *testing.T)        { test.Sub.Test(t) }

var rnd = rand.New(rand.NewSource(0))

func rndn(min, max int) int {
	return rnd.Intn(max-min) + min
}

func randDec() string {
	b := make([]byte, rndn(5, 50))
	for i := range b {
		b[i] = '0' + byte(rndn(0, 10))
	}
	if rnd.Intn(10) != 0 {
		b[rndn(2, len(b))] = '.'
	}
	if b[0] == '0' {
		if b[1] == '0' && b[2] != '.' {
			b = b[1:]
		}
		b[0] = '-'
	}
	return string(b)
}

var randDecs = func() (a [5000]string) {
	for i := range a {
		a[i] = randDec()
	}
	return a
}()

func TestBig_Float(t *testing.T) {
	for i, test := range randDecs {
		flt, ok := new(big.Float).SetString(test)
		if !ok {
			t.Fatal("!ok")
		}
		fv := new(big.Float).SetPrec(flt.Prec())
		xf := new(decimal.Big).SetFloat(flt).Float(fv)
		if xf.String() != flt.String() {
			t.Fatalf("#%d: wanted %s, got %s", i, flt, xf.String())
		}
	}
}

// TestDecimal_Format tests Decimal.Format. The test cases are largely taken
// from the fmt package's test cases.
func TestDecimal_Format(t *testing.T) {
	for i, s := range []struct {
		format string
		input  string
		want   string
	}{
		{"%s", ".12", "0.12"},
		{"%s", "12", "12"},
		{"%.5g", "1", "1"},
		{"%s", "12.34", "12.34"},
		{"%.3g", "12.34", "12.3"},
		{"'%5.2f'", "0.", "' 0.00'"},
		{"%.10f", "0.1234567891", "0.1234567891"},
		{"%.10f", "0.01", "0.0100000000"},
		{"%.10f", "0.0000000000000000000000000000000000000000000000000000000000001", "0.0000000000"},
		{"%+.3e", "0.0", "+0.000e-01"}, // +00 -> -01
		{"%+.3e", "1.0", "+1.000e+00"},
		{"%+.3f", "-1.0", "-1.000"},
		{"%+.3F", "-1.0", "-1.000"},
		{"%+07.2f", "1.0", "+001.00"},
		{"%+07.2f", "-1.0", "-001.00"},
		{"%-07.2f", "1.0", "1.00   "},
		{"%-07.2f", "-1.0", "-1.00  "},
		{"%+-07.2f", "1.0", "+1.00  "},
		{"%+-07.2f", "-1.0", "-1.00  "},
		{"%-+07.2f", "1.0", "+1.00  "},
		{"%-+07.2f", "-1.0", "-1.00  "},
		{"%+10.2f", "+1.0", "     +1.00"},
		{"%+10.2f", "-1.0", "     -1.00"},
		{"% .3E", "-1.0", "-1.000E+00"},
		{"% .3e", "1.0", " 1.000e+00"},
		{"%+.3g", "0.0", "+0.0"}, // += .0
		{"%+.3g", "1.0", "+1"},
		{"%+.3g", "-1.0", "-1"},
		{"% .3g", "-1.0", "-1"},
		{"% .3g", "1.0", " 1"},
		{"%#g", "1e-323", "1.00000e-323"},
		{"%#g", "-1.0", "-1.00000"},
		{"%#g", "1.1", "1.10000"},
		{"%#g", "123456.0", "123456."},
		{"%#g", "1234567.0", "1.234567e+06"},
		{"%#g", "1230000.0", "1.23000e+06"},
		{"%#g", "1000000.0", "1.00000e+06"},
		{"%#.0f", "1.0", "1."},
		{"%#.0e", "1.0", "1.e+00"},
		{"%#.0g", "1.0", "1."},
		{"%#.0g", "1100000.0", "1.e+06"},
		{"%#.4f", "1.0", "1.0000"},
		{"%#.4e", "1.0", "1.0000e+00"},
		{"%#.4g", "1.0", "1.000"},
		{"%#.4g", "100000.0", "1.000e+05"},
		{"%#.0f", "123.0", "123."},
		{"%#.0e", "123.0", "1.e+02"},
		{"%#.0g", "123.0", "1.e+02"},
		{"%#.4f", "123.0", "123.0000"},
		{"%#.4e", "123.0", "1.2300e+02"},
		{"%#.4g", "123.0", "123.0"},
		{"%#.4g", "123000.0", "1.230e+05"},
		{"%#9.4g", "1.0", "    1.000"},
		{"%.68f", "1.0", zeroFill("1.", 68, "")},
		{"%.68f", "-1.0", zeroFill("-1.", 68, "")},
		// float infinites and NaNs
		{"%f", "+Inf", "Infinity"},
		{"%.1f", "-Inf", "-Infinity"},
		{"% f", "NaN", " NaN"},
		{"%20f", "+Inf", "            Infinity"},
		{"% 20F", "+Inf", "            Infinity"},
		{"% 20e", "-Inf", "           -Infinity"},
		{"%+20E", "-Inf", "           -Infinity"},
		{"% +20g", "-Inf", "           -Infinity"},
		{"%+-20G", "+Inf", "+Infinity           "},
		{"%20e", "NaN", "                 NaN"},
		{"% +20E", "NaN", "                +NaN"},
		{"% -20g", "NaN", " NaN                "},
		{"%+-20G", "NaN", "+NaN                "},
		// Zero padding does not apply to infinities and NaN.
		{"%+020e", "+Inf", "           +Infinity"},
		{"%-020f", "-Inf", "-Infinity           "},
		{"%-020E", "NaN", "NaN                 "},
	} {
		//t.Run(strconv.Itoa(i), func(t *testing.T) {
		z, _ := new(decimal.Big).SetString(s.input)
		got := fmt.Sprintf(s.format, z)
		if got != s.want {
			t.Fatalf(`#%d: printf("%s", "%s")
got   : %q
wanted: %q
`, i, s.format, s.input, got, s.want)
		}
		//})
	}
}

// zeroFill generates zero-filled strings of the specified width. The length
// of the suffix (but not the prefix) is compensated for in the width calculation.
func zeroFill(prefix string, width int, suffix string) string {
	return prefix + strings.Repeat("0", width-len(suffix)) + suffix
}

func TestBig_Int(t *testing.T) {
	for i, test := range randDecs {
		a, ok := new(decimal.Big).SetString(test)
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
		b, ok := new(big.Int).SetString(iv, 10)
		if !ok {
			t.Fatal("!ok")
		}
		if n := a.Int(nil); n.Cmp(b) != 0 {
			t.Fatalf("#%d: wanted %q, got %q", i, b, n)
		}
	}
}

func TestBig_Int64(t *testing.T) {
	for i, test := range randDecs {
		a, ok := new(decimal.Big).SetString(test)
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
		n, ok := a.Int64()
		gv, err := strconv.ParseInt(iv, 10, 64)
		if (err == nil) != ok {
			t.Fatalf("#%d: wanted %t, got %t", i, err == nil, ok)
		}
		if ok && (n != gv) {
			t.Fatalf("#%d: wanted %d, got %d", i, gv, n)
		}
	}
}

func TestBig_Uint64(t *testing.T) {
	for i, test := range randDecs {
		a, ok := new(decimal.Big).SetString(test)
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
		n, ok := a.Uint64()
		if _, err := strconv.ParseUint(iv, 10, 64); (err == nil) != ok {
			t.Fatalf("#%d: wanted %t, got %t", i, err == nil, ok)
		}
		if !ok {
			continue
		}
		if ns := strconv.FormatUint(n, 10); ns != iv {
			t.Fatalf("#%d: wanted %q, got %q", i, iv, ns)
		}
	}
}

func TestBig_IsInt(t *testing.T) {
	allZeros := func(s string) bool {
		for _, c := range s {
			if c != '0' {
				return false
			}
		}
		return true
	}
	for i, test := range randDecs {
		x, ok := new(decimal.Big).SetString(test)
		if !ok {
			t.Fatal("TestBig_IsInt !ok")
		}
		j := strings.IndexByte(test, '.')
		if got := x.IsInt(); got != (j < 0 || allZeros(test[j+1:])) {
			t.Fatalf("#%d: (%q).IsInt() == %t", i, test, got)
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
// 		x, ok := new(decimal.Big).SetString(v.a)
// 		if !ok {
// 			t.Fatal("invalid SetString")
// 		}
// 		if fs := fmt.Sprintf(v.format, x); fs != v.b {
// 			t.Fatalf("#%d: wanted %q, got %q:", i, v.b, fs)
// 		}
// 	}
// }

func TestParallel(t *testing.T) {
	x := decimal.New(4, 0)
	y := decimal.New(3, 0)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			m := new(decimal.Big)
			m.Add(x, y)
			m.Mul(m, y)
			m.Quo(m, x)
			m.Sub(m, y)
			m.FMA(m, x, y)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestBig_Prec(t *testing.T) {
	// confirmed to work inside internal/arith/intlen_test.go
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
		bd, _ := new(decimal.Big).SetString(test.v)
		r, _ := new(decimal.Big).SetString(test.res)
		if bd.Round(test.to).Cmp(r) != 0 {
			t.Fatalf(`#%d:
wanted: %q
got   : %q
`, i, test.res, bd)
		}
	}
}

func TestBig_Scan(t *testing.T) {
	// TODO(eric): write this test
}

func TestBig_SetFloat64(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing all 32-bit floats in short mode")
	}

	const eps = 1e-15
	z := decimal.WithPrecision(17)
	for x := uint32(0); x != math.MaxUint32; x++ {
		f := float64(math.Float32frombits(x))
		zf, _ := z.SetFloat64(f).Float64()
		if math.Float64bits(zf) != math.Float64bits(f) {
			if isSpecial(f) || isSpecial(zf) || math.Abs(zf-f) > eps {
				t.Fatalf(`#%d:
wanted: %g
got   : %g
`, x, f, zf)
			}
		}
	}
}

func isSpecial(f float64) bool { return math.IsInf(f, 0) || math.IsNaN(f) }
