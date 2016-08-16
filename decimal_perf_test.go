package decimal

import "testing"

var x Big

func BenchmarkAdd(b *testing.B) {
	y := New(5678, 3)
	for i := 0; i < b.N; i++ {
		x.Add(&x, y)
	}
}

func BenchmarkSub(b *testing.B) {
	y := New(5678, 3)
	for i := 0; i < b.N; i++ {
		x.Sub(&x, y)
	}
}

func BenchmarkQuo(b *testing.B) {
	y := New(1987, 3)
	for i := 0; i < b.N; i++ {
		x.Quo(&x, y)
	}
}

func BenchmarkMul(b *testing.B) {
	y := New(11234, 4)
	for i := 0; i < b.N; i++ {
		x.Mul(&x, y)
	}
}

func BenchmarkSqrt_PerfectSquare(b *testing.B) {
	bench(b.N, []*Big{
		newbig(nil, "1"), newbig(nil, "4"), newbig(nil, "9"), newbig(nil, "16"), newbig(nil, "36"),
		newbig(nil, "49"), newbig(nil, "64"), newbig(nil, "81"), newbig(nil, "100"), newbig(nil, "121"),
	}, 0)
}

func BenchmarkSqrt_FastPath(b *testing.B)    { bench(b.N, cs, 8) }
func BenchmarkSqrt_DefaultPath(b *testing.B) { bench(b.N, cs, 25) }

var cs = []*Big{
	newbig(nil, "75"), newbig(nil, "57"), newbig(nil, "23"), newbig(nil, "250"), newbig(nil, "111"),
	newbig(nil, "3.1"), newbig(nil, "69"), newbig(nil, "4.12e-1"),
}

var b Big

func bench(n int, cs []*Big, prec int32) {
	b.SetPrec(prec)
	for i := 0; i < n; i++ {
		b.Sqrt(cs[i%len(cs)])
	}
}
