package math

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestSqrt(t *testing.T) {
	for i, test := range [...]struct {
		inp  string
		out  string
		prec int32
	}{
		// special values
		{"+inf", "+inf", 0}, {"0", "0", 0},
		// simple perfect squares
		{"25", "5", 1}, {"100", "10", 2},
		{"31415", "177.2427713", 10},
	} {
		z := new(decimal.Big)
		z.Context.SetPrecision(test.prec)

		x, _ := new(decimal.Big).SetString(test.inp)
		x.Context.SetPrecision(test.prec)

		Sqrt(z, x)

		want, _ := new(decimal.Big).SetString(test.out)
		if want.Cmp(z) != 0 {
			t.Fatalf(`#%d:
want: %q
got:  %q
`, i, want, z)
		}
	}
}

func TestDecimal_Hypot(t *testing.T) {
	tests := [...]struct {
		p, q *decimal.Big
		c    int32
		a    string
	}{
		{decimal.New(1, 0), decimal.New(4, 0), 15, "4.12310562561766"},
		{decimal.New(1, 0), decimal.New(4, 0), 10, "4.1231056256"},
		{Pi, Pi, 75, "4.442882938158366247015880990060693698614621689375690223085395606956434793099"},
		{decimal.New(-12, 0), decimal.New(599, 0), 2, "599.12"},
		{decimal.New(1234, 3), decimal.New(987654123, 5), 2, "9876.54"},
		{decimal.New(3, 0), decimal.New(4, 0), 0, "5"},
	}
	var a decimal.Big
	for i, v := range tests {
		a.Context.SetPrecision(v.c)
		if got := Hypot(&a, v.p, v.q).String(); got != v.a {
			t.Errorf("#%d: wanted %q, got %q", i, v.a, got)
		}
	}
}
