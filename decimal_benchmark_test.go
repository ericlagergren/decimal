package decimal

import (
	"math/rand"
	"strconv"
	"testing"
)

var d *Decimal

func newDecimal() *Decimal {
	return New(rand.Int63(), rand.Int63())
}

func BenchmarkNew(b *testing.B) {
	r := make([]struct {
		v int64
		e int64
	}, b.N)
	for i := range r {
		r[i].v = rand.Int63()
		r[i].e = rand.Int63()
	}
	b.ReportAllocs()
	b.ResetTimer()

	for _, v := range r {
		d = New(v.v, v.e)
	}
}

// func BenchmarkNewFromFloat(b *testing.B) {
// 	r := make([]float64, b.N)
// 	for i := range r {
// 		r[i] = rand.Float64()
// 	}
// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for _, v := range r {
// 		d = NewFromFloat(v)
// 	}
// }

// func BenchmarkNewFromFloatWithScale(b *testing.B) {
// 	r := make([]struct {
// 		v float64
// 		e int64
// 	}, b.N)
// 	for i := range r {
// 		r[i].v = rand.Float64()
// 		r[i].e = rand.Int63()
// 	}
// 	b.ReportAllocs()
// 	b.ResetTimer()

// 	for _, v := range r {
// 		d = NewFromFloatWithScale(v.v, v.e)
// 	}
// }

func BenchmarkNewFromString(b *testing.B) {
	r := make([]string, b.N)
	for i := range r {
		if i%2 == 0 {
			r[i] = strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
		} else {
			r[i] = strconv.FormatInt(rand.Int63(), 10)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()

	for _, v := range r {
		d, _ = NewFromString(v)
	}
}

func BenchmarkBinomial(b *testing.B) {
	var d Decimal
	for i := b.N - 1; i >= 0; i-- {
		d.Binomial(1000, 990)
	}
}

func BenchmarkAdd(b *testing.B) {
	d1 := newDecimal()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(Decimal).Add(d, d1)
	}
}

func BenchmarkMul(b *testing.B) {
	d1 := newDecimal()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(Decimal).Mul(d, d1)
	}
}

func BenchmarkDiv(b *testing.B) {
	d1 := newDecimal()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(Decimal).Div(d, d1)
	}
}

func BenchmarkSub(b *testing.B) {
	d1 := newDecimal()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(Decimal).Sub(d, d1)
	}
}
