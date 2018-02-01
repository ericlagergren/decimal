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

func adjustPrecision(prec int) int { return int(math.Ceil(float64(prec) * 1.1)) }

type testFunc func(prec int) string

func TestPiBenchmarks(t *testing.T) {
	for _, test := range [...]struct {
		name string
		fn   testFunc
	}{
		{"dec-Go", func(prec int) string {
			return calcPiGo(prec).String()
		}},
		{"dec-GDA", func(prec int) string {
			return calcPiGDA(prec).String()
		}},
		{"apd", func(prec int) string {
			return calcPi_apd(uint32(prec)).String()
		}},
		{"shopSpring", func(prec int) string {
			return calcPi_shopSpring(int32(prec)).String()
		}},
		{"inf", func(prec int) string {
			return calcPi_inf(prec).String()
		}},
	} {
		var ctx decimal.Context
		for _, prec := range [...]int{9, 19, 38, 100} {
			ctx.Precision = prec

			str := test.fn(prec)
			name := test.name

			var x decimal.Big
			if _, ok := ctx.SetString(&x, str); !ok {
				t.Fatalf("%s (%d): bad input: %q", name, prec, str)
			}

			var act decimal.Big
			ctx.SetString(&act, pi)
			if act.Cmp(&x) != 0 {
				t.Fatalf(`%s (%d): bad output:
want: %q
got : %q
`, name, prec, &act, &x)
			}
		}
	}
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

func calcPi_inf(prec int) *inf.Dec {
	var (
		lasts = inf.NewDec(0, 0)
		t     = inf.NewDec(3, 0)
		s     = inf.NewDec(3, 0)
		n     = inf.NewDec(1, 0)
		na    = inf.NewDec(0, 0)
		d     = inf.NewDec(0, 0)
		da    = inf.NewDec(24, 0)

		work = adjustPrecision(prec)
	)

	for s.Cmp(lasts) != 0 {
		lasts.Set(s)
		n.Add(n, na)
		na.Add(na, infEight)
		d.Add(d, da)
		da.Add(da, infThirtyTwo)
		t.Mul(t, n)
		t.QuoRound(t, d, inf.Scale(work), inf.RoundHalfUp)
		s.Add(s, t)
	}
	// -1 because inf's precision == digits after radix
	return s.Round(s, inf.Scale(prec-1), inf.RoundHalfUp)
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

		work = int32(adjustPrecision(int(prec)))
	)

	for s.Cmp(lasts) != 0 {
		lasts = s
		n = n.Add(na)
		na = na.Add(ssdecEight)
		d = d.Add(da)
		da = da.Add(ssdecThirtyTwo)
		t = t.Mul(n)
		t = t.DivRound(d, work)
		s = s.Add(t)
	}
	// -1 because shopSpring's prec == digits after radix
	return s.Round(prec - 1)
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

func calcPiGo(prec int) *decimal.Big {
	var (
		ctx = decimal.Context{
			Precision:     adjustPrecision(prec),
			OperatingMode: decimal.Go,
		}

		lasts = new(decimal.Big)
		t     = decimal.New(3, 0)
		s     = decimal.New(3, 0)
		n     = decimal.New(1, 0)
		na    = new(decimal.Big)
		d     = new(decimal.Big)
		da    = decimal.New(24, 0)
		eps   = decimal.New(1, prec)
	)

	for {
		ctx.Set(lasts, s)
		ctx.Add(n, n, na)
		ctx.Add(na, na, eight)
		ctx.Add(d, d, da)
		ctx.Add(da, da, thirtyTwo)
		ctx.Mul(t, t, n)
		ctx.Quo(t, t, d)
		ctx.Add(s, s, t)
		if ctx.Sub(lasts, s, lasts).CmpAbs(eps) < 0 {
			return s.Round(prec)
		}
	}
}

func calcPiGDA(prec int) *decimal.Big {
	var (
		ctx = decimal.Context{
			Precision:     adjustPrecision(prec),
			OperatingMode: decimal.GDA,
		}

		lasts = new(decimal.Big)
		t     = decimal.New(3, 0)
		s     = decimal.New(3, 0)
		n     = decimal.New(1, 0)
		na    = new(decimal.Big)
		d     = new(decimal.Big)
		da    = decimal.New(24, 0)
	)

	for s.Cmp(lasts) != 0 {
		ctx.Set(lasts, s)
		ctx.Add(n, n, na)
		ctx.Add(na, na, eight)
		ctx.Add(d, d, da)
		ctx.Add(da, da, thirtyTwo)
		ctx.Mul(t, t, n)
		ctx.Quo(t, t, d)
		ctx.Add(s, s, t)
	}
	return s.Round(prec)
}

func calcPi_apd(prec uint32) *apd.Decimal {
	var (
		ctx   = apd.BaseContext.WithPrecision(uint32(adjustPrecision(int(prec))))
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
		ctx.Add(n, n, na)
		ctx.Add(na, na, apdEight)
		ctx.Add(d, d, da)
		ctx.Add(da, da, apdThirtyTwo)
		ctx.Mul(t, t, n)
		ctx.Quo(t, t, d)
		ctx.Add(s, s, t)
	}
	ctx.Precision = prec
	ctx.Round(s, s)
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

func benchPiGo(b *testing.B, prec int) {
	var ls *decimal.Big
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPiGo(prec)
		}
	}
	gs = ls
}

func benchPiGDA(b *testing.B, prec int) {
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

func benchPi_inf(b *testing.B, prec int) {
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
