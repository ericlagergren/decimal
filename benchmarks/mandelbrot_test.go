package benchmarks

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
