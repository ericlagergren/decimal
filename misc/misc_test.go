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
