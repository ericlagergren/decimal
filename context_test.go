package decimal

import "testing"

func TestCondition_String(t *testing.T) {
	for i, test := range [...]struct {
		c Condition
		s string
	}{
		{Clamped, "clamped"},
		{Clamped | Underflow, "clamped, underflow"},
		{Inexact | Rounded | Subnormal, "inexact, rounded, subnormal"},
		{1 << 31, "unknown(2147483648)"},
	} {
		s := test.c.String()
		if s != test.s {
			t.Fatalf("#%d: wanted %q, got %q", i, test.s, s)
		}
	}
}
