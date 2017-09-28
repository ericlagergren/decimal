package decimal

import (
	"testing"
)

// Benchmarks from http://www.bytereef.org/mpdecimal/benchmarks.html

func newd(c int64, m int32, p int32) *Big {
	d := New(c, m)
	d.Context.SetPrecision(p)
	return d
}

var (
	eight     = New(8, 0)
	thirtyTwo = New(32, 0)
)

func calcPi(prec int32) *Big {
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

func Test_calcPi(t *testing.T) {
	p := New(3141592653, 9)
	c := calcPi(10)
	if p.Cmp(c) != 0 {
		t.Fatalf("wanted %q, got %q", p, c)
	}
}

func BenchmarkPifloat64Baseline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10000; j++ {
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
		}
	}
}

var gs *Big

func pi(b *testing.B, prec int32) {
	var ls *Big
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10000; j++ {
			ls = calcPi(prec)
		}
	}
	gs = ls
}

func BenchmarkPi9(b *testing.B)   { pi(b, 9) }
func BenchmarkPi19(b *testing.B)  { pi(b, 19) }
func BenchmarkPi38(b *testing.B)  { pi(b, 38) }
func BenchmarkPi100(b *testing.B) { pi(b, 100) }

func BenchmarkMandelbrot10000000(b *testing.B) { domb(b, 10000000, 19) }

func domb(b *testing.B, iter int, prec int32) {
	var (
		ls *Big
		x0 = newd(222, 3, prec)
		y0 = newd(333, 3, prec)
	)
	for i := 0; i < b.N; i++ {
		ls = mandelbrot(x0, y0, iter, prec)
	}
	gs = ls
}

func mandelbrot(x0, y0 *Big, i int, prec int32) *Big {
	var (
		two = newd(2, 0, prec)
		x   = new(Big).Set(x0)
		y   = new(Big).Set(y0)
		xx  = new(Big).Mul(x, x)
		yy  = new(Big).Mul(y, y)
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
