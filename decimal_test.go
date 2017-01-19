package decimal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

func didPanic(f func()) (ok bool) {
	defer func() { ok = recover() != nil }()
	f()
	return ok
}

func newbig(t *testing.T, s string) *Big {
	x, ok := new(Big).SetString(s)
	if !ok {
		if t == nil {
			panic("wanted true got false during set")
		}
		t.Fatal("wanted true got false during set")
	}
	testFormZero(t, x, "newbig")
	return x
}

var bigZero = new(Big)

// testFormZero verifies that if z == 0, z.form == zero.
func testFormZero(t *testing.T, z *Big, name string) {
	iszero := z.Cmp(bigZero) == 0
	if iszero && z.form != zero {
		t.Errorf("%s: z == 0, but form not marked zero: %v", name, z.form)
	}
	if !iszero && z.form == zero {
		t.Errorf("%s: z != 0, but form marked zero", name)
	}
}

// Verify that ErrNaN implements the error interface.
var _ error = ErrNaN{}

func TestBig_Abs(t *testing.T) {
	for i, test := range [...]string{"-1", "1", "50", "-50", "0", "-0"} {
		x := newbig(t, test)
		if test[0] == '-' {
			test = test[1:]
		}
		if xs := x.Abs(x).String(); xs != test {
			t.Fatalf("#%d: wanted %s, got %s", i, test, xs)
		}
	}
}

