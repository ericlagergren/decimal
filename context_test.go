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

func TestToNearestAwayQuoRounding(t *testing.T) {

	for _, test := range [...]struct {
		x        int64
		y        int64
		expected string // x div y round 2
	}{
		{1, 9, "0.11"},
		{2, 9, "0.22"},
		{3, 9, "0.33"},
		{4, 9, "0.44"},
		{5, 9, "0.56"},
		{6, 9, "0.67"},
		{7, 9, "0.78"},
		{8, 9, "0.89"},
	} {

		x := New(test.x, 1)
		y := New(test.y, 1)

		// Setup Precision=2 and ToNearestAway Rounding
		z := New(0, 1)
		z.Context.Precision = 2
		z.Context.RoundingMode = ToNearestAway

		actual := z.Quo(x, y).String()
		expected := test.expected

		if actual != expected {
			t.Errorf("Quo(%d,%d) result %s, expected %s", test.x, test.y, actual, expected)
		}
	}

}

func TestNonStandardRoundingModes(t *testing.T) {
	for i, test := range [...]struct {
		value    int64
		mode     RoundingMode
		expected int64
	}{
		{55, ToNearestTowardZero, 5},
		{25, ToNearestTowardZero, 2},
		{16, ToNearestTowardZero, 2},
		{11, ToNearestTowardZero, 1},
		{10, ToNearestTowardZero, 1},
		{-10, ToNearestTowardZero, -1},
		{-11, ToNearestTowardZero, -1},
		{-16, ToNearestTowardZero, -2},
		{-25, ToNearestTowardZero, -2},
		{-55, ToNearestTowardZero, -5},
		{55, AwayFromZero, 6},
		{25, AwayFromZero, 3},
		{16, AwayFromZero, 2},
		{11, AwayFromZero, 2},
		{10, AwayFromZero, 1},
		{-10, AwayFromZero, -1},
		{-11, AwayFromZero, -2},
		{-16, AwayFromZero, -2},
		{-25, AwayFromZero, -3},
		{-55, AwayFromZero, -6},
	} {
		v := New(test.value, 1)
		v.Context.RoundingMode = test.mode
		r, ok := v.RoundToInt().Int64()
		if !ok {
			t.Fatalf("#%d: failed to convert result to int64", i)
		}
		if test.expected != r {
			t.Fatalf("#%d: wanted %d, got %d", i, test.expected, r)
		}
	}
}
