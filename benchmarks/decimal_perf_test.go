package decimal

import (
	"testing"

	"github.com/cockroachdb/apd"
	"github.com/ericlagergren/decimal"
)

func newd(c int64, m int32, p int32) *decimal.Big {
	d := decimal.New(c, m)
	d.Context.SetPrecision(p)
	return d
}

var (
	eight        = decimal.New(8, 0)
	thirtyTwo    = decimal.New(32, 0)
	apdEight     = apd.New(8, 0)
	apdThirtyTwo = apd.New(32, 0)
)

func calcPifloat() float64 {
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

func calcPi(prec int32) *decimal.Big {
	var (
		lasts = newd(0, 0, prec)
		t     = newd(3, 0, prec)
		s     = newd(3, 0, prec)
		n     = newd(1, 0, prec)
		na    = newd(0, 0, prec)
		d     = newd(0, 0, prec)
		da    = newd(24, 0, prec)
	)
	for s.Cmp(lasts) != 0 {
		lasts.Set(s)
		n.Add(n, na)
		na.Add(na, eight)
		d.Add(d, da)
		da.Add(da, thirtyTwo)
		t.Mul(t, n)
		t.Quo(t, d)
		s.Add(s, t).Round(prec)
	}
	return s
}

func calcPiapd(prec uint32) *apd.Decimal {
	var (
		c     = apd.BaseContext.WithPrecision(prec)
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
	return s
}

var (
	gf    float64
	gs    *decimal.Big
	apdgs *apd.Decimal
)

const rounds = 10000

func benchPi(b *testing.B, prec int32) {
	var ls *decimal.Big
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPi(prec)
		}
	}
	gs = ls
}

func benchPiapd(b *testing.B, prec uint32) {
	var ls *apd.Decimal
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			ls = calcPiapd(prec)
		}
	}
	apdgs = ls
}

func BenchmarkPi_Baseline(b *testing.B) {
	var lf float64
	for i := 0; i < b.N; i++ {
		for j := 0; j < rounds; j++ {
			lf = calcPifloat()
		}
	}
	gf = lf
}

func BenchmarkPi_9(b *testing.B)   { benchPi(b, 9) }
func BenchmarkPi_19(b *testing.B)  { benchPi(b, 19) }
func BenchmarkPi_38(b *testing.B)  { benchPi(b, 38) }
func BenchmarkPi_100(b *testing.B) { benchPi(b, 100) }

func BenchmarkPi_apd_9(b *testing.B)   { benchPiapd(b, 9) }
func BenchmarkPi_apd_19(b *testing.B)  { benchPiapd(b, 19) }
func BenchmarkPi_apd_38(b *testing.B)  { benchPiapd(b, 38) }
func BenchmarkPi_apd_100(b *testing.B) { benchPiapd(b, 100) }

/*

func BenchmarkMandelbrot10000000(b *testing.B) { domb(b, 10000000, 19) }

func domb(b *testing.B, iter int, prec int32) {
	var (
		ls *decimal.Big
		x0 = newd(222, 3, prec)
		y0 = newd(333, 3, prec)
	)
	for i := 0; i < b.N; i++ {
		ls = mandelbrot(x0, y0, iter, prec)
	}
	gs = ls
}

func mandelbrot(x0, y0 *decimal.Big, i int, prec int32) *decimal.Big {
	var (
		two = newd(2, 0, prec)
		x   = new(decimal.Big).Set(x0)
		y   = new(decimal.Big).Set(y0)
		xx  = new(decimal.Big).Mul(x, x)
		yy  = new(decimal.Big).Mul(y, y)
	)
	for ; i >= 0; i-- {
		y.Mul(x, y)
		y.Mul(y, two)
		y.Add(y, y0).Round(prec)
		x.Sub(xx, yy)
		x.Add(x, x0).Round(prec)
		xx.Mul(x, x).Round(prec)
		yy.Mul(y, y).Round(prec)
	}
	return x
}
*/
