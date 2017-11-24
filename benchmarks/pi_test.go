package benchmarks

import (
	"math"
	"testing"

	"github.com/apmckinlay/gsuneido/util/dnum"
	"github.com/cockroachdb/apd"
	"github.com/ericlagergren/decimal"
	ssdec "github.com/shopspring/decimal"
	"gopkg.in/inf.v0"
)

const pi = "3.141592653589793238462643383279502884197169399375105820974944592307816406286208998628034825342117067982148086513282306647093844609550582231725359408128481117450284102701938521105559644622948954930381964428810975665933446128475648233786783165271201909145648566923460348610454326648213394"

func adjustPrecision(prec int32) int32 {
	return int32(math.Ceil(float64(prec) * 1.1))
}

type testFunc func(prec int32) string

func TestPiBenchmarks(t *testing.T) {
	for _, test := range [...]struct {
		name string
		fn   testFunc
	}{
		{"dec-Go", func(prec int32) string {
			return calcPiGo(prec).String()
		}},
		{"dec-GDA", func(prec int32) string {
			return calcPiGDA(prec).String()
		}},
		{"apd", func(prec int32) string {
			return calcPi_apd(uint32(prec)).String()
		}},
		{"shopSpring", func(prec int32) string {
			return calcPi_shopSpring(prec).String()
		}},
		{"inf", func(prec int32) string {
			return calcPi_inf(prec).String()
		}},
	} {
		for _, prec := range [...]int32{9, 19, 38, 100} {
			str := test.fn(prec)
			name := test.name

			x := new(decimal.Big)
			x.Context.Precision = int(prec)
			if _, ok := x.SetString(str); !ok {
				t.Fatalf("%s (%d): bad input: %q", name, prec, str)
			}

			act := new(decimal.Big)
			act.SetString(pi)
			act.Round(int(prec))
			if act.Cmp(x) != 0 {
				t.Fatalf(`%s (%d): bad output:
want: %q
got : %q
`, name, prec, act, x)
			}
		}
	}
}

func newd(c int64, m int32, p int, mode decimal.OperatingMode) *decimal.Big {
	d := decimal.New(c, int(m))
	d.Context.Precision = p
	d.Context.OperatingMode = mode
	return d
}

var (
	eight          = decimal.New(8, 0)
	thirtyTwo      = decimal.New(32, 0)
	apdEight       = apd.New(8, 0)
	apdThirtyTwo   = apd.New(32, 0)
	dnumEight      = dnum.NewDnum(false, 8, 0)
	dnumThirtyTwo  = dnum.NewDnum(false, 32, 0)
	ssdecEight     = ssdec.New(8, 0)
	ssdecThirtyTwo = ssdec.New(32, 0)
	infEight       = inf.NewDec(8, 0)
	infThirtyTwo   = inf.NewDec(32, 0)
)

func calcPi_inf(prec int32) *inf.Dec {
	var (
		lasts = inf.NewDec(0, 0)
		t     = inf.NewDec(3, 0)
		s     = inf.NewDec(3, 0)
		n     = inf.NewDec(1, 0)
		na    = inf.NewDec(0, 0)
		d     = inf.NewDec(0, 0)
		da    = inf.NewDec(24, 0)

		op = prec - 1 // -1 because inf's precision == digits after radix
	)
	prec = adjustPrecision(prec)

	for s.Cmp(lasts) != 0 {
		lasts.Set(s)
		n.Add(n, na)
		na.Add(na, infEight)
		d.Add(d, da)
		da.Add(da, infThirtyTwo)
		t.Mul(t, n)
		t.QuoRound(t, d, inf.Scale(prec), inf.RoundHalfUp)
		s.Add(s, t)
	}
	return s.Round(s, inf.Scale(op), inf.RoundHalfUp)
}

