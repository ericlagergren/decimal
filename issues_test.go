package decimal

import (
	"math/big"
	"testing"
)

func TestIssue20(t *testing.T) {
	x := New(10240000000000, 0)
	x.Mul(x, New(976563, 9))
	if v, _ := x.Int64(); v != 10000005120 {
		t.Fatal("error int64: ", v, x.Int(nil).Int64())
	}
}

func TestIssue65(t *testing.T) {
	const expected = "999999999000000000000000000000"
	r, _ := new(big.Rat).SetString(expected)
	r2 := new(Big).SetRat(r).Rat(nil)
	if r.Cmp(r2) != 0 {
		t.Fatalf("expected %q, got %q", r, r2)
	}
}

func TestIssue71(t *testing.T) {
	x, _ := new(Big).SetString("-433997231707950814777029946371807573425840064343095193931191306942897586882.200850175108941825587256711340679426793690849230895605323379098449524300541372392806145820741928")
	y := New(5, 0)
	ctx := Context{RoundingMode: ToZero, Precision: 364}

	z := new(Big)
	ctx.Quo(z, x, y)

	r, _ := new(Big).SetString("-86799446341590162955405989274361514685168012868619038786238261388579517376.4401700350217883651174513422681358853587381698461791210646758196899048601082744785612291641483856")
	if z.Cmp(r) != 0 || z.Scale() != r.Scale() {
		t.Fatalf(`Quo(%s, %s)
wanted: %s (%d)
got   : %s (%d)
`, x, y, r, -r.Scale(), z, -z.Scale())
	}
}
