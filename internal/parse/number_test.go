package parse

import "testing"

func TestParseSpecial(t *testing.T) {
	for i, test := range [...]struct {
		data string
		s    Special
		sign bool
	}{
		{"inf", Inf, false}, {"-infinity", Inf, true}, {"-iNfInItY", Inf, true},
		{"+INF", Inf, false}, {"nan", QNaN, false}, {"qNaN", QNaN, false},
		{"-SNAN", SNaN, true}, {"na", Invalid, false}, {"-nan123", QNaN, true},
		{"snan3123123", SNaN, false}, {"fnan213123213", Invalid, false},
	} {
		o, sign := ParseSpecial(test.data)
		if o != test.s || sign != test.sign {
			t.Fatalf("#%d: wanted %q:%t, got %q:%t",
				i, test.s, test.sign, o, sign)
		}
	}
}

var (
	S          Special
	B          bool
	benchmarks = [...]string{
		"NaN", "foobar", "infinity", "Infinity", "snan", "+Inf", "-iNfInItY",
		"SNAN", "QNAN", "+inf", "+baaaz", "infinitY", "zzycczz", "not-a-number",
		"-INF", "sNAN", "Snan", "fnI+", "inFINity", "", "nan123123", "qnan231234",
	}
)

func BenchmarkParseSpecial(b *testing.B) {
	var ls Special
	var lb bool
	for i := 0; i < b.N; i++ {
		ls, lb = ParseSpecial(benchmarks[i%len(benchmarks)])
	}
	S = ls
	B = lb
}

func BenchmarkManual(b *testing.B) {
	var ls Special
	var lb bool
	for i := 0; i < b.N; i++ {
		s := benchmarks[i%len(benchmarks)]
		if len(s) > 0 {
			lb = s[0] == '-'
			if lb || s[0] == '+' {
				s = s[1:]
			}
		}
		switch s {
		default:
			ls = Invalid
		case "NaN", "qNaN", "nan", "qnan":
			ls = QNaN
		case "sNaN", "snan":
			ls = SNaN
		case "Inf", "Infinity", "inf", "infinity":
			ls = Inf
		}
	}
	S = ls
	B = lb
}
