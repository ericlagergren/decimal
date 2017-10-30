package arith

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ericlagergren/decimal/internal/arith/pow"
)

func TestLength(t *testing.T) {
	tests := [...]struct {
		i int64
		l int
	}{
		{i: 0, l: 1},
		{i: 1, l: 1},
		{i: 10, l: 2},
		{i: 100, l: 3},
		{i: 1000, l: 4},
		{i: 10000, l: 5},
		{i: 100000, l: 6},
		{i: 1000000, l: 7},
		{i: 10000000, l: 8},
		{i: 100000000, l: 9},
		{i: 1000000000, l: 10},
		{i: 10000000000, l: 11},
		{i: 100000000000, l: 12},
		{i: 1000000000000, l: 13},
		{i: 10000000000000, l: 14},
		{i: 100000000000000, l: 15},
		{i: 1000000000000000, l: 16},
		{i: 10000000000000000, l: 17},
		{i: 100000000000000000, l: 18},
		{i: 1000000000000000000, l: 19},
	}
	for i, v := range tests {
		if l := Length(v.i); l != v.l {
			t.Fatalf("#%d: wanted %d, got %d", i, v.l, l)
		}
	}
}

func TestBigLength(t *testing.T) {
	tests := [...]struct {
		i *big.Int
		l int
	}{
		{i: big.NewInt(0), l: 1},
		{i: big.NewInt(1), l: 1},
		{i: big.NewInt(10), l: 2},
		{i: big.NewInt(100), l: 3},
		{i: big.NewInt(1000), l: 4},
		{i: big.NewInt(10000), l: 5},
		{i: big.NewInt(100000), l: 6},
		{i: big.NewInt(1000000), l: 7},
		{i: big.NewInt(10000000), l: 8},
		{i: big.NewInt(100000000), l: 9},
		{i: big.NewInt(1000000000), l: 10},
		{i: big.NewInt(10000000000), l: 11},
		{i: big.NewInt(100000000000), l: 12},
		{i: big.NewInt(1000000000000), l: 13},
		{i: big.NewInt(10000000000000), l: 14},
		{i: big.NewInt(100000000000000), l: 15},
		{i: big.NewInt(1000000000000000), l: 16},
		{i: big.NewInt(10000000000000000), l: 17},
		{i: big.NewInt(100000000000000000), l: 18},
		{i: big.NewInt(1000000000000000000), l: 19},
	}
	for i, v := range tests {
		if l := BigLength(v.i); l != v.l {
			t.Fatalf("#%d: wanted %d, got %d", i, v.l, l)
		}
	}

	if testing.Short() {
		return
	}

	// Test a really long one.
	x := new(big.Int).Exp(big.NewInt(10), big.NewInt(1e5), nil)
	n := len(x.String())
	if l := BigLength(x); l != n {
		t.Fatalf("exp(10, 1e5): wanted %d, got %d", n, l)
	}

	const overflowCutoff = 14267572532

	// Randomly chosen length so its bit-length is a smidge above overflowCutoff
	// to speed up this looong test.
	nat := make([]big.Word, 222932222)
	nat[0] = 0xDEADBEEF
	for bp := 1; bp < len(nat); bp *= 2 {
		copy(nat[bp:], nat[:bp])
	}
	x.SetBits(nat)

	// Used by math/big.nat to determine the size of the output buffer.
	n = int(float64(x.BitLen())/math.Log2(10)) + 1

	// We're allowed to be +1 larger, but not smaller.
	if l := BigLength(x); l-n > 1 {
		t.Fatalf("really freaking big: wanted %d, got %d", n, l)
	}
}

var lengths = func() []*big.Int {
	var n [4096]*big.Int
	for i := range n {
		n[i] = pow.BigTen(uint64(rand.Int63n(5000)))
	}
	return n[:]
}()

var gl int

func BenchmarkLogarithm(b *testing.B) {
	var ll int
	for i := 0; i < b.N; i++ {
		for _, x := range lengths {
			ll = logLength(x)
		}
	}
	gl = ll
}

func BenchmarkLogarithmNoCmp(b *testing.B) {
	var ll int
	for i := 0; i < b.N; i++ {
		for _, x := range lengths {
			ll = logLengthNoCmp(x)
		}
	}
	gl = ll
}

func BenchmarkReductions(b *testing.B) {
	var ll int
	for i := 0; i < b.N; i++ {
		for _, x := range lengths {
			ll = reductionLength(x)
		}
	}
	gl = ll
}
