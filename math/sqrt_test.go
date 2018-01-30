package math

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestDecimal_Hypot(t *testing.T) {
	pi := Pi(decimal.WithPrecision(100))
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
		if Hypot(z, p, q).Cmp(a) != 0 {
			t.Fatalf(`#%d:
wanted: %q
got:    %q
`, i, a, z)
		}
	}
}

func TestIssue69(t *testing.T) {
	x := new(decimal.Big)
	maxSqrt := uint64(4294967295)
	if testing.Short() {
		maxSqrt = 1e7
	}
	for i := maxSqrt; i != 0; i-- {
		x.SetUint64(i * i)
		v, _ := decimal.Raw(Sqrt(x, x))
		if *v != i {
			t.Fatalf(`Sqrt(%d)
wanted: %d (0)
got   : %d (%d)
`, i*i, i, v, -x.Scale())
		}
	}
}

func TestIssue73(t *testing.T) {
	x := decimal.New(16, 2)
	z := decimal.WithPrecision(4100)
	Sqrt(z, x)
	r := decimal.New(4, 1)
	if z.Cmp(r) != 0 || z.Scale() != r.Scale() || z.Context.Conditions != r.Context.Conditions {
		t.Fatalf(`Sqrt(%d)
wanted: %s (%d) %s
got   : %s (%d) %s
`, x, r, -r.Scale(), r.Context.Conditions, z, -z.Scale(), z.Context.Conditions)
	}
}

func TestIssue75(t *testing.T) {
	x := decimal.New(576, 2)
	z := decimal.WithPrecision(2)
	Sqrt(z, x)
	r := decimal.New(24, 1)
	if z.Cmp(r) != 0 || z.Scale() != r.Scale() || z.Context.Conditions != r.Context.Conditions {
		t.Fatalf(`Sqrt(%d)
wanted: %s (%d) %s
got   : %s (%d) %s
`, x, r, -r.Scale(), r.Context.Conditions, z, -z.Scale(), z.Context.Conditions)
	}
}
