package decimal

import (
	"encoding/json"
	"encoding/xml"
	"math"
	"math/big"
	"testing"
	"time"
)

var testTableScientificNotation = map[string]string{
	"1e9":        "1000000000",
	"2.41E-3":    "0.00241",
	"24.2E-4":    "0.00242",
	"243E-5":     "0.00243",
	"1e-5":       "0.00001",
	"245E3":      "245000",
	"1.2345E-1":  "0.12345",
	"0e5":        "0",
	"0e-5":       "0",
	"123.456e0":  "123.456",
	"123.456e2":  "12345.6",
	"123.456e10": "1234560000000",
}

func init() {
	// add negatives
	for f, s := range testTable {
		if f > 0 {
			testTable[-f] = "-" + s
		}
	}
	for e, s := range testTableScientificNotation {
		if string(e[0]) != "-" && s != "0" {
			testTableScientificNotation["-"+e] = "-" + s
		}
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		a int64
		b int64
		c string
	}{
		{4, 1, "0.4"},
	}
	for _, v := range tests {
		got := New(v.a, v.b)
		if gs := got.String(); gs != v.c {
			t.Errorf("wanted %q got %q (%v)",
				v.c, gs, *got)
		}
	}
}

func TestNewFromFloat(t *testing.T) {
	var err float64
	for f, s := range testTable {
		d := NewFromFloat(f)
		if d.String() != s {
			err++
			// t.Errorf("expected %s, got %s (%d, %d)",
			// 	s, d.String(), d.compact, d.scale)
		}
	}

	// Some margin of error is acceptable when converting from
	// a float. On a table of roughly 9,000 entries an acceptable
	// margin of error is around 450.
	// Currently, using Gaussian/banker's rounding our margin
	// of error is roughly 215 per 9,000 entries, for a rate of
	// around 2.3%.
	if err >= 0.05*float64(len(testTable)) {
		t.Errorf("expected error rate to be < 0.05%% of table, got %.f", err)
	}

	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for _, n := range shouldPanicOn {
		var d *Decimal
		if !didPanic(func() { d = NewFromFloat(n) }) {
			t.Fatalf("expected panic when creating a Decimal from %v, got %v instead", n, d.String())
		}
	}
}

func TestNewFromString(t *testing.T) {
	for _, s := range testTable {
		d, err := NewFromString(s)
		if err != nil {
			t.Errorf("error while parsing %s", s)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.mantissa.String(), d.scale)
		}
	}

	for e, s := range testTableScientificNotation {
		d, err := NewFromString(e)
		if err != nil {
			t.Errorf("error while parsing %s", e)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s (%d, %d)",
				s, d.String(),
				d.compact, d.scale)
		}
	}
}

func TestNewFromStringErrs(t *testing.T) {
	tests := []string{
		"",
		"qwert",
		"-",
		".",
		"-.",
		".-",
		"234-.56",
		"234-56",
		"2-",
		"..",
		"2..",
		"..2",
		".5.2",
		"8..2",
		"8.1.",
		"1e",
		"1-e",
		"1e9e",
		"1ee9",
		"1ee",
		"1eE",
		"1e-",
		"1e-.",
		"1e1.2",
		"123.456e1.3",
		"1e-1.2",
		"123.456e-1.3",
		"123.456Easdf",
	}

	for _, s := range tests {
		_, err := NewFromString(s)

		if err == nil {
			t.Errorf("error expected when parsing %s", s)
		}
	}
}

func TestNewFromFloatWithScale(t *testing.T) {
	type Inp struct {
		float float64
		exp   int64
	}
	tests := map[Inp]string{
		Inp{123.4, 3}:      "123.4",
		Inp{123.4, 1}:      "123.4",
		Inp{123.412345, 0}: "123",
		Inp{123.412345, 2}: "123.41",
		Inp{123.412345, 5}: "123.41234",
		Inp{123.412345, 6}: "123.412345",
		Inp{123.412345, 7}: "123.412345",
	}

	// add negatives
	for p, s := range tests {
		if p.float > 0 {
			tests[Inp{-p.float, p.exp}] = "-" + s
		}
	}

	for input, s := range tests {
		d := NewFromFloatWithScale(input.float, input.exp)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.mantissa.String(), d.scale)
		}
	}

	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for _, n := range shouldPanicOn {
		var d *Decimal
		if !didPanic(func() { d = NewFromFloatWithScale(n, 0) }) {
			t.Fatalf("Expected panic when creating a Decimal from %v, got %v instead", n, d.String())
		}
	}
}

func TestJSON(t *testing.T) {
	for _, s := range testTable {
		var doc struct {
			Amount Decimal `json:"amount"`
		}
		docStr := `{"amount":"` + s + `"}`
		err := json.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else if doc.Amount.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, doc.Amount.String(),
				doc.Amount.mantissa.String(), doc.Amount.scale)
		}

		out, err := json.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}
	}
}

func TestBadJSON(t *testing.T) {
	for _, testCase := range []string{
		"]o_o[",
		"{",
		`{"amount":""`,
		`{"amount":""}`,
		`{"amount":"nope"}`,
		`0.333`,
	} {
		var doc struct {
			Amount Decimal `json:"amount"`
		}
		err := json.Unmarshal([]byte(testCase), &doc)
		if err == nil {
			t.Errorf("expected error, got %+v", doc)
		}
	}
}

