package math

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestSin(t *testing.T) {
	(tests{
		{"42", "-0.916521547915634", 16},
	}).run(t, Sin)
}

func TestTan(t *testing.T) {
	(tests{
		{"42", "2.291387992437485", 16},
		{"42", "2.29138799243748609", 18},
	}).run(t, Tan)
}

func TestCos(t *testing.T) {
	(tests{
		{"4", "-0.653643620863611", 16},
		{"42", "-0.39998531498835129", 18},
	}).run(t, Cos)
}

func TestCot(t *testing.T) {
	(tests{
		{"4", "0.863691154450616", 16},
		{"42", "0.4364167060752728", 18},
	}).run(t, Cot)
}

type tests []struct {
	x, want string
	prec    int32
}

func (ts tests) run(t *testing.T, fn func(z, x *decimal.Big) *decimal.Big) {
	for i, test := range ts {
		d, ok := new(decimal.Big).SetString(test.x)
		if !ok {
			t.Fatal(ok)
		}
		d.SetPrecision(test.prec)
		fn(d, d)
		ds := d.String()
		if ds != test.want {
			t.Fatalf(`#%d:
want: %q
got : %q`,
				i, test.want, ds)
		}
	}
}
