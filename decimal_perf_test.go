package decimal

import (
	"math/big"
	"sync"
	"testing"
)

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

var intPool = sync.Pool{
	New: func() interface{} { return new(big.Int) },
}

var zeroPool = sync.Pool{
	New: func() interface{} { return new(big.Int) },
}

var getInt = func(x int64) *big.Int {
	return intPool.Get().(*big.Int).SetInt64(x)
}

var putInt = func(x *big.Int) {
	intPool.Put(x)
}