func TestXML(t *testing.T) {
	for _, s := range testTable {
		var doc struct {
			XMLName xml.Name `xml:"account"`
			Amount  Decimal  `xml:"amount"`
		}
		docStr := `<account><amount>` + s + `</amount></account>`
		err := xml.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else if doc.Amount.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, doc.Amount.String(),
				doc.Amount.mantissa.String(), doc.Amount.scale)
		}

		out, err := xml.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}
	}
}

func TestBadXML(t *testing.T) {
	for _, testCase := range []string{
		"o_o",
		"<abc",
		"<account><amount>7",
		`<html><body></body></html>`,
		`<account><amount></amount></account>`,
		`<account><amount>nope</amount></account>`,
		`0.333`,
	} {
		var doc struct {
			XMLName xml.Name `xml:"account"`
			Amount  Decimal  `xml:"amount"`
		}
		err := xml.Unmarshal([]byte(testCase), &doc)
		if err == nil {
			t.Errorf("expected error, got %+v", doc)
		}
	}
}

func TestDecimal_Dim(t *testing.T) {
	tests := [...]struct {
		x, y *Decimal
		a    string
	}{
		{New(5, 0), New(0, 0), "5"},
		{New(0, 0), New(5, 0), "0"},
	}
	for i, v := range tests {
		got := new(Decimal).Dim(v.x, v.y)
		if gs := got.String(); gs != v.a {
			t.Errorf("#%d: wanted %q, got %q", i, v.a, gs)
		}
	}
}

func TestDecimal_Exp(t *testing.T) {
	tests := []struct {
		x  *Decimal
		y  *Decimal
		m  *Decimal
		s  string
		sm string
		b  bool
	}{
		{New(69, 0), New(5, 1), nil, "8.3066238629180748", "", true},
		{New(24000, 4), New(3, 0), nil, "13.824", "", true},
		{New(4, 0), New(4, 0), nil, "256", "", true},
		{New(4, 0), New(4, 0), New(3, 0), "256", "1", true},
		{New(500000, 0), New(4, 0), nil, "62500000000000000000000", "", true},
		{New(-123, 2), New(201, 2), nil, "-1.5160351613631016", "", true},
		{New(123, 2), New(201, 2), nil, "1.5160351613631016", "", true},

		// Should (1<<63)-2 * 4 should cause an overflow of our
		// scale while means we'll get nil for d and false.
		{New(0, math.MaxInt64-1), New(4, 0), nil, "<nil>", "", false},
	}

	for i, v := range tests {
		got, ok := v.x.Exp(v.x, v.y, v.m)
		if gs := got.String(); gs != v.s || ok != v.b {
			t.Errorf("#%d:\nexpected %q and %q and %t\n     got %q and %q and %t",
				i, v.s, v.sm, v.b, got, v.m, ok)
		}
	}
}

func TestDecimal_Floor(t *testing.T) {
	type testData struct {
		input    string
		expected string
	}
	tests := []testData{
		{"1.999", "1"},
		{"1", "1"},
		{"1.01", "1"},
		{"0", "0"},
		{"0.9", "0"},
		{"0.1", "0"},
		{"-0.9", "-1"},
		{"-0.1", "-1"},
		{"-1.00", "-1"},
		{"-1.01", "-2"},
		{"-1.999", "-2"},
		{"3777893186295716170956.8", "3777893186295716170956"},
	}
	for _, test := range tests {
		d, _ := NewFromString(test.input)
		expected, _ := NewFromString(test.expected)
		exp := expected.String()
		got := d.Floor(d)
		if gs := got.String(); gs != exp {
			t.Errorf("Floor(%s): got %s, expected %s", test.input, got, exp)
		}
	}
}