func calcPi_shopSpring(prec int32) ssdec.Decimal {
	var (
		lasts = ssdec.New(0, 0)
		t     = ssdec.New(3, 0)
		s     = ssdec.New(3, 0)
		n     = ssdec.New(1, 0)
		na    = ssdec.New(0, 0)
		d     = ssdec.New(0, 0)
		da    = ssdec.New(24, 0)

		op = prec - 1 // -1 because shopSpring's prec == digits after radix
	)
	prec = adjustPrecision(prec)

	for s.Cmp(lasts) != 0 {
		lasts = s
		n = n.Add(na)
		na = na.Add(ssdecEight)
		d = d.Add(da)
		da = da.Add(ssdecThirtyTwo)
		t = t.Mul(n)
		t = t.DivRound(d, prec)
		s = s.Add(t)
	}
	return s.Round(op)
}

func calcPi_dnum() dnum.Dnum {
	var (
		lasts = dnum.NewDnum(false, 0, 0)
		t     = dnum.NewDnum(false, 3, 0)
		s     = dnum.NewDnum(false, 3, 0)
		n     = dnum.NewDnum(false, 1, 0)
		na    = dnum.NewDnum(false, 0, 0)
		d     = dnum.NewDnum(false, 0, 0)
		da    = dnum.NewDnum(false, 24, 0)
	)
	for dnum.Cmp(s, lasts) != 0 {
		lasts = s
		n = dnum.Add(n, na)
		na = dnum.Add(na, dnumEight)
		d = dnum.Add(d, da)
		da = dnum.Add(da, dnumThirtyTwo)
		t = dnum.Mul(t, n)
		t = dnum.Div(t, d)
		s = dnum.Add(s, t)
	}
	return s
}

func calcPi_float() float64 {
	var (
		lasts = 0.0
		t     = 3.0
		s     = 3.0
		n     = 1.0
		na    = 0.0
		d     = 0.0
		da    = 24.0
	)
	for s != lasts {
		lasts = s
		n += na
		na += 8
		d += da
		da += 32
		t = (t * n) / d
		s = t
	}
	return s
}

func calcPiGo(p int32) *decimal.Big {
	op := int(p)
	prec := int(adjustPrecision(p))
	var (
		lasts = newd(0, 0, prec, decimal.Go)
		t     = newd(3, 0, prec, decimal.Go)
		s     = newd(3, 0, prec, decimal.Go)
		n     = newd(1, 0, prec, decimal.Go)
		na    = newd(0, 0, prec, decimal.Go)
		d     = newd(0, 0, prec, decimal.Go)
		da    = newd(24, 0, prec, decimal.Go)
	)
	for s.Round(prec).Cmp(lasts) != 0 {
		lasts.Set(s)
		n.Add(n, na)
		na.Add(na, eight)
		d.Add(d, da)
		da.Add(da, thirtyTwo)
		t.Mul(t, n)
		t.Quo(t, d)
		s.Add(s, t)
	}
	return s.Round(op)
}

func calcPiGDA(p int32) *decimal.Big {
	op := int(p)
	prec := int(adjustPrecision(p))
	var (
		lasts = newd(0, 0, prec, decimal.GDA)
		t     = newd(3, 0, prec, decimal.GDA)
		s     = newd(3, 0, prec, decimal.GDA)
		n     = newd(1, 0, prec, decimal.GDA)
		na    = newd(0, 0, prec, decimal.GDA)
		d     = newd(0, 0, prec, decimal.GDA)
		da    = newd(24, 0, prec, decimal.GDA)
	)
	for s.Cmp(lasts) != 0 {
		lasts.Set(s)
		n.Add(n, na)
		na.Add(na, eight)
		d.Add(d, da)
		da.Add(da, thirtyTwo)
		t.Mul(t, n)
		t.Quo(t, d)
		s.Add(s, t)
	}
	return s.Round(op)
}

