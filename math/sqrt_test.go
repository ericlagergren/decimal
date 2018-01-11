package math_test

import (
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
)

func TestDecimal_Hypot(t *testing.T) {
	pi := math.Pi(decimal.WithPrecision(100))
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
		6: {pi.String(), pi.String(), 75, "4.4428829381583662470158809900606936986146216893756902230853956069564347931"},
		7: {"95", "95", 2, "1.3e+2"},
	}
	for i, v := range tests {
		z := decimal.WithPrecision(v.c)
		p, _ := new(decimal.Big).SetString(v.p)
		q, _ := new(decimal.Big).SetString(v.q)
		a, _ := new(decimal.Big).SetString(v.a)
		if math.Hypot(z, p, q).Cmp(a) != 0 {
			t.Fatalf(`#%d:
wanted: %q
got:    %q
`, i, a, z)
		}
	}
}