func TestDecimal_Sqrt(t *testing.T) {
	tests := []struct {
		v, s, p int64
		a       string
	}{
		{17, -4, 15, "412.310562561766054"},
		{17, -5, 15, "1303.840481040529742"}, // 1
		{17, 0, 10, "4.1231056256"},
		{17, 0, 13, "4.1231056256176"},
		{17, 0, 15, "4.12310562561766"}, // trim trailing zero
		{2500, 0, 0, "50"},
		{1234, 3, 17, "1.11085552615990527"}, // 6
		{69, 0, 40, "8.3066238629180748525842627449074920102322"},
		{0, 0, 0, "0"},
		{1234567890, 3, 50, "1111.1111060555555440541666143353469245878409860134351"}, // 9 trim trailing zero
		{2, 0, 3901, "1.4142135623730950488016887242096980785696718753769480731766797379907324784621070388503875343276415727350138462309122970249248360558507372126441214970999358314132226659275055927557999505011527820605714701095599716059702745345968620147285174186408891986095523292304843087143214508397626036279952514079896872533965463318088296406206152583523950547457502877599617298355752203375318570113543746034084988471603868999706990048150305440277903164542478230684929369186215805784631115966687130130156185689872372352885092648612494977154218334204285686060146824720771435854874155657069677653720226485447015858801620758474922657226002085584466521458398893944370926591800311388246468157082630100594858704003186480342194897278290641045072636881313739855256117322040245091227700226941127573627280495738108967504018369868368450725799364729060762996941380475654823728997180326802474420629269124859052181004459842150591120249441341728531478105803603371077309182869314710171111683916581726889419758716582152128229518488472089694633862891562882765952635140542267653239694617511291602408715510135150455381287560052631468017127402653969470240300517495318862925631385188163478001569369176881852378684052287837629389214300655869568685964595155501644724509836896036887323114389415576651040883914292338113206052433629485317049915771756228549741438999188021762430965206564211827316726257539594717255934637238632261482742622208671155839599926521176252698917540988159348640083457085181472231814204070426509056532333398436457865796796519267292399875366617215982578860263363617827495994219403777753681426217738799194551397231274066898329989895386728822856378697749662519966583525776198939322845344735694794962952168891485492538904755828834526096524096542889394538646625744927556381964410316979833061852019379384940057156333720548068540575867999670121372239475821426306585132217408832382947287617393647467837431960001592188807347857617252211867490424977366929207311096369721608933708661156734585334833295254675851644710757848602463600834449114818587655554286455123314219926311332517970608436559704352856410087918500760361009159465670676883605571740076756905096136719401324935605240185999105062108163597726431380605467010293569971042425105781749531057255934984451126922780344913506637568747760283162829605532422426957534529028838768446429173282770888318087025339852338122749990812371892540726475367850304821591801886167108972869229201197599880703818543332536460211082299279293072871780799888099176741774108983060800326311816427988231171543638696617029999341616148786860180455055539869131151860103863753250045581860448040750241195184305674533683613674597374423988553285179308960373898915173195874134428817842125021916951875593444387396189314549999906107587049090260883517636224749757858858368037457931157339802099986622186949922595913276423619410592100328026149874566599688874067956167391859572888642473463585886864496822386006983352642799056283165613913942557649062065186021647263033362975075697870606606856498160092718709292153132368281356988937097416504474590960537472796524477094099241238710614470543986743647338477454819100872886222149589529591187892149179833981083788278153065562315810360648675873036014502273208829351341387227684176678436905294286984908384557445794095986260742499549168028530773989382960362133539875320509199893607513906444495768456993471276364507163279154701597733548638939423257277540038260274785674172580951416307159597849818009443560379390985590168272154034581581521004936662953448827107292396602321638238266612626830502572781169451035379371568823365932297823192986064679789864092085609558142614363631004615594332550474493975933999125419532300932175304476533964706627611661753518754646209676345587386164880198848497479264045065444896910040794211816925796857563784881498986416854994916357614484047021033989215342377037233353115645944389703653166721949049351882905806307401346862641672470110653463493916407146285"},
	}
	for i, v := range tests {
		a := New(v.v, v.s)
		a.SetContext(Context{Prec: v.p})
		got := a.Sqrt(a)
		if gs := got.String(); gs != v.a {
			t.Errorf("#%d: want %q, got %q", i, v.a, gs)
		}
	}
}

func TestDecimal_Trunc(t *testing.T) {
	tests := []struct {
		a *Decimal
		b int64
		c string
	}{
		{New(1234, 2), 1, "12.3"},
		{New(123456789, 8), 5, "1.23456"},
		{New(123456789, -4), 2, "1234567890000"},
	}
	for _, v := range tests {
		got := v.a.Trunc(v.a, v.b)
		if gs := got.String(); gs != v.c {
			t.Errorf("want %q, got %q", v.c, gs)
		}
	}
}

func TestDecimal_Equals(t *testing.T) {
	tests := [...]struct {
		x, y       *Decimal
		gt, lt, eq bool
	}{
		{New(5, 0), New(5, 0), false, false, true},
		{New(4, 0), New(1, 0), true, false, false},
		{New(1, 0), New(4, 0), false, true, false},
		{New(0, 0), zero, false, false, true},
		{New(0, 1e5), zero, false, false, true},
		{zero, New(0, 1e5), false, false, true},
		{New(-5, 0), New(1234, 3), false, true, false},
		{New(1234, 3), New(-5, 0), true, false, false},
		{New(-5, 0), New(-5, 0), false, false, true},
	}
	for i, v := range tests {
		gt := v.x.GreaterThan(v.y)
		lt := v.x.LessThan(v.y)
		eq := v.x.Equals(v.y)
		if gt != v.gt || lt != v.lt || eq != v.eq {
			t.Errorf("#%d (%s and %s):\n   got gt: %t, lt: %t, and eq: %t\nwanted gt: %t, lt: %t, and eq: %t,",
				i, v.x, v.y, v.gt, v.lt, v.eq, gt, lt, eq)
		}
	}
}

func TestDecimal_Cmp(t *testing.T) {
	cmpTests := [...]struct {
		x, xs int64
		y, ys int64
		r     int
	}{
		{0, 0, 0, 0, 0},
		{1, 0, 1, 0, 0},
		{34986, 3, 56, 0, -1},
		{1234, 3, 1234, 4, 1},
		{-1234, 0, 1234, 4, -1},
		{123, 2, -1234, 0, 1},
	}

	for i, a := range cmpTests {
		x := New(a.x, a.xs)
		y := New(a.y, a.ys)
		r := x.Cmp(y)
		if r != a.r {
			t.Errorf("#%d (%s and %s) got %v; want %v", i, x, y, r, a.r)
		}
	}
}

