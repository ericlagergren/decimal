package misc

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestCmpTotal(t *testing.T) {
	for i, test := range [...]struct {
		x, y string
		r    int
	}{
		0: {"qnan", "snan", +1},
		1: {"400", "qnan", -1},
		2: {"snan", "snan", 0},
		3: {"snan", "1e+9", +1},
		4: {"qnan", "1e+12", +1},
		5: {"-qnan", "12", -1},
		6: {"12", "qnan", -1},
	} {
		x := new(decimal.Big)
		x.Context.OperatingMode = decimal.GDA
		x.SetString(test.x)

		y := new(decimal.Big)
		y.Context.OperatingMode = decimal.GDA
		y.SetString(test.y)

		if r := CmpTotal(x, y); r != test.r {
			t.Fatalf("#%d: CmpTotal(%q, %q): got %d, wanted %d",
				i, x, y, r, test.r)
		}
	}
}

func TestShift(t *testing.T) {
	for i, test := range [...]struct {
		prec  int
		x     string
		shift int
		r     string
	}{
		0: {9, "34", 8, "400000000"},
		1: {9, "12", 9, "0"},
		2: {9, "123456789", -2, "1234567"},
		3: {9, "123456789", 0, "123456789"},
		4: {9, "123456789", +2, "345678900"},
	} {
		z := new(decimal.Big)
		z.Context.Precision = test.prec
		z.Context.OperatingMode = decimal.GDA

		x, _ := new(decimal.Big).SetString(test.x)
		r, _ := new(decimal.Big).SetString(test.r)

		Shift(z, x, test.shift)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: Shift(%q, %d)
wanted %q:
got    %q:
`, i, x, test.shift, r, z)
		}
	}
}
