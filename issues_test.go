package decimal

import (
	"math/big"
	"testing"
)

func Test_Issue20(t *testing.T) {
	x := New(10240000000000, 0)
	x.Mul(x, New(976563, 9))
	if v, _ := x.Int64(); v != 10000005120 {
		t.Fatal("error int64: ", v, x.Int(nil).Int64())
	}
}

func Test_Issue65(t *testing.T) {
	const expected = "999999999000000000000000000000"
	r, _ := new(big.Rat).SetString(expected)
	r2 := new(Big).SetRat(r).Rat(nil)
	if r.Cmp(r2) != 0 {
		t.Fatalf("expected %q, got %q", r, r2)
	}
}