func TestDecimal_Ceil(t *testing.T) {
	type testData struct {
		input    string
		expected string
	}
	tests := []testData{
		{"1.999", "2"},
		{"1", "1"},
		{"1.01", "2"},
		{"0", "0"},
		{"0.9", "1"},
		{"0.1", "1"},
		{"-0.9", "0"},
		{"-0.1", "0"},
		{"-1.00", "-1"},
		{"-1.01", "-1"},
		{"-1.999", "-1"},
		{"3777893186295716170956.8", "3777893186295716170957"},
	}
	for _, test := range tests {
		d, _ := NewFromString(test.input)
		expected, _ := NewFromString(test.expected)
		exp := expected.String()
		got := d.Ceil(d)
		if gs := got.String(); gs != exp {
			t.Errorf("Ceil(%s): got %s, expected %s", test.input, gs, exp)
		}
	}
}

// func TestDecimal_RoundAndStringFixed(t *testing.T) {
// 	type testData struct {
// 		input         string
// 		places        int64
// 		expected      string
// 		expectedFixed string
// 	}
// 	tests := []testData{
// 		{"1.454", 0, "1", ""},
// 		{"1.454", 1, "1.4", ""},
// 		{"1.454", 2, "1.45", ""},
// 		{"1.454", 3, "1.454", ""},
// 		{"1.454", 4, "1.454", "1.4540"},
// 		{"1.454", 5, "1.454", "1.45400"},
// 		{"1.554", 0, "2", ""},
// 		{"1.554", 1, "1.6", ""},
// 		{"1.554", 2, "1.55", ""},
// 		{"0.554", 0, "1", ""},
// 		{"0.454", 0, "0", ""},
// 		{"0.454", 5, "0.454", "0.45400"},
// 		{"0", 0, "0", ""},
// 		{"0", 1, "0", "0.0"},
// 		{"0", 2, "0", "0.00"},
// 		{"0", -1, "0", ""},
// 		{"5", 2, "5", "5.00"},
// 		{"5", 1, "5", "5.0"},
// 		{"5", 0, "5", ""},
// 		{"500", 2, "500", "500.00"},
// 		{"545", -1, "550", ""},
// 		{"545", -2, "500", ""},
// 		{"545", -3, "1000", ""},
// 		{"545", -4, "0", ""},
// 		{"499", -3, "0", ""},
// 		{"499", -4, "0", ""},
// 	}

// 	// add negative number tests
// 	for _, test := range tests {
// 		expected := test.expected
// 		if expected != "0" {
// 			expected = "-" + expected
// 		}
// 		expectedStr := test.expectedFixed
// 		if strings.ContainsAny(expectedStr, "123456789") && expectedStr != "" {
// 			expectedStr = "-" + expectedStr
// 		}
// 		tests = append(tests,
// 			testData{"-" + test.input, test.places, expected, expectedStr})
// 	}

// 	for _, test := range tests {
// 		d, err := NewFromString(test.input)
// 		if err != nil {
// 			panic(err)
// 		}
// 		tmp := d.Clone()

// 		// test Round
// 		expected, err := NewFromString(test.expected)
// 		if err != nil {
// 			panic(err)
// 		}

// 		d1 := d.Clone()
// 		got := d1.Round(test.places)
// 		if !got.Equals(expected) {
// 			t.Errorf("Rounding %s to %d places, got %s, expected %s",
// 				tmp, test.places, got, expected)
// 		}

// 		// test StringFixed
// 		if test.expectedFixed == "" {
// 			test.expectedFixed = test.expected
// 		}
// 		gotStr := d.StringFixed(test.places)
// 		if gotStr != test.expectedFixed {
// 			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
// 				tmp, test.places, gotStr, test.expectedFixed)
// 		}
// 	}
// }

func TestDecimal_Uninitialized(t *testing.T) {
	a := &Decimal{}
	b := &Decimal{}

	decs := [...]*Decimal{
		New(0, 0),
		// New(0, 0).rescale(10),
		New(0, 0).Abs(a),
		New(0, 0).Add(a, b),
		New(0, 0).Sub(a, b),
		New(0, 0).Mul(a, b),
		// New(0, 0).Div(a, New(1, -1)),
		// New(0, 0).Round(2),
		New(0, 0).Floor(a),
		New(0, 0).Ceil(a),
		New(0, 0).Trunc(a, 2),
	}

	for _, d := range decs {
		if d.String() != "0" {
			t.Errorf("expected 0, got %s", d.String())
		}
		// if d.StringFixed(3) != "0.000" {
		// 	t.Errorf("expected 0, got %s", d.StringFixed(3))
		// }
	}

	if a.Cmp(b) != 0 {
		t.Errorf("a != b")
	}
	if a.Scale() != 0 {
		t.Errorf("a.Scale() != 0")
	}
	if a.Int64() != 0 {
		t.Errorf("a.IntPar() != 0")
	}
	f, _ := a.Float64()
	if f != 0 {
		t.Errorf("a.Float64() != 0")
	}
	if a.Rat(nil).RatString() != "0" {
		t.Errorf("a.Rat() != 0, got %s", a.Rat(nil).RatString())
	}
}

