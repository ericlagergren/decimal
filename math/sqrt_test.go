package math

import (
	"testing"

	"github.com/EricLagergren/decimal"
)

func TestSqrt(t *testing.T) {
	for i, test := range [...]struct{ v, sqrt string }{
		{"25", "5"},
		{"100", "10"},
		{"250", "15.8113883008418966"},
		{"1000", "31.6227766016837933"},
	} {
		a, ok := new(decimal.Big).SetString(test.v)
		if !ok {
			t.Fatal("wanted true, got false")
		}
		z := Sqrt(new(decimal.Big), a)
		if zs := z.String(); zs != test.sqrt {
			t.Fatalf("#%d: Sqrt(%s): got %s, wanted %q", i, test.v, zs, test.sqrt)
		}
	}
}