func calcPi_apd(prec uint32) *apd.Decimal {
	var (
		c     = apd.BaseContext.WithPrecision(uint32(adjustPrecision(int32(prec))))
		lasts = apd.New(0, 0)
		t     = apd.New(3, 0)
		s     = apd.New(3, 0)
		n     = apd.New(1, 0)
		na    = apd.New(0, 0)
		d     = apd.New(0, 0)
		da    = apd.New(24, 0)
	)
	for s.Cmp(lasts) != 0 {
		lasts.Set(s)
		c.Add(n, n, na)
		c.Add(na, na, apdEight)
		c.Add(d, d, da)
		c.Add(da, da, apdThirtyTwo)
		c.Mul(t, t, n)
		c.Quo(t, t, d)
		c.Add(s, s, t)
	}
	c.Precision = prec
	c.Round(s, s)
	return s
}

var (
	gf      float64
	gs      *decimal.Big
	apdgs   *apd.Decimal
	dnumgs  dnum.Dnum
	ssdecgs ssdec.Decimal
	infs    *inf.Dec
)

const rounds = 10000

func benchPiGo(b *testing.B, prec int32) {
	var ls *decimal.Big
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPiGo(prec)
		}
	}
	gs = ls
}

func benchPiGDA(b *testing.B, prec int32) {
	var ls *decimal.Big
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPiGDA(prec)
		}
	}
	gs = ls
}

func benchPi_apd(b *testing.B, prec uint32) {
	var ls *apd.Decimal
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPi_apd(prec)
		}
	}
	apdgs = ls
}

func BenchmarkPi_dnum(b *testing.B) {
	var ls dnum.Dnum
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPi_dnum()
		}
	}
	dnumgs = ls
}

func benchPi_shopspring(b *testing.B, prec int32) {
	var ls ssdec.Decimal
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPi_shopSpring(prec)
		}
	}
	ssdecgs = ls
}

func benchPi_inf(b *testing.B, prec int32) {
	var ls *inf.Dec
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPi_inf(prec)
		}
	}
	infs = ls
}

func BenchmarkPi_BaselineFloat64(b *testing.B) {
	var lf float64
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			lf = calcPi_float()
		}
	}
	gf = lf
}

func BenchmarkPi_decimal_Go_9(b *testing.B)   { benchPiGo(b, 9) }
func BenchmarkPi_decimal_Go_19(b *testing.B)  { benchPiGo(b, 19) }
func BenchmarkPi_decimal_Go_38(b *testing.B)  { benchPiGo(b, 38) }
func BenchmarkPi_decimal_Go_100(b *testing.B) { benchPiGo(b, 100) }

func BenchmarkPi_decimal_GDA_9(b *testing.B)   { benchPiGDA(b, 9) }
func BenchmarkPi_decimal_GDA_19(b *testing.B)  { benchPiGDA(b, 19) }
func BenchmarkPi_decimal_GDA_38(b *testing.B)  { benchPiGDA(b, 38) }
func BenchmarkPi_decimal_GDA_100(b *testing.B) { benchPiGDA(b, 100) }

func BenchmarkPi_apd_9(b *testing.B)   { benchPi_apd(b, 9) }
func BenchmarkPi_apd_19(b *testing.B)  { benchPi_apd(b, 19) }
func BenchmarkPi_apd_38(b *testing.B)  { benchPi_apd(b, 38) }
func BenchmarkPi_apd_100(b *testing.B) { benchPi_apd(b, 100) }

func BenchmarkPi_shopspring_9(b *testing.B)   { benchPi_shopspring(b, 9) }
func BenchmarkPi_shopspring_19(b *testing.B)  { benchPi_shopspring(b, 19) }
func BenchmarkPi_shopspring_38(b *testing.B)  { benchPi_shopspring(b, 38) }
func BenchmarkPi_shopspring_100(b *testing.B) { benchPi_shopspring(b, 100) }

func BenchmarkPi_inf_9(b *testing.B)   { benchPi_inf(b, 9) }
func BenchmarkPi_inf_19(b *testing.B)  { benchPi_inf(b, 19) }
func BenchmarkPi_inf_38(b *testing.B)  { benchPi_inf(b, 38) }
func BenchmarkPi_inf_100(b *testing.B) { benchPi_inf(b, 100) }
