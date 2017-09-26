package parse

import "testing"

func TestParseSpecial(t *testing.T) {
	for i, test := range [...]struct {
		data string
		s    Special
	}{
		{"inf", PInf}, {"-infinity", NInf}, {"-iNfInItY", NInf}, {"+INF", PInf},
		{"nan", QNaN}, {"qNaN", QNaN}, {"SNAN", SNaN}, {"na", Invalid},
	} {
		o := ParseSpecial(test.data)
		if o != test.s {
			t.Fatalf("#%d: wanted %q, got %q", i, test.s, o)
		}
	}
}

var (
	S          Special
	benchmarks = [...]string{
		"NaN", "foobar", "infinity", "Infinity", "snan", "+Inf", "-iNfInItY",
		"SNAN", "QNAN", "+inf", "+baaaz", "infinitY", "zzycczz", "not-a-number",
		"-INF", "sNAN", "Snan", "fnI+", "inFINity", "",
	}
)

func BenchmarkParseSpecial(b *testing.B) {
	var ls Special
	for i := 0; i < b.N; i++ {
		ls = ParseSpecial(benchmarks[i%len(benchmarks)])
	}
	S = ls
}

func BenchmarkManual(b *testing.B) {
	var ls Special
	for i := 0; i < b.N; i++ {
		switch benchmarks[i%len(benchmarks)] {
		default:
			ls = Invalid
		case "NaN", "qNaN", "nan", "qnan":
			ls = QNaN
		case "sNaN", "snan":
			ls = SNaN
		case "Inf", "+Inf", "Infinity", "+Infinity",
			"+inf", "+infinity", "inf":
			ls = PInf
		case "-Inf", "-Infinity", "-inf", "-infinity":
			ls = NInf
		}
	}
	S = ls
}
