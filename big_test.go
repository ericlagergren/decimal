package decimal_test

import (
	"math"
	"math/big"
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

func TestBig_Float(t *testing.T) {
	for i, test := range [...]string{
		"42", "3.14156", "23423141234", ".44444", "1e+1222", "12e-444", "0",
	} {
		flt, ok := new(big.Float).SetString(test)
		if !ok {
			t.Fatal("!ok")
		}
		fv := new(big.Float).SetPrec(flt.Prec())
		xf := new(decimal.Big).SetFloat(flt).Float(fv)
		if xf.String() != flt.String() {
			t.Fatalf("#%d: wanted %f, got %f", i, flt, xf)
		}
	}
}

func TestBig_Int(t *testing.T) {
	for i, test := range [...]string{
		"1.234", "4.567", "11111111111111111111111111111111111.2",
		"1234234.2321", "121111111111", "44444444.241", "1241.1",
		"4", "5123", "1.2345123134123414123123213", "0.11", ".1",
	} {
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
		if !ok {
			t.Fatal("!ok")
		}
		if ns := strconv.FormatInt(n, 10); ns != iv {
			t.Fatalf("#%d: wanted %q, got %q", i, iv, ns)
		}
	}
}

func TestBig_IsInt(t *testing.T) {
	for i, test := range [...]string{
		"1.087581170583171279366331325163992810993060588169144153517806339238748036659594606503711549623097075801903290898984816913699837852618679612062658508694865627080580343806827457751585727929883451128788810220782555198023845932678964045544369555311671308165766927777574386318610481491980102511680466744045522904137471213980283536704254600843996379022514957521",
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
		x, ok := new(decimal.Big).SetString(s)
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
		return
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
