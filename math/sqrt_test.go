package math

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestSqrt(t *testing.T) {
	for i, test := range [...]struct {
		inp  string
		out  string
		prec int
	}{
		// special values
		0: {"+inf", "+inf", 0},
		1: {"0", "0", 0},
		// simple perfect squares
		2: {"25", "5", 1},
		3: {"100", "10", 2},
		4: {"123.456", "11.11107556", 10},
		5: {"17", "4.12310562561766", 15},
		6: {"19.7392088021787172376689819", "4.4428829381583662470", 20},
	} {
		z := new(decimal.Big)
		z.Context.Precision = test.prec

		x, ok := new(decimal.Big).SetString(test.inp)
		if !ok {
			t.Fatal("SetString returned false")
		}
		x.Context.Precision = test.prec

		Sqrt(z, x)

		want, ok := new(decimal.Big).SetString(test.out)
		if !ok {
			t.Fatal("SetString returned false")
		}
		if want.Cmp(z) != 0 {
			t.Fatalf(`#%d: Sqrt(%s)
want: %q
got:  %q
`, i, test.inp, want, z)
		}
	}
}

func TestDecimal_Hypot(t *testing.T) {
	tests := [...]struct {
		p, q string
		c    int
		a    string
	}{
		0: {"1", "4", 15, "4.12310562561766"},
		1: {"1", "4", 10, "4.123105626"},
		2: {"1", "2", 2, "2.2"},
		3: {"-12", "599", 5, "599.12"},
		4: {"1.234", "98.76543", 6, "98.7731"},
		5: {"3", "4", 1, "5"},
		6: {_Pi.String(), _Pi.String(), 75, "4.4428829381583662470158809900606936986146216893756902230853956069564347931"},
	}
	var z decimal.Big
	for i, v := range tests {
		z.Context.Precision = v.c
		p, _ := new(decimal.Big).SetString(v.p)
		q, _ := new(decimal.Big).SetString(v.q)
		a, _ := new(decimal.Big).SetString(v.a)
		if Hypot(&z, p, q).Cmp(a) != 0 {
			t.Errorf(`#%d:
wanted: %q
got:    %q
`, i, a, &z)
		}
	}
}
