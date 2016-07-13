package math

import (
	"fmt"
	"testing"

	"github.com/EricLagergren/decimal"
)

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
		b.SetPrec(test.prec)
		a, ok := b.SetString(test.v)
		if !ok {
			t.Fatal("wanted true, got false")
		}
		z := Sqrt(a, a)
		if zs := z.String(); zs != test.sqrt {
			fmt.Println(*z)
			t.Fatalf("#%d: Sqrt(%s): got %s, wanted %q", i, test.v, zs, test.sqrt)
		}
	}
}

func New(s string) *decimal.Big {
	b, ok := new(decimal.Big).SetString(s)
	if !ok {
		panic("!ok")
	}
	return b
}

func BenchmarkSqrt_PerfectSquare(b *testing.B) {
	bench(b.N, []*decimal.Big{
		New("1"), New("4"), New("9"), New("16"), New("36"),
		New("49"), New("64"), New("81"), New("100"), New("121"),
	}, 0)
}

func BenchmarkSqrt_FastPath(b *testing.B)    { bench(b.N, cs, 8) }
func BenchmarkSqrt_DefaultPath(b *testing.B) { bench(b.N, cs, 25) }

var cs = []*decimal.Big{
	New("75"), New("57"), New("23"), New("250"), New("111"),
	New("3.1"), New("69"), New("4.12e-1"),
}

var b decimal.Big

func bench(n int, cs []*decimal.Big, prec int32) {
	b.SetPrec(prec)
	for i := 0; i < n; i++ {
		Sqrt(&b, cs[i%len(cs)])
	}
}