func TestBig_Add(t *testing.T) {
	type inp struct {
		a   string
		b   string
		res string
	}

	inputs := [...]inp{
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
			t.Errorf("#%d: wanted %s, got %s", i, inp.res, cs)
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

func TestBig_Cmp(t *testing.T) {
	const (
		lesser  = -1
		equal   = 0
		greater = +1
	)

	samePtr := New(0, 0)
	large, ok := new(Big).SetString(strings.Repeat("9", 500))
	if !ok {
		t.Fatal(ok)
	}
	for i, test := range [...]struct {
		a, b *Big
		v    int
	}{
		// Simple
		0: {New(1, 0), New(0, 0), greater},
		1: {New(0, 0), New(1, 0), lesser},
		2: {New(0, 0), New(0, 0), equal},
		// Fractional
		3: {New(9876, 3), New(1234, 0), lesser},
		4: {New(1234, 3), New(50, 25), greater},
		// Same pointers
		5: {samePtr, samePtr, equal},
		// Large int vs large big.Int
		6: {New(99999999999, 0), large, lesser},
		7: {large, New(999999999999999999, 0), greater},
		8: {New(4, 0), New(4, 0), equal},
		9: {New(4, 0), new(Big).Quo(New(12, 0), New(3, 0)), equal},
		// z.scale < 0
		10: {large, new(Big).Set(large), equal},
		// Differing signs
		11: {new(Big).Set(large).Neg(large), large, lesser},
		12: {new(Big).Quo(new(Big).Set(large), New(314156, 5)), large, lesser},
		13: {New(1234, 3), newbig(t, "1000000000000000000000000000000.234"), lesser},

		// Broken tests
		// Cmp does not compare non-compact numbers of different scale correctly.
		14: {newbig(t, "10000000000000000000"),
			newbig(t, "100000000000000000000").SetScale(1), equal},
	} {
		r := test.a.Cmp(test.b)
		if test.v != r {
			t.Fatalf("#%d: wanted %d, got %d", i, test.v, r)
		}
	}
}

/*func TestBig_Exp(t *testing.T) {
	tests := []struct {
		dec  string
		exp  string
		prec int32
	}{
		0:  {"-8.748656950366438", "0.000158675", 6},
		1:  {"40.40850241721978", "354151937244564830", 18},
		2:  {"73.30000879940332", "6.82007805E+31", 9},
		3:  {"35.89159984662575", "3868332175374135.326826", 22},
		4:  {"-4.1512363035379", "0.0157449389235511838", 18},
		5:  {"-68.12323977553022", "2.59688595E-30", 9},
		6:  {"-60.614962073263406", "4.734307E-27", 7},
		7:  {"-4.865041952853346", "0.0077115046651", 11},
		8:  {"19.704966352217582", "361208659.046814473743298", 24},
		9:  {"-21.85578630459976", "3.222201E-10", 7},
		10: {"82.87588357365792", "9.8296695672260140384636E+35", 23},
		11: {"-25.506698605453636", "8.36722685890E-12", 12},
		12: {"-76.89354159563261", "4.0323590E-34", 8},
		13: {"-70.2633346084568", "3.055072349E-31", 10},
		14: {"-21.75372021081381", "3.56844782783E-10", 12},
		15: {"2.6624827767715686", "14.331827692113042", 17},
		16: {"-96.83919622158838", "8.7754914822403566323918E-43", 23},
		17: {"97.54660128490326", "2.311802E+42", 7},
		18: {"19.67234900470102", "349617061.9295282556", 19},
		19: {"-19.988601487526466", "2.0847821167279748110589E-9", 23},
		20: {"-61.56525338816619", "1.830417572784093673193404E-27", 25},
		21: {"-29.48332735888171", "1.5687495703867765401E-13", 20},
		22: {"-84.74682272069396", "1.5664716288673E-37", 14},
		23: {"-5.141987940031129", "0.00584606", 6},
		24: {"-59.64186269471252", "1.2527607590076679E-26", 17},
		25: {"57.01140301919159", "5.750925436484517E+24", 16},
		26: {"-53.47126566461959", "5.994105485396352375E-24", 19},
		27: {"94.39473267778467", "9.888070E+40", 7},
		28: {"-1.5172773737968157", "0.21930817", 8},
		29: {"-59.57754736169733", "1.3360E-26", 5},
		30: {"-57.08958595213939", "1.60808072677E-25", 12},
		31: {"73.65129808384759", "9.6906E+31", 5},
		32: {"-51.00479595622606", "7.061526050698371484E-23", 19},
		33: {"-78.34101448930855", "9.48264955E-35", 9},
		34: {"-94.76401480997879", "6.99054901284194E-42", 15},
		35: {"-64.30445473402426", "1.182851288281360976177090E-28", 25},
		36: {"-84.83774023774372", "1.4303343141056497E-37", 17},
		37: {"-65.41153068461759", "3.90960760510178E-29", 15},
		38: {"52.32265526524813", "5.289814713107165786531E+22", 22},
		39: {"0.2856256494736158", "1.330594253347893", 16},
		40: {"-53.73245080200248", "4.61629035852671E-24", 15},
		41: {"95.05660578698794", "1.91672303300E+41", 12},
		42: {"27.37684913226701", "775558407201.332", 15},
		43: {"-72.62941915220554", "2.867107906457210561E-32", 19},
		44: {"-31.77381246319696", "1.58784672711822E-14", 15},
		45: {"48.19485014316953", "852623843246004296497.0344", 25},
		46: {"-26.63866583913405", "2.6975805448955996221E-12", 20},
		47: {"0.8074038069587886", "2.2421", 5},
		48: {"-35.836180275711826", "2.7324024E-16", 8},
		49: {"-48.751960790015346", "6.71881134599976028512284E-22", 24},
	}
	for i, v := range tests {
		x := new(Big).SetPrecision(v.prec)
		a := newbig(t, v.dec)
		x.Exp(a)
		xs := string(x.format(true, upper))
		if xs != v.exp {
			t.Fatalf("#%d: exp(%s): wanted %s, got %s", i, v.dec, v.exp, xs)
		}
	}
}*/

func TestBig_IsBig(t *testing.T) {
	for i, test := range [...]struct {
		a   *Big
		big bool
	}{
		0: {newbig(t, "100"), false},
		1: {newbig(t, "-100"), false},
		2: {newbig(t, "5000"), false},
		3: {newbig(t, "-5000"), false},
		4: {newbig(t, "9999999999999999999999999999"), true},
		5: {newbig(t, "1000.5000E+500"), true},
		6: {newbig(t, "1000.5000E-500"), true},
	} {
		if ib := test.a.IsBig(); ib != test.big {
			t.Fatalf("#%d: wanted %t, got %t", i, test.big, ib)
		}
	}
}

func TestBig_Int(t *testing.T) {
	for i, test := range [...]string{
		"1.234", "4.567", "11111111111111111111111111111111111.2",
		"1234234.2321", "121111111111", "44444444.241", "1241.1",
		"4", "5123", "1.2345123134123414123123213", "0.11", ".1",
	} {
		a, ok := new(Big).SetString(test)
		if !ok {
			t.Fatalf("#%d: !ok", i)
		}
		iv := test
		switch x := strings.IndexByte(test, '.'); {
		case x > 0:
			iv = test[:x]
		case x == 0:
			iv = "0"
		}
		n := a.Int()
		if n.String() != iv {
			t.Fatalf("#%d: wanted %q, got %q", i, iv, n.String())
		}
	}
}

func TestBig_Int64(t *testing.T) {
	for i, test := range [...]string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
		"100", "200", "300", "400", "500", "600", "700", "800", "900",
		"1000", "2000", "4000", "5000", "6000", "7000", "8000", "9000",
		"1000000", "2000000", "-12", "-500", "-13123213", "12.000000",
	} {
		a, ok := new(Big).SetString(test)
		if !ok {
			t.Fatalf("#%d: !ok", i)
		}
		iv := test
		switch x := strings.IndexByte(test, '.'); {
		case x > 0:
			iv = test[:x]
		case x == 0:
			iv = "0"
		}
		n := a.Int64()
		if ns := strconv.FormatInt(n, 10); ns != iv {
			t.Fatalf("#%d: wanted %q, got %q", i, iv, ns)
		}
	}
}

func TestBig_IsInt(t *testing.T) {
	for i, test := range [...]string{
		"0 int",
		"-0 int",
		"1 int",
		"-1 int",
		"0.5",
		"1.23",
		"1.23e1",
		"1.23e2 int",
		"0.000000001e+8",
		"0.000000001e+9 int",
		"1.2345e200 int",
		"Inf",
		"+Inf",
		"-Inf",
		"-inf",
	} {
		s := strings.TrimSuffix(test, " int")
		x, ok := new(Big).SetString(s)
		if !ok {
			t.Fatal("TestBig_IsInt !ok")
		}
		want := s != test
		if got := x.IsInt(); got != want {
			t.Errorf("#%d: %s.IsInt() == %t", i, s, got)
		}
	}
}

// func TestBig_Format(t *testing.T) {
// 	tests := [...]struct {
// 		format string
// 		a      string
// 		b      string
// 	}{
// 		0: {format: "%e", a: "1.234", b: "1.234"},
// 		1: {format: "%s", a: "1.2134124124", b: "1.2134124124"},
// 		2: {format: "%e", a: "1.00003e-12", b: "1.00003e-12"},
// 		3: {format: "%E", a: "1.000003E-12", b: "1.000003E-12"},
// 	}
// 	for i, v := range tests {
// 		x, ok := new(Big).SetString(v.a)
// 		if !ok {
// 			t.Fatal("invalid SetString")
// 		}
// 		if fs := fmt.Sprintf(v.format, x); fs != v.b {
// 			t.Fatalf("#%d: wanted %q, got %q:", i, v.b, fs)
// 		}
// 	}
// }

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

func TestBig_Modf(t *testing.T) {
	tests := [...]struct {
		dec  string
		intg string
		frac string
	}{
		0:  {"296474.3772789836", "296474", "0.3772789836"},
		1:  {"-317556.65040295396", "-317556", "-0.65040295396"},
		2:  {"832564.9806826359", "832564", "0.9806826359"},
		3:  {"934740.1797369102", "934740", "0.1797369102"},
		4:  {"520774.2085155975", "520774", "0.2085155975"},
		5:  {"487789.5461755025", "487789", "0.5461755025"},
		6:  {"-562938.1338471398", "-562938", "-0.1338471398"},
		7:  {"77234.6153124352", "77234", "0.6153124352"},
		8:  {"-190793.61336112698", "-190793", "-0.61336112698"},
		9:  {"-117612.82472773292", "-117612", "-0.82472773292"},
		10: {"562490.1936370123", "562490", "0.1936370123"},
		11: {"-454463.82465413434", "-454463", "-0.82465413434"},
		12: {"-468197.66342017427", "-468197", "-0.66342017427"},
		13: {"-45385.99390947411", "-45385", "-0.99390947411"},
		14: {"108976.52613041783", "108976", "0.52613041783"},
		15: {"-883838.5535850908", "-883838", "-0.5535850908"},
		16: {"710903.7632352882", "710903", "0.7632352882"},
		17: {"-647550.5910254063", "-647550", "-0.5910254063"},
		18: {"929294.3969446819", "929294", "0.3969446819"},
		19: {"367607.3691864349", "367607", "0.3691864349"},
		20: {"377847.0999681826", "377847", "0.0999681826"},
		21: {"-921604.174125825", "-921604", "-0.174125825"},
		22: {"76410.7862004172", "76410", "0.7862004172"},
		23: {"-104096.36392393638", "-104096", "-0.36392393638"},
		24: {"940700.46632265", "940700", "0.46632265"},
		25: {"-536862.4033232268", "-536862", "-0.4033232268"},
		26: {"675503.2444769265", "675503", "0.2444769265"},
		27: {"737754.1066881085", "737754", "0.1066881085"},
		28: {"-812094.9541646338", "-812094", "-0.9541646338"},
		29: {"577545.4240398374", "577545", "0.4240398374"},
		30: {"-573554.9376775054", "-573554", "-0.9376775054"},
		31: {"-546642.96324421", "-546642", "-0.96324421"},
		32: {"162519.7570301781", "162519", "0.7570301781"},
		33: {"612961.6010606149", "612961", "0.6010606149"},
		34: {"196102.13226522505", "196102", "0.13226522505"},
		35: {"832033.3624345269", "832033", "0.3624345269"},
		36: {"-344966.9944758781", "-344966", "-0.9944758781"},
		37: {"-933325.047928792", "-933325", "-0.047928792"},
		38: {"-202012.0305155943", "-202012", "-0.0305155943"},
		39: {"424408.7393911381", "424408", "0.7393911381"},
		40: {"667823.2651242441", "667823", "0.2651242441"},
		41: {"270847.74042637344", "270847", "0.74042637344"},
		42: {"-829152.1560524672", "-829152", "-0.1560524672"},
		43: {"-666573.9418051436", "-666573", "-0.9418051436"},
		44: {"905488.3131974128", "905488", "0.3131974128"},
		45: {"439093.1743144549", "439093", "0.1743144549"},
		46: {"-357757.9398403404", "-357757", "-0.9398403404"},
		47: {"-705790.0738368309", "-705790", "-0.0738368309"},
		48: {"565130.6182316178", "565130", "0.6182316178"},
		49: {"-358703.4697589793", "-358703", "-0.4697589793"},
		50: {"783249.3845349478", "783249", "0.3845349478"},
		51: {"0", "0", "0"},
		52: {"0.0", "0", "0"},
		53: {"1.0", "1", "0"},
		54: {"0.1", "0", "0.1"},
		55: {"1", "1", "0"},
		56: {"100000000000000000000", "100000000000000000000", "0"},
		57: {"100000000000000000000.1", "100000000000000000000", "0.1"},
		58: {"100000000000000000000.0", "100000000000000000000", "0"},
		59: {"0.000000000000000000001", "0", "0.000000000000000000001"},
	}
	for i, v := range tests {
		dec := newbig(t, v.dec)
		integ, frac := new(Big).Modf(dec)
		m := new(Big).Add(integ, frac)
		vig := newbig(t, v.intg)
		vfr := newbig(t, v.frac)
		if m.Cmp(dec) != 0 || integ.Cmp(vig) != 0 || vfr.Cmp(frac) != 0 {
			t.Fatalf("#%d: Modf(%s) wanted (%s, %s), got (%s, %s)",
				i, v.dec, v.intg, v.frac, integ, frac)
		}
		testFormZero(t, integ, fmt.Sprintf("#%d: integral part", i))
		testFormZero(t, frac, fmt.Sprintf("#%d: fractional part", i))
	}
}

func TestBig_Mul(t *testing.T) {
	for i, v := range mulTestTable {
		d1 := newbig(t, v.d1)
		d2 := newbig(t, v.d2)
		r := d1.Mul(d1, d2)
		if s := r.String(); s != v.res {
			t.Fatalf("#%d: wanted %s got %s", i, v.res, s)
		}
	}
}

func TestBig_Prec(t *testing.T) {
	// confirmed to work inside internal/arith/intlen_test.go
}

func TestBig_Quo(t *testing.T) {
	huge1, ok := new(Big).SetString("12345678901234567890.1234")
	if !ok {
		t.Fatal("invalid")
	}
	huge2, ok := new(Big).SetString("239482394823948239843298432984.4324324234324234324")
	if !ok {
		t.Fatal("invalid")
	}

	huge3, ok := new(Big).SetString("10000000000000000000000000000000000000000")
	if !ok {
		t.Fatal("invalid")
	}
	huge4, ok := new(Big).SetString("10000000000000000000000000000000000000000")
	if !ok {
		t.Fatal("invalid")
	}

	tests := [...]struct {
		a *Big
		b *Big
		p int32
		r string
	}{
		0:  {a: New(10, 0), b: New(2, 0), r: "5"},
		1:  {a: New(1234, 3), b: New(-2, 0), r: "-0.617"},
		2:  {a: New(10, 0), b: New(3, 0), r: "3.333333333333333"},
		3:  {a: New(100, 0), b: New(3, 0), p: 4, r: "33.33"},
		4:  {a: New(-405, 1), b: New(1257, 2), r: "-3.221957040572792"},
		5:  {a: New(-991242141244124, 7), b: New(235325235323, 3), r: "-0.4212222033406559"},
		6:  {a: huge1, b: huge2, r: "5.155150928864855e-11"},
		7:  {a: New(1000, 0), b: New(20, 0), r: "50"},
		8:  {a: huge3, b: huge4, r: "1"},
		9:  {a: New(100, 0), b: New(1, 0), r: "100"},
		10: {a: New(10, 0), b: New(1, 0), r: "10"},
		11: {a: New(1, 0), b: New(10, 0), r: "0.1"},
	}
	for i, v := range tests {
		if v.p != 0 {
			v.a.SetPrecision(v.p)
		} else {
			v.a.SetPrecision(DefaultPrec)
		}
		q := v.a.Quo(v.a, v.b)
		if qs := q.String(); qs != v.r {
			t.Fatalf("#%d: wanted %q, got %q", i, v.r, qs)
		}
	}
}

func TestBig_Round(t *testing.T) {
	for i, test := range [...]struct {
		v   string
		to  int32
		res string
	}{
		{"5.5", 1, "6"},
		{"1.234", 2, "1.2"},
		{"1", 1, "1"},
		{"9.876", 0, "9.876"},
		{"5.65", 2, "5.6"},
		{"5.0002", 2, "5"},
		{"0.000158674", 6, "0.000158674"},
	} {
		bd := newbig(t, test.v)
		if rs := bd.Round(test.to).String(); rs != test.res {
			t.Fatalf("#%d: wanted %s, got %s", i, test.res, rs)
		}
	}
}

func TestBig_SetFloat64(t *testing.T) {
	tests := map[float64]string{
		123.4:          "123.4",
		123.42:         "123.42",
		123.412345:     "123.412345",
		123.4123456:    "123.4123456",
		123.41234567:   "123.41234567",
		123.412345678:  "123.412345678",
		123.4123456789: "123.4123456789",
	}

	// add negatives
	for p, s := range tests {
		if p > 0 {
			tests[-p] = "-" + s
		}
	}

	var d Big
	for input, s := range tests {
		d.SetFloat64(input)
		if ds := d.String(); ds != s {
			t.Errorf("wanted %s, got %s", s, ds)
		}
	}

	if !didPanic(func() { d.SetFloat64(math.NaN()) }) {
		t.Fatalf("wanted panic when creating a Big from NaN, got %s instead",
			d.String())
	}

	if testing.Short() {
		return
	}

	var err float64
	for f, s := range testTable {
		d.SetFloat64(f)
		if d.String() != s {
			err++
		}
	}

	// Some margin of error is acceptable when converting from
	// a float. On a table of roughly 9,000 entries an acceptable
	// margin of error is around 450. Using Gaussian/banker's rounding our
	// margin of error is roughly 215 per 9,000 entries, for a rate of around
	// 2.3%.
	if err >= 0.05*float64(len(testTable)) {
		t.Errorf("wanted error rate to be < 0.05%% of table, got %.f", err)
	}
}

func TestBig_Sign(t *testing.T) {
	for i, test := range [...]struct {
		x string
		s int
	}{
		0: {"-Inf", 0},
		1: {"-1", -1},
		2: {"-0", 0},
		3: {"+0", 0},
		4: {"+1", +1},
		5: {"+Inf", 0},
		6: {"100", 1},
		7: {"-100", -1},
	} {
		x, ok := new(Big).SetString(test.x)
		if !ok {
			t.Fatal(ok)
		}
		s := x.Sign()
		if s != test.s {
			t.Errorf("#%d: %s.Sign() = %d; want %d", i, test.x, s, test.s)
		}
	}
}

func TestBig_SignBit(t *testing.T) {
	x := New(1<<63-1, 0)
	tests := [...]struct {
		a *Big
		b bool
	}{
		0: {a: New(-1, 0), b: true},
		1: {a: New(1, 0), b: false},
		2: {a: x.Mul(x, x), b: false},
		3: {a: new(Big).Neg(x), b: true},
	}
	for i, v := range tests {
		sb := v.a.Signbit()
		if sb != v.b {
			t.Fatalf("#%d: wanted %t, got %t", i, v.b, sb)
		}
	}
}

func TestBig_String(t *testing.T) {
	x := New(1<<63-1, 0)
	tests := [...]struct {
		a *Big
		b string
	}{
		0: {a: New(10, 1), b: "1"},                  // Trim trailing zeros
		1: {a: New(12345, 3), b: "12.345"},          // Normal decimal
		2: {a: New(-9876, 2), b: "-98.76"},          // Negative
		3: {a: New(-1e5, 0), b: strconv.Itoa(-1e5)}, // Large number
		4: {a: New(0, -50), b: "0"},                 // "0"
		5: {a: x.Mul(x, x), b: "85070591730234615847396907784232501249"},
	}
	for i, s := range tests {
		str := s.a.String()
		if str != s.b {
			t.Fatalf("#%d: wanted %q, got %q", i, s.b, str)
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
		var b Big
		b.SetPrecision(test.prec)
		a, ok := b.SetString(test.v)
		if !ok {
			t.Fatal("wanted true, got false")
		}
		a.Sqrt(a)
		if zs := a.String(); zs != test.sqrt {
			t.Fatalf("#%d: Sqrt(%s): got %s, wanted %q", i, test.v, zs, test.sqrt)
		}
	}
}

func TestBig_Sub(t *testing.T) {
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
			t.Errorf("#%d: wanted %s, got %s", i, inp.r, cs)
		}
	}
}
