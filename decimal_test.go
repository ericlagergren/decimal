package decimal

import (
	"strconv"
	"testing"
)

func TestBig_Add(t *testing.T) {
	type inp struct {
		a   string
		b   string
		res string
	}

	inputs := []inp{
		0: {a: "2", b: "3", res: "5"},
		1: {a: "2454495034", b: "3451204593", res: "5905699627"},
		2: {a: "24544.95034", b: ".3451204593", res: "24545.2954604593"},
		3: {a: ".1", b: ".1", res: "0.2"},
		4: {a: ".1", b: "-.1", res: "0"},
		5: {a: "0", b: "1.001", res: "1.001"},
		6: {a: "123456789123456789.12345", b: "123456789123456789.12345", res: "246913578246913578.2469"},
		7: {a: ".999999999", b: ".00000000000000000000000000000001", res: "0.99999999900000000000000000000001"},
	}

	for i, inp := range inputs {
		a, ok := new(Big).SetString(inp.a)
		if !ok {
			t.FailNow()
		}
		b, ok := new(Big).SetString(inp.b)
		if !ok {
			t.FailNow()
		}
		c := a.Add(a, b)
		if cs := c.String(); cs != inp.res {
			t.Errorf("#%d: expected %s, got %s", i, inp.res, cs)
		}
	}
}

func TestBig_BitLen(t *testing.T) {
	var x Big
	const maxCompact = (1<<63 - 1) - 1
	tests := [...]struct {
		a *Big
		b int
	}{
		0:  {a: New(0, 0), b: 0},
		1:  {a: New(12, 0), b: 4},
		2:  {a: New(50, 0), b: 6},
		3:  {a: New(12345, 0), b: 14},
		4:  {a: New(123456789, 0), b: 27},
		5:  {a: New(maxCompact, 0), b: 63},
		6:  {a: x.Add(New(maxCompact, 0), New(maxCompact, 0)), b: 64},
		7:  {a: New(1000, 0), b: 10},
		8:  {a: New(10, -2), b: 10},
		9:  {a: New(1e6, 0), b: 20},
		10: {a: New(10, -5), b: 20},
		11: {a: New(1e8, 0), b: 27},
		12: {a: New(10, -7), b: 27},
	}
	for i, v := range tests {
		if b := v.a.BitLen(); b != v.b {
			t.Errorf("#%d: wanted %d, got %d", i, v.b, b)
		}
	}
}

func TestBig_Neg(t *testing.T) {
	tests := [...]struct {
		a, b *Big
	}{
		0: {a: New(1, 0), b: New(-1, 0)},
		1: {a: New(999999999999, -1000), b: New(-999999999999, -1000)},
		2: {a: New(-512, 2), b: New(512, 2)},
	}
	var b Big
	for i, v := range tests {
		b.Neg(v.a)

		bs := v.b.String()
		if gs := b.String(); gs != bs {
			t.Fatalf("#%d: wanted %q, got %q", i, bs, gs)
		}
	}
}

func TestBig_Mul(t *testing.T) {
	tests := [...]struct {
		a *Big
		b *Big
		c string
	}{
		0: {a: New(0, 0), b: New(0, 0), c: "0"},
		1: {a: New(12345, 3), b: New(54321, 3), c: "670.592745"},
		2: {a: New(1, -8), b: New(2, 0), c: "0.0000000200"},
	}
	for i, v := range tests {
		r := v.a.Mul(v.a, v.b)
		if s := r.String(); r != v.c {
			t.Fatalf("#%d: wanted %q got %q", i, v.c, s)
		}
	}
}

func TestBig_Prec(t *testing.T) {
	// confirmed to work inside internal/arith/intlen_test.go
}

func TestBig_String(t *testing.T) {
	tests := [...]struct {
		a *Big
		b string
	}{
		0: {a: New(10, 1), b: "1"},                  // Trim trailing zeros
		1: {a: New(12345, 3), b: "12.345"},          // Normal decimal
		2: {a: New(-9876, 2), b: "-98.76"},          // Negative
		3: {a: New(-1e5, 0), b: strconv.Itoa(-1e5)}, // Large number
	}
	for i, s := range tests {
		str := s.a.String()
		if str != s.b {
			t.Fatalf("#%d: wanted %q, got %q", i, s.b, str)
		}
	}
}

func TestDecimal_Sub(t *testing.T) {
	inputs := [...]struct {
		a string
		b string
		r string
	}{
		0: {a: "2", b: "3", r: "-1"},
		1: {a: "12", b: "3", r: "9"},
		2: {a: "-2", b: "9", r: "-11"},
		3: {a: "2454495034", b: "3451204593", r: "-996709559"},
		4: {a: "24544.95034", b: ".3451204593", r: "24544.6052195407"},
		5: {a: ".1", b: "-.1", r: "0.2"},
		6: {a: ".1", b: ".1", r: "0"},
		7: {a: "0", b: "1.001", r: "-1.001"},
		8: {a: "1.001", b: "0", r: "1.001"},
		9: {a: "2.3", b: ".3", r: "2"},
	}

	for i, inp := range inputs {
		a, ok := new(Big).SetString(inp.a)
		if !ok {
			t.FailNow()
		}
		b, ok := new(Big).SetString(inp.b)
		if !ok {
			t.FailNow()
		}
		c := a.Sub(a, b)
		if cs := c.String(); cs != inp.r {
			t.Errorf("#%d: expected %s, got %s", i, inp.r, cs)
		}
	}
}
