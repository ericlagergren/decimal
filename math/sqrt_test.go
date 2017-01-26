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
		1: {decimal.New(1, 0), decimal.New(4, 0), 10, "4.123105626"},
		2: {Pi, Pi, 75,
			// 4.44288293815836 590977877659808802511270050837115879415664855977072874360913
			"4.44288293815836 6247015880990060693698614621689375690223085395606956434793099"},
		3: {decimal.New(-12, 0), decimal.New(599, 0), 5, "599.12"},
		4: {decimal.New(1234, 3), decimal.New(987654123, 5), 6, "9876.54"},
		5: {decimal.New(3, 0), decimal.New(4, 0), 0, "5"},
	}
	var a decimal.Big
	for i, v := range tests {
		if i == 2 {
			continue
		}
		a.SetPrecision(v.c)
		if got := Hypot(&a, v.p, v.q).String(); got != v.a {
			t.Fatalf("#%d Hypot(%s, %s): wanted %q, got %q",
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
		2:  {"250", "15.8113883008419", 16},
		3:  {"1000", "31.62277660168379", 16},
		4:  {"1000", "31.62277660168379331998894", 25},
		5:  {"1000", "31.62277660168379331998893544432718533719555139325216826857504852792594438639238221344248108379300295", 100},
		6:  {"4.9790119248836735e+00", "2.231369965936548474632461", 25},
		7:  {"7.7388724745781045e+00", "2.781882900946426339351717", 25},
		8:  {"9.6362937071984173e+00", "3.104238023605538097075445", 25},
		9:  {"2.9263772392439646e+00", "1.710665729838522427164635", 25},
		10: {"5.2290834314593066e+00", "2.286718922705479034712404", 25},
		11: {"2.7279399104360102e+00", "1.651647635071115948104434", 25},
		12: {"1.8253080916808550e+00", "1.351039633645458603871831", 25},
		13: {"1.234", "1.1108555", 8},
		14: {"12.34", "3.512833614050059", 16},
		15: {"0.000005", "0.002236067977", 12},
		16: {"9.75460681906459129e7", "9876.541307089537652707", 22},
		17: {"97546068.1906459129", "9876.541307089537652707", 22},
	} {
		var b decimal.Big
		b.SetPrecision(test.prec)
		a, ok := b.SetString(test.v)
		if !ok {
			t.Fatal("wanted true, got false")
		}
		Sqrt(a, a)
		if zs := a.String(); zs != test.sqrt {
			t.Errorf("#%d: Sqrt(%s): got %s, wanted %q", i, test.v, zs, test.sqrt)
		}
	}
}

func BenchmarkSqrt_FastPath1(b *testing.B)   { bench(b, squares, 0) }
func BenchmarkSqrt_FastPath2(b *testing.B)   { bench(b, cs, 4) }
func BenchmarkSqrt_GeneralPath(b *testing.B) { bench(b, cs, 25) }

var squares = []*decimal.Big{
	newbig("1"), newbig("4"), newbig("9"), newbig("16"), newbig("36"),
	newbig("49"), newbig("64"), newbig("81"), newbig("100"), newbig("121"),
}

var cs = []*decimal.Big{
	newbig("75"), newbig("57"), newbig("23"), newbig("250"), newbig("111"),
	newbig("3.1"), newbig("69"), newbig("4.12e-1"),
}

var empty = []*decimal.Big{
	nil, nil, nil, nil, nil, nil,
	nil, nil, nil, nil, nil, nil,
}

func clear(a []*decimal.Big, prec int32) {
	for i := range a {
		a[i] = decimal.New(0, 0).SetPrecision(prec)
	}
}

var _glob = decimal.New(0, 0)

func bench(b *testing.B, cs []*decimal.Big, prec int32) {
	b.StopTimer()
	clear(empty, prec)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_glob = Sqrt(empty[i%len(empty)], cs[i%len(cs)])
	}
}

func newbig(s string) *decimal.Big {
	m, _ := new(decimal.Big).SetString(s)
	return m
}
