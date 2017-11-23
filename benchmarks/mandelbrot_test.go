package benchmarks

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

const mbrotiters = 10000000

var (
	two = decimal.New(2, 0)
)

func BenchmarkMandelbrot_float64(b *testing.B) {
	var lf float64
	for i := 0; i < b.N; i++ {
		x0, y0 := 0.222, 0.333
		x, y := x0, y0
		xx, yy := x*x, y*y
		for j := 0; j < mbrotiters; j++ {
			y = x * y
			y = (y * 2) + y0
			x = xx - yy
			x += x0
			xx = x * x
			yy = y * y
		}
	}
	gf = lf
}

func BenchmarkMandelbrot_decimal_GDA_9(b *testing.B)  { benchMbrotGDA(b, 9) }
func BenchmarkMandelbrot_decimal_GDA_16(b *testing.B) { benchMbrotGDA(b, 16) }
func BenchmarkMandelbrot_decimal_GDA_19(b *testing.B) { benchMbrotGDA(b, 19) }
func BenchmarkMandelbrot_decimal_GDA_34(b *testing.B) { benchMbrotGDA(b, 34) }
func BenchmarkMandelbrot_decimal_GDA_38(b *testing.B) { benchMbrotGDA(b, 38) }

func BenchmarkMandelbrot_decimal_Go_9(b *testing.B)  { benchMbrotGo(b, 9) }
func BenchmarkMandelbrot_decimal_Go_16(b *testing.B) { benchMbrotGo(b, 16) }
func BenchmarkMandelbrot_decimal_Go_19(b *testing.B) { benchMbrotGo(b, 19) }
func BenchmarkMandelbrot_decimal_Go_34(b *testing.B) { benchMbrotGo(b, 34) }
func BenchmarkMandelbrot_decimal_Go_38(b *testing.B) { benchMbrotGo(b, 38) }

func benchMbrotGDA(b *testing.B, prec int) {
	var (
		ls  *decimal.Big
		ctx = decimal.Context{Precision: prec, OperatingMode: decimal.GDA}
		x0  = decimal.WithContext(ctx).SetMantScale(222, 3)
		y0  = decimal.WithContext(ctx).SetMantScale(333, 3)
	)
	for i := 0; i < b.N; i++ {
		ls = mbrotGDA(x0, y0, ctx, prec)
	}
	gs = ls
}

func mbrotGDA(x0, y0 *decimal.Big, ctx decimal.Context, prec int) *decimal.Big {
	var (
		x  = decimal.WithContext(ctx).Set(x0)
		y  = decimal.WithContext(ctx).Set(y0)
		xx = decimal.WithContext(ctx).Mul(x, x)
		yy = decimal.WithContext(ctx).Mul(y, y)
	)
	for i := 0; i < mbrotiters; i++ {
		y.Mul(x, y)
		y.FMA(y, two, y0) // y.Mul(y, two); y.Add(y, y0)
		x.Sub(xx, yy)
		x.Add(x, x0)
		xx.Mul(x, x)
		yy.Mul(y, y)
	}
	return x
}

func benchMbrotGo(b *testing.B, prec int) {
	var (
		ls  *decimal.Big
		ctx = decimal.Context{Precision: prec, OperatingMode: decimal.GDA}
		x0  = decimal.WithContext(ctx).SetMantScale(222, 3)
		y0  = decimal.WithContext(ctx).SetMantScale(333, 3)
	)
	for i := 0; i < b.N; i++ {
		ls = mbrotGo(x0, y0, ctx, prec)
	}
	gs = ls
}

func mbrotGo(x0, y0 *decimal.Big, ctx decimal.Context, prec int) *decimal.Big {
	var (
		x  = decimal.WithContext(ctx).Set(x0)
		y  = decimal.WithContext(ctx).Set(y0)
		xx = decimal.WithContext(ctx).Mul(x, x)
		yy = decimal.WithContext(ctx).Mul(y, y)
	)
	for i := 0; i < mbrotiters; i++ {
		y.Mul(x, y)
		y.FMA(y, two, y0).Round(prec) // y.Mul(y, two); y.Add(y, y0)
		x.Sub(xx, yy)
		x.Add(x, x0).Round(prec)
		xx.Mul(x, x).Round(prec)
		yy.Mul(y, y).Round(prec)
	}
	return x
}
