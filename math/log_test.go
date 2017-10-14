package math

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ericlagergren/decimal"
)

type truer struct{}

func (t truer) Next() bool { return true }

type testLentzer struct{ recv *decimal.Big }

func (l *testLentzer) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	f = l.recv                                       // f
	Δ = new(decimal.Big)                             // Δ
	C = new(decimal.Big)                             // C
	D = new(decimal.Big)                             // D
	eps = decimal.New(1, l.recv.Context.Precision()) // eps

	const n = 1
	f.Context.SetPrecision(l.recv.Context.Precision() * n)
	C.Context.SetPrecision(l.recv.Context.Precision() * n)
	D.Context.SetPrecision(l.recv.Context.Precision() * n)
	return
}

// alg1 is algorithm 2.4.1 from Cuyt.
type alg1 struct {
	truer
	testLentzer

	z *decimal.Big
	k int64
	a int64
	t Term
}

func logAlg1(z, x *decimal.Big) *decimal.Big {
	x0 := new(decimal.Big).Copy(x)
	x0.Sub(x0, one)
	g := alg1{
		testLentzer: testLentzer{recv: alias(z, x)},
		z:           x0,
		t:           Term{A: new(decimal.Big).Set(x0), B: new(decimal.Big)},
	}
	return z.Quo(x0, Lentz(z, &g))
}

func (a *alg1) Term() Term {
	a.k++
	if a.k&1 == 0 && a.k != 1 {
		a.a++
		a.t.A.SetMantScale(a.a*a.a, 0)
		a.t.A.Mul(a.t.A, a.z)
	}
	a.t.B.SetMantScale(a.k, 0)
	fmt.Println(a.t)
	return a.t
}

// alg2 is algorithm 2.4.4 from Cuyt.
type alg2 struct {
	truer
	testLentzer

	z  *decimal.Big
	zp *decimal.Big // z*z
	z2 *decimal.Big // z*2
	k  int64
	t  Term
}

func logAlg2(z, x *decimal.Big) *decimal.Big {
	x0 := new(decimal.Big).Copy(x)
	x0.Sub(x0, one)
	g := alg2{
		testLentzer: testLentzer{recv: alias(z, x)},
		z:           x0,
		zp:          new(decimal.Big).Mul(x0, x0),
		z2:          new(decimal.Big).Mul(x0, two),
		t:           Term{A: new(decimal.Big), B: new(decimal.Big)},
	}
	return z.Quo(g.z2, Lentz(z, &g))
}

func (a *alg2) Term() Term {
	a.t.A.SetMantScale(a.k*a.k, 0)
	a.t.A.Mul(a.t.A, a.zp)
	a.t.A.Neg(a.t.A)
	a.t.B.SetMantScale(2*a.k, 0)
	a.t.B.Mul(a.t.B, a.z2)
	a.k++
	return a.t
}

// alg3 is algorithm 2.4.7 from Cuyt.
type alg3 struct {
	truer
	testLentzer

	z  *decimal.Big
	z1 *decimal.Big // z+1
	k  int64
	a  int64
	t  Term
}

func logAlg3(z, x *decimal.Big) *decimal.Big {
	x0 := new(decimal.Big).Copy(x)
	x0.Sub(x0, one)
	g := alg3{
		testLentzer: testLentzer{recv: alias(z, x)},
		z:           x0,
		z1:          new(decimal.Big).Mul(x0, one),
		t:           Term{A: new(decimal.Big), B: new(decimal.Big)},
	}
	return z.Quo(x0, Lentz(z, &g))
}

func (a *alg3) Term() Term {
	if a.k == 0 {
		a.t.B.Add(one, a.z)
		a.a = 1
	} else if a.k&1 == 1 {
		a.t.A.Sub(a.t.A, a.z)
		a.t.B.SetMantScale(2, 0)
	} else {
		a.a += 2
		a.t.B.SetMantScale(a.a, 0)
		a.t.B.Mul(a.t.B, a.z1)
	}
	a.k++
	return a.t
}

func BenchmarkLog_alg1_9(b *testing.B)    { benchmarkAlg(b, logAlg1, 9, 100000) }
func BenchmarkLog_alg1_19(b *testing.B)   { benchmarkAlg(b, logAlg1, 19, 100000) }
func BenchmarkLog_alg1_38(b *testing.B)   { benchmarkAlg(b, logAlg1, 38, 100000) }
func BenchmarkLog_alg1_5000(b *testing.B) { benchmarkAlg(b, logAlg1, 5000, 1) }

func BenchmarkLog_alg2_9(b *testing.B)    { benchmarkAlg(b, logAlg2, 9, 100000) }
func BenchmarkLog_alg2_19(b *testing.B)   { benchmarkAlg(b, logAlg2, 19, 100000) }
func BenchmarkLog_alg2_38(b *testing.B)   { benchmarkAlg(b, logAlg2, 38, 100000) }
func BenchmarkLog_alg2_5000(b *testing.B) { benchmarkAlg(b, logAlg2, 5000, 1) }

func BenchmarkLog_alg3_9(b *testing.B)    { benchmarkAlg(b, logAlg3, 9, 100000) }
func BenchmarkLog_alg3_19(b *testing.B)   { benchmarkAlg(b, logAlg3, 19, 100000) }
func BenchmarkLog_alg3_38(b *testing.B)   { benchmarkAlg(b, logAlg3, 38, 100000) }
func BenchmarkLog_alg3_5000(b *testing.B) { benchmarkAlg(b, logAlg3, 5000, 1) }

var gb *decimal.Big

var randstr = func() string {
	b := make([]byte, 5000)
	for i := range b {
		b[i] = byte(rand.Int()%10 + '0')
	}
	b[rand.Int()%(len(b)/5)] = '.'
	return string(b)
}()

func benchmarkAlg(b *testing.B, g func(z, x *decimal.Big) *decimal.Big, prec int32, iters int) {
	lb := new(decimal.Big)
	lb.Context.SetPrecision(prec)

	x, ok := new(decimal.Big).SetString(randstr)
	if !ok {
		panic("!ok")
	}
	x.Round(prec)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; i < iters; j++ {
			lb = g(lb, x)
		}
	}
	gb = lb
}
