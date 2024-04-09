package c

import (
	"math/bits"
	"testing"
)

func TestMaxScale(t *testing.T) {
	check := func(t *testing.T, max int64, maxInf int64) {
		if MaxScale != max {
			t.Fatalf("MaxScale: got %d, expected %d", MaxScale, max)
		}
		if MaxScaleInf != maxInf {
			t.Fatalf("MaxScaleInf: got %d, expected %d", MaxScaleInf, maxInf)
		}
	}
	switch n := bits.UintSize; n {
	case 32:
		if is32bit != 1 {
			t.Fatalf("is32bit: got %d, expected %d", is32bit, 1)
		}
		if is64bit != 0 {
			t.Fatalf("is64bit: got %d, expected %d", is64bit, 0)
		}
		check(t, maxScale32, maxScaleInf32)
	case 64:
		if is32bit != 0 {
			t.Fatalf("is32bit: got %d, expected %d", is32bit, 0)
		}
		if is64bit != 1 {
			t.Fatalf("is64bit: got %d, expected %d", is64bit, 1)
		}
		check(t, maxScale64, maxScaleInf64)
	default:
		t.Fatalf("unknown bit size: %d", n)
	}
}
