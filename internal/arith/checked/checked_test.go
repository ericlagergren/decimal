package checked

import (
	"testing"

	"github.com/EricLagergren/decimal/internal/c"
)

const maxInt64 = 1<<63 - 1

func TestAdd(t *testing.T) {
	tests := [...]struct {
		a, b int64
		r    int64
		ok   bool
	}{
		{1, 2, 3, true},
		{3, 3, 6, true},
		{maxInt64, 2, 0, false},
		{maxInt64 - 2, 2, maxInt64, true},
		{-maxInt64, maxInt64, 0, true},
		{-maxInt64, -maxInt64, 0, false},
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

const maxInt32 = 1<<31 - 1

func TestAdd32(t *testing.T) {
	tests := [...]struct {
		a, b int32
		r    int32
		ok   bool
	}{
		{1, 2, 3, true},
		{3, 3, 6, true},
		{maxInt32, 2, 0, false},
		{maxInt32 - 2, 2, maxInt32, true},
		{-maxInt32, maxInt32, 0, true},
		{-maxInt32, -maxInt32, 0, false},
	}
	for i, v := range tests {
		r, ok := Add32(v.a, v.b)

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

// Configure these parameters to test the varying checked
// functions below.
func init() {
	x = 500
	y = maxInt64
}

var x, y int64

func Benchmark1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var n int64
		if !_checkedAdd1(x, y, &n) {
			globalxx = n
		}
	}
}

func Benchmark2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := _checkedAdd2(x, y)
		if n == c.Inflated {
			globalxx = n
		}
	}
}

func Benchmark3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n, ok := Add(x, y)
		if !ok {
			globalxx = n
		}
	}
}

// This could screw with escape analysis, but only returns
// a bool and requires one branch.
func _checkedAdd1(x, y int64, z *int64) bool {
	sum := x + y
	return (sum^x)&(sum^y) >= 0
}

// This requires two branches, but doesn't mess with escape analysis. (Second branch is when you check for sum == c.Inflated.)
func _checkedAdd2(x, y int64) (sum int64) {
	sum = x + y
	if (sum^x)&(sum^y) < 0 {
		return c.Inflated
	}
	return sum
}

// This returns two values, but only requires one branch and does
// not mess with escape analysis.
//
// Currently this is the fastest. It is
func _checkedAdd3(x, y int64) (sum int64, ok bool) {
	return Add(x, y)
}
