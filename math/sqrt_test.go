package math

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestDecimal_Hypot(t *testing.T) {
	tests := [...]struct {
		p, q *decimal.Big
		c    int32
		a    string
	}{
		0: {decimal.New(1, 0), decimal.New(4, 0), 15, "4.12310562561766"},
		1: {decimal.New(1, 0), decimal.New(4, 0), 10, "4.1231056256"},
		2: {Pi, Pi, 75, "4.442882938158366247015880990060693698614621689375690223085395606956434793099"},
		3: {decimal.New(-12, 0), decimal.New(599, 0), 5, "599.12"},
		4: {decimal.New(1234, 3), decimal.New(987654123, 5), 6, "9876.54"},
		5: {decimal.New(3, 0), decimal.New(4, 0), 0, "5"},
	}
	var a decimal.Big
	for i, v := range tests {
		a.SetPrecision(v.c)
		if got := Hypot(&a, v.p, v.q).String(); got != v.a {
			t.Errorf("#%d Hypot(%s, %s): wanted %q, got %q",
				i, v.p, v.q, v.a, got)
		}
	}
}

func TestSqrt(t *testing.T) {
	for i, test := range [...]struct {
		v    string
		sqrt string
		prec int32
	}{
		0:  {"25", "5", 0},
		1:  {"100", "10", 0},
		2:  {"250", "15.8113883008418966", 16},
		3:  {"1000", "31.6227766016837933", 16},
		4:  {"1000", "31.6227766016837933199889354", 25},
		5:  {"1000", "31.6227766016837933199889354443271853371955513932521682685750485279259443863923822134424810837930029518", 100},
		6:  {"4.9790119248836735e+00", "2.2313699659365484746324612", 25},
		7:  {"7.7388724745781045e+00", "2.7818829009464263393517169", 25},
		8:  {"9.6362937071984173e+00", "3.1042380236055380970754451", 25},
		9:  {"2.9263772392439646e+00", "1.7106657298385224271646351", 25},
		10: {"5.2290834314593066e+00", "2.2867189227054790347124042", 25},
		11: {"2.7279399104360102e+00", "1.651647635071115948104434", 25},
		12: {"1.8253080916808550e+00", "1.3510396336454586038718314", 25},
	} {
		var b decimal.Big
		b.SetPrecision(test.prec)
		a, ok := b.SetString(test.v)
		if !ok {
			t.Fatal("wanted true, got false")
		}
		Sqrt(a, a)
		if zs := a.String(); zs != test.sqrt {
			t.Fatalf("#%d: Sqrt(%s): got %s, wanted %q", i, test.v, zs, test.sqrt)
		}
	}
}

func BenchmarkSqrt_PerfectSquare(b *testing.B) {
	bench(b.N, []*decimal.Big{
		newbig("1"), newbig("4"), newbig("9"), newbig("16"), newbig("36"),
		newbig("49"), newbig("64"), newbig("81"), newbig("100"), newbig("121"),
	}, 0)
}

func BenchmarkSqrt_FastPath(b *testing.B)    { bench(b.N, cs, 8) }
func BenchmarkSqrt_DefaultPath(b *testing.B) { bench(b.N, cs, 25) }

var cs = []*decimal.Big{
	newbig("75"), newbig("57"), newbig("23"), newbig("250"), newbig("111"),
	newbig("3.1"), newbig("69"), newbig("4.12e-1"),
}

var _b decimal.Big

func bench(n int, cs []*decimal.Big, prec int32) {
	_b.SetPrecision(prec)
	for i := 0; i < n; i++ {
		Sqrt(&_b, cs[i%len(cs)])
	}
}

func newbig(s string) *decimal.Big {
	m, _ := new(decimal.Big).SetString(s)
	return m
}
