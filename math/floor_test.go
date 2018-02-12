package math

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestFloor(t *testing.T) {
	for i, s := range [...]struct {
		x, r string
	}{
		0: {"-2", "-2"},
		1: {"-2.5", "-3"},
		2: {"-2.4", "-3"},
		3: {"5", "5"},
		4: {"0.005", "0"},
		5: {"-0.0005", "-1"},
		6: {"2.9", "2"},
	} {
		x, _ := new(decimal.Big).SetString(s.x)
		r, _ := new(decimal.Big).SetString(s.r)

		z := Floor(x, x)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: floor(%s)
wanted: %s
got   : %s
`, i, s.x, s.r, z)
		}
	}
}

func TestCeil(t *testing.T) {
	for i, s := range [...]struct {
		x, r string
	}{
		0: {"-2", "-2"},
		1: {"-2.5", "-2"},
		2: {"-2.4", "-2"},
		3: {"5", "5"},
		4: {"0.005", "1"},
		5: {"-0.0005", "-0"},
		6: {"2.9", "3"},
	} {
		x, _ := new(decimal.Big).SetString(s.x)
		r, _ := new(decimal.Big).SetString(s.r)

		z := Ceil(x, x)
		if z.Cmp(r) != 0 {
			t.Fatalf(`#%d: floor(%s)
wanted: %s
got   : %s
`, i, s.x, s.r, z)
		}
	}
}
