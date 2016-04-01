package arith

import (
	"math/big"
	"testing"
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

	// Test a really long one.
	ten := big.NewInt(10)
	x := big.NewInt(1)
	// Tested up to like 7 zeros but it takes *forever* to run.
	for i := 0; i < 100000; i++ {
		x.Mul(x, ten)
	}
	n := len(x.String())
	if l := BigLength(x); l != n {
		t.Fatalf("#%d: wanted %d, got %d", len(tests), n, l)
	}
}