func TestDecimal_Add(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                                               "5",
		Inp{"2454495034", "3451204593"}:                             "5905699627",
		Inp{"24544.95034", ".3451204593"}:                           "24545.2954604593",
		Inp{".1", ".1"}:                                             "0.2",
		Inp{".1", "-.1"}:                                            "0",
		Inp{"0", "1.001"}:                                           "1.001",
		Inp{"123456789123456789.12345", "123456789123456789.12345"}: "246913578246913578.2469",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Add(a, b)
		if cs := c.String(); cs != res {
			t.Errorf("expected %s, got %s", res, cs)
		}
	}
}

func TestDecimal_Sub(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "-1",
		Inp{"12", "3"}:                    "9",
		Inp{"-2", "9"}:                    "-11",
		Inp{"2454495034", "3451204593"}:   "-996709559",
		Inp{"24544.95034", ".3451204593"}: "24544.6052195407",
		Inp{".1", "-.1"}:                  "0.2",
		Inp{".1", ".1"}:                   "0",
		Inp{"0", "1.001"}:                 "-1.001",
		Inp{"1.001", "0"}:                 "1.001",
		Inp{"2.3", ".3"}:                  "2",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Sub(a, b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_BitLen(t *testing.T) {
	for i, v := range [...]struct {
		d1, d2 *Decimal
		a      int64
	}{
		{New(5, 0), new(Decimal).SetInt(big.NewInt(5)), 3},
		{New(30, 0), new(Decimal).SetInt(big.NewInt(30)), 5},
		{New(1234, 0), new(Decimal).SetInt(big.NewInt(1234)), 11},
		{New(69, 0), new(Decimal).SetInt(big.NewInt(69)), 7},
		{New(25, 0), new(Decimal).SetInt(big.NewInt(25)), 5},
		{New(math.MaxInt64, 0), new(Decimal).SetInt(big.NewInt(math.MaxInt64)), 63},
	} {
		if v.d1.BitLen() != v.d2.BitLen() || v.d1.BitLen() != v.a {
			t.Errorf("#%d wanted %d, got %d and %d", i, v.a, v.d1.BitLen(), v.d2.BitLen())
		}
	}
}

func TestDecimal_Binomial(t *testing.T) {
	var z Decimal
	for _, test := range []struct {
		n, k int64
		want string
	}{
		{0, 0, "1"},
		{0, 1, "0"},
		{1, 0, "1"},
		{1, 1, "1"},
		{1, 10, "0"},
		{4, 0, "1"},
		{4, 1, "4"},
		{4, 2, "6"},
		{4, 3, "4"},
		{4, 4, "1"},
		{10, 1, "10"},
		{10, 9, "10"},
		{10, 5, "252"},
		{11, 5, "462"},
		{11, 6, "462"},
		{100, 10, "17310309456440"},
		{100, 90, "17310309456440"},
		{1000, 10, "263409560461970212832400"},
		{1000, 990, "263409560461970212832400"},
	} {
		if got := z.Binomial(test.n, test.k).String(); got != test.want {
			t.Errorf("Binomial(%d, %d) = %s; want %s",
				test.n, test.k, got, test.want)
		}
	}
}

func TestDecimal_Rat(t *testing.T) {
	tests := [...]struct {
		v, s int64
		a    string
	}{
		{5, 0, "5/1"},
		{5, -5, "500000/1"},
		{1234, 3, "617/500"},
	}
	for i, v := range tests {
		x := New(v.v, v.s)
		if xs := x.Rat(nil).String(); xs != v.a {
			t.Errorf("#%d: wanted %s, got %s",
				i, v.a, xs)
		}
	}

	// Test that the non-nil big.Rat is used.
	x := New(5, 0)
	r := new(big.Rat)
	if xs := x.Rat(r).String(); xs != r.String() {
		t.Errorf("wanted %s, got %s and %s", "5/1", xs, r.String())
	}
}

func TestDecimal_MulRange(t *testing.T) {
	mulRangesN := [...]struct {
		a, b int64
		prod string
	}{
		{0, 0, "0"},
		{1, 1, "1"},
		{1, 2, "2"},
		{1, 3, "6"},
		{10, 10, "10"},
		{0, 100, "0"},
		{0, 1e9, "0"},
		{1, 0, "1"},                    // empty range
		{100, 1, "1"},                  // empty range
		{1, 10, "3628800"},             // 10!
		{1, 20, "2432902008176640000"}, // 20!
		{1, 100,
			"933262154439441526816992388562667004907159682643816214685929" +
				"638952175999932299156089414639761565182862536979208272237582" +
				"51185210916864000000000000000000000000", // 100!
		},
	}
	mulRangesZ := [...]struct {
		a, b int64
		prod string
	}{
		// entirely positive ranges are covered by mulRangesN
		{-1, 1, "0"},
		{-2, -1, "2"},
		{-3, -2, "6"},
		{-3, -1, "-6"},
		{1, 3, "6"},
		{-10, -10, "-10"},
		{0, -1, "1"},                      // empty range
		{-1, -100, "1"},                   // empty range
		{-1, 1, "0"},                      // range includes 0
		{-1e9, 0, "0"},                    // range includes 0
		{-1e9, 1e9, "0"},                  // range includes 0
		{-10, -1, "3628800"},              // 10!
		{-20, -2, "-2432902008176640000"}, // -20!
		{-99, -1,
			"-933262154439441526816992388562667004907159682643816214685929" +
				"638952175999932299156089414639761565182862536979208272237582" +
				"511852109168640000000000000000000000", // -99!
		},
	}
	var tmp Decimal
	// test entirely positive ranges
	for i, r := range mulRangesN {
		prod := tmp.MulRange(r.a, r.b).String()
		if prod != r.prod {
			t.Errorf("#%da: got %s; want %s", i, prod, r.prod)
		}
	}
	// test other ranges
	for i, r := range mulRangesZ {
		prod := tmp.MulRange(r.a, r.b).String()
		if prod != r.prod {
			t.Errorf("#%db: got %s; want %s", i, prod, r.prod)
		}
	}
}

func TestDecimal_Neg(t *testing.T) {
	tests := [...]struct {
		v, s int64
		e    string
	}{
		{5, 0, "-5"},
		{-1234, 0, "1234"},
		{-1234, 3, "1.234"},
		{math.MaxInt64, 0, "-9223372036854775807"},
		{-math.MaxInt64, 200, "0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009223372036854775807"},
		{math.MaxInt64, 200, "-0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009223372036854775807"},
	}
	for i, v := range tests {
		x := New(v.v, v.s)
		x.Neg(x)
		if xs := x.String(); xs != v.e {
			t.Errorf("#%d: wanted %s, got %s",
				i, v.e, xs)
		}
	}
}

func TestDecimal_Hypot(t *testing.T) {
	tests := [...]struct {
		p, q *Decimal
		c    int64
		a    string
	}{
		{New(1, 0), New(4, 0), 15, "4.12310562561766"},
		{New(1, 0), New(4, 0), 10, "4.1231056256"},
		{newFromString("3.141592653589793238462643383279502884197169399375105820974944592307816406286"), newFromString("3.141592653589793238462643383279502884197169399375105820974944592307816406286"), 75, "4.442882938158366247015880990060693698614621689375690223085395606956434793099"},
		{New(-12, 0), New(599, 0), 2, "599.12"},
		{New(1234, 3), New(987654123, 5), 2, "9876.54"},
		{New(3, 0), New(4, 0), 0, "5"},
	}
	for i, v := range tests {
		v.p.SetContext(Context{Prec: v.c})
		v.q.SetContext(Context{Prec: v.c})
		if got := Hypot(v.p, v.q).String(); got != v.a {
			t.Errorf("#%d: wanted %q, got %q", i, v.a, got)
		}
	}
}

func TestFizzBuzz(t *testing.T) {
	t.Log("Seriously? Why are you testing this?")
	x := New(100, 0)
	if testing.Verbose() {
		FizzBuzz(x)
	}
	t.Log("You're ridiculous.")
}

func TestDecimal_Fib(t *testing.T) {
	tests := [...]struct {
		x int64
		a string
	}{
		{17, "1597"},
		{30, "832040"},
		{48, "4807526976"},
		{80, "23416728348467685"},
		{110, "43566776258854844738105"},
		{172, "394810887814999156320699623170776339"},
		{262, "2542592393026885507715496646813780220945054040571721231"},
		{300, "222232244629420445529739893461909967206666939096499764990979600"},
	}
	for i, v := range tests {
		x := New(v.x, 0)
		got := x.Fib(x)
		if gs := got.String(); gs != v.a {
			t.Errorf("#%d: wanted %q, got %q", i, v.a, got)
		}
	}

}

// func TestDecimal_Log10(t *testing.T) {
// 	tests := [...]struct {
// 		v, s, p int64
// 		a       string
// 	}{
// 		{1234, 3, 16, "0.0913151596972228"},
// 	}
// 	for i, v := range tests {
// 		x := New(v.v, v.s)
// 		x.SetContext(Context{Prec: v.p})
// 		x.Log10(x)
// 		if got := x.String(); got != v.a {
// 			t.Errorf("#%d want %q, got %q", i, v.a, got)
// 		}
// 	}
// }

func TestDecimal_Mul(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "6",
		Inp{"2454495034", "3451204593"}:   "8470964534836491162",
		Inp{"24544.95034", ".3451204593"}: "8470.964534836491162",
		Inp{".1", ".1"}:                   "0.01",
		Inp{"0", "1.001"}:                 "0",
		Inp{"123400000", "450"}:           "55530000000",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		if s := a.Mul(a, b).String(); s != res {
			t.Errorf("expected %s, got %s", res, s)
		}
	}
}

func TestDecimal_Div(t *testing.T) {
	// a := New(1234, 4)
	// b := New(3456, 4)
	// a.ctx.Prec = 25
	// fmt.Println(a, b)
	// fmt.Println(a.Div(a, b))
	// fmt.Println(new(Decimal).IntDiv(a, b))
}

func TestDecimal_Quo(t *testing.T) {
	quoTests := [...]struct {
		x, y string
		q, r string
	}{
		{
			"476217953993950760840509444250624797097991362735329973741718102894495832294430498335824897858659711275234906400899559094370964723884706254265559534144986498357",
			"9353930466774385905609975137998169297361893554149986716853295022578535724979483772383667534691121982974895531435241089241440253066816724367338287092081996",
			"50911",
			"1",
		},
		{
			"11510768301994997771168",
			"1328165573307167369775",
			"8",
			"885443715537658812968",
		},
	}
	for i, test := range quoTests {
		x, _ := NewFromString(test.x)
		y, _ := NewFromString(test.y)
		expectedQ, _ := NewFromString(test.q)
		// expectedR, _ := NewFromString(test.r)

		// r := new(Decimal)
		// q, r := new(Decimal).QuoRem(x, y, r)
		q := new(Decimal).IntDiv(x, y)

		if !q.Equals(expectedQ) { //|| r.Cmp(expectedR) != 0 {
			t.Errorf("#%d got (%s) want (%s)", // %s %s
				i, q, expectedQ) //r, expectedQ, expectedR)
		}
	}
}

// func TestDecimal_Div(t *testing.T) {
// 	type Inp struct {
// 		a string
// 		b string
// 	}

// 	inputs := map[Inp]string{
// 		Inp{"6", "3"}:                            "2",
// 		Inp{"10", "2"}:                           "5",
// 		Inp{"2.2", "1.1"}:                        "2",
// 		Inp{"-2.2", "-1.1"}:                      "2",
// 		Inp{"12.88", "5.6"}:                      "2.3",
// 		Inp{"1023427554493", "43432632"}:         "23563.5628642767953828", // rounded
// 		Inp{"1", "434324545566634"}:              "0.0000000000000023",
// 		Inp{"1", "3"}:                            "0.3333333333333333",
// 		Inp{"2", "3"}:                            "0.6666666666666667", // rounded
// 		Inp{"10000", "3"}:                        "3333.3333333333333333",
// 		Inp{"10234274355545544493", "-3"}:        "-3411424785181848164.3333333333333333",
// 		Inp{"-4612301402398.4753343454", "23.5"}: "-196268144782.9138440146978723",
// 	}

// 	for inp, expected := range inputs {
// 		num, err := NewFromString(inp.a)
// 		if err != nil {
// 			t.FailNow()
// 		}
// 		denom, err := NewFromString(inp.b)
// 		if err != nil {
// 			t.FailNow()
// 		}
// 		got := num.Div(num, denom)
// 		if got.String() != expected {
// 			t.Errorf("expected %s when dividing %v by %v, got %v",
// 				expected, num, denom, got)
// 		}
// 	}

// 	type Inp2 struct {
// 		n    int64
// 		exp  int64
// 		n2   int64
// 		exp2 int64
// 	}

// 	// test code path where exp > 0
// 	inputs2 := map[Inp2]string{
// 		Inp2{124, 10, 3, 1}: "41333333333.3333333333333333",
// 		Inp2{124, 10, 3, 0}: "413333333333.3333333333333333",
// 		Inp2{124, 10, 6, 1}: "20666666666.6666666666666667",
// 		Inp2{124, 10, 6, 0}: "206666666666.6666666666666667",
// 		Inp2{10, 10, 10, 1}: "1000000000",
// 	}

// 	for inp, expectedAbs := range inputs2 {
// 		for i := -1; i <= 1; i += 2 {
// 			for j := -1; j <= 1; j += 2 {
// 				n := inp.n * int64(i)
// 				n2 := inp.n2 * int64(j)
// 				num := New(n, inp.exp)
// 				denom := New(n2, inp.exp2)
// 				expected := expectedAbs
// 				if i != j {
// 					expected = "-" + expectedAbs
// 				}
// 				got := num.Div(num, denom)
// 				if got.String() != expected {
// 					t.Errorf("expected %s when dividing %v by %v, got %v",
// 						expected, num, denom, got)
// 				}
// 			}
// 		}
// 	}
// }

func TestDecimal_Overflow(t *testing.T) {
	x := New(1, math.MaxInt64)
	if !didPanic(func() {
		x.Mul(x, New(1, math.MaxInt64))
	}) {
		t.Fatalf("should have gotten an overflow panic")
	}
	y := New(1, math.MaxInt64)
	if !didPanic(func() {
		y.Mul(y, New(1, math.MaxInt64))
	}) {
		t.Fatalf("should have gotten an overflow panic")
	}
}

func TestDecimalScale_TooSmall(t *testing.T) {
	if !didPanic(func() {
		New(1, math.MinInt64)
	}) {
		t.Fatalf("should have gotten a too small panic")
	}
}

func TestDecimal_ExtremeValues(t *testing.T) {
	// NOTE(vadim): this test takes pretty much forever
	if testing.Short() {
		t.Skip()
	}

	// NOTE(vadim): Seriously, the numbers invovled are so large that this
	// test will take way too long, so mark it as success if it takes over
	// 1 second. The way this test typically fails (integer overflow) is that
	// a wrong result appears quickly, so if it takes a long time then it is
	// probably working properly.
	// Why even bother testing this? Completeness, I guess. -Vadim
	const timeLimit = 1 * time.Second
	test := func(f func()) {
		c := make(chan bool)
		go func() {
			f()
			close(c)
		}()
		select {
		case <-c:
		case <-time.After(timeLimit):
		}
	}

	test(func() {
		a := New(123, math.MaxInt64)
		got := a.Floor(a)
		if !got.Equals(NewFromFloat(0)) {
			t.Errorf("Error: got %s, expected 0", got)
		}
	})
	test(func() {
		a := New(123, math.MaxInt64)
		got := a.Ceil(a)
		if !got.Equals(NewFromFloat(1)) {
			t.Errorf("Error: got %s, expected 1", got)
		}
	})
	test(func() {
		got := New(123, math.MaxInt64).Rat(nil).FloatString(10)
		expected := "0.0000000000"
		if got != expected {
			t.Errorf("Error: got %s, expected %s", got, expected)
		}
	})
}

func TestJacobi(t *testing.T) {
	testCases := []struct {
		x, y   int64
		result int
	}{
		{0, 1, 1},
		{0, -1, 1},
		{1, 1, 1},
		{1, -1, 1},
		{0, 5, 0},
		{1, 5, 1},
		{2, 5, -1},
		{-2, 5, -1},
		{2, -5, -1},
		{-2, -5, 1},
		{3, 5, -1},
		{5, 5, 0},
		{-5, 5, 0},
		{6, 5, 1},
		{6, -5, 1},
		{-6, 5, 1},
		{-6, -5, -1},
	}

	var x, y Decimal

	for i, test := range testCases {
		x.SetInt64(test.x)
		y.SetInt64(test.y)
		expected := test.result
		actual := Jacobi(&x, &y)
		if actual != expected {
			t.Errorf("#%d: Jacobi(%d, %d) = %d, but expected %d", i, test.x, test.y, actual, expected)
		}
	}
}

func TestJacobiPanic(t *testing.T) {
	const failureMsg = "test failure"
	defer func() {
		msg := recover()
		if msg == nil || msg == failureMsg {
			panic(msg)
		}
		t.Log(msg)
	}()
	x := New(1, 0)
	y := New(2, 0)
	// Jacobi should panic when the second argument is even.
	Jacobi(x, y)
	panic(failureMsg)
}

func TestInt64(t *testing.T) {
	for _, testCase := range []struct {
		Dec   string
		Int64 int64
	}{
		{"0.01", 0},
		{"12.1", 12},
		{"9999.999", 9999},
		{"-32768.01234", -32768},
		// Test overflown.
		{"1.1234567890123456789012345678901234567890", overflown},
	} {
		d, err := NewFromString(testCase.Dec)
		if err != nil {
			t.Fatal(err)
		}
		if d.Int64() != testCase.Int64 {
			t.Errorf("expect %d, got %d", testCase.Int64, d.Int64())
		}
	}
}

func newFromString(s string) *Decimal {
	d, err := NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestDecimal_Modf(t *testing.T) {
	tests := []struct {
		a *Decimal
		b string
		c string
	}{
		{New(12345, 3), "12", "0.345"},
		{newFromString("483570327845851669882.4704"),
			"483570327845851669882",
			"0.4704"},
		{NewFromFloat(1234.567), "1234", "0.567"},
		{New(-1234567890, 4), "-123456", "-0.789"},
		{New(98765, 5), "0", "0.98765"},
	}
	for i, v := range tests {
		gotI, gotF := Modf(v.a)
		expI, expF := v.b, v.c
		if gotI.String() != expI ||
			gotF.String() != expF {
			t.Errorf("#%d: expected %q and %q, got %q and %q",
				i, expI, expF, gotI.String(), gotF.String())
		}
	}
}

func TestDecimal_Min(t *testing.T) {
	for i, v := range [...]struct {
		min, max *Decimal
	}{
		{New(0, 0), New(1, 0)},
		{New(100, 0), New(1000, 0)},
		{New(-500, 0), New(-450, 0)},
		{New(1234, 3), New(1234, 2)},
	} {
		min := Min(v.min, v.max)
		if min != v.min {
			t.Errorf("#%d: want %q, got %q", i, v.min, min)
		}
	}
}

func TestDecimal_Max(t *testing.T) {
	for i, v := range [...]struct {
		min, max *Decimal
	}{
		{New(0, 0), New(1, 0)},
		{New(100, 0), New(1000, 0)},
		{New(-500, 0), New(-450, 0)},
		{New(1234, 3), New(1234, 2)},
	} {
		max := Max(v.min, v.max)
		if max != v.max {
			t.Errorf("#%d: want %q, got %q", i, v.max, max)
		}
	}
}

// old tests after this line

func TestDecimal_Scale(t *testing.T) {
	a := New(1234, -3)
	if a.Scale() != -3 {
		t.Errorf("error")
	}
}

func TestDecimal_Abs1(t *testing.T) {
	a := New(-1234, -4)
	b := New(1234, -4)

	c := a.Abs(a)
	if c.Cmp(b) != 0 {
		t.Errorf("error")
	}
}

func TestDecimal_Abs2(t *testing.T) {
	a := New(-1234, -4)
	b := New(1234, -4)

	c := b.Abs(b)
	if c.Cmp(a) == 0 {
		t.Errorf("error")
	}
}

func TestDecimal_ScalesNotEqual(t *testing.T) {
	a := New(1234, 2)
	b := New(1234, 3)
	if a.Equals(b) {
		t.Errorf("%q should not equal %q", a, b)
	}
}

func TestDecimal_Cmp2(t *testing.T) {
	a := New(123, 3)
	b := New(1234, 2)

	if a.Cmp(b) != -1 {
		t.Errorf("Error")
	}
}

func didPanic(f func()) bool {
	ret := false
	func() {
		defer func() {
			if message := recover(); message != nil {
				ret = true
			}
		}()
		f()
	}()
	return ret

}
