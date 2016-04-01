package checked

import "testing"

const maxInt = 1<<63 - 1

func TestAdd(t *testing.T) {
	tests := [...]struct {
		a, b int64
		r    int64
		ok   bool
	}{
		{1, 2, 3, true},
		{3, 3, 6, true},
		{maxInt, 2, 0, false},
		{maxInt - 2, 2, maxInt, true},
		{-maxInt, maxInt, 0, true},
		{-maxInt, -maxInt, 0, false},
	}
	for i, v := range tests {
		r, ok := Add(v.a, v.b)

		// Make sure overflow check works.
		if ok != v.ok {
			t.Fatalf("#%d: wanted %t, got %t", i, v.ok, ok)
		}

		// Make sure result is valid iff no overflow.
		if v.ok && r != v.r {
			t.Fatalf("#%d: wanted %d, got %d", i, v.r, r)
		}
	}
}

var globalxx int64
var globalok bool

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globalxx, globalok = Add(int64(i), int64(i+1))
	}
}

func BenchmarkMul(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globalxx, globalok = Mul(int64(i), int64(i+1))
	}
}

func BenchmarkSub(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globalxx, globalok = Sub(int64(i), int64(i+1))
	}
}

func BenchmarkAddSelf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globalxx, globalok = Add(int64(i), int64(i))
	}
}

func BenchmarkMulSelf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globalxx, globalok = Mul(int64(i), 2)
	}
}
