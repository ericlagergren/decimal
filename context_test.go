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
