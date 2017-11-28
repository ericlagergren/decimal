//line number.rl:1
package parse

import "fmt"

type Special uint8

const (
	Invalid Special = iota
	QNaN
	SNaN
	Inf
)

func (s Special) String() string {
	switch s {
	case Invalid:
		return "invalid"
	case QNaN:
		return "qNaN"
	case SNaN:
		return "sNaN"
	case Inf:
		return "+inf"
	default:
		panic(fmt.Sprintf("unknown(%d)", s))
	}
}

func ParseSpecial(data string) (s Special, sign bool) {
	sign = len(data) > 0 && data[0] == '-'
	cs, p, pe := 0, 0, len(data)

//line number.go:32
	const parser_start int = 1
	const parser_first_final int = 15
	const parser_error int = 0

	const parser_en_main int = 1

//line number.go:40
	{
		cs = parser_start
	}

//line number.go:45
	{
		if p == pe {
			goto _test_eof
		}
		switch cs {
		case 1:
			goto st_case_1
		case 0:
			goto st_case_0
		case 2:
			goto st_case_2
		case 3:
			goto st_case_3
		case 4:
			goto st_case_4
		case 15:
			goto st_case_15
		case 5:
			goto st_case_5
		case 6:
			goto st_case_6
		case 7:
			goto st_case_7
		case 8:
			goto st_case_8
		case 16:
			goto st_case_16
		case 9:
			goto st_case_9
		case 10:
			goto st_case_10
		case 11:
			goto st_case_11
		case 12:
			goto st_case_12
		case 13:
			goto st_case_13
		case 14:
			goto st_case_14
		}
		goto st_out
	st_case_1:
		switch data[p] {
		case 43:
			goto st2
		case 45:
			goto st2
		case 73:
			goto st3
		case 78:
			goto st9
		case 81:
			goto st11
		case 83:
			goto st12
		case 105:
			goto st3
		case 110:
			goto st9
		case 113:
			goto st11
		case 115:
			goto st12
		}
		goto st0
	st_case_0:
	st0:
		cs = 0
		goto _out
	st2:
		if p++; p == pe {
			goto _test_eof2
		}
	st_case_2:
		switch data[p] {
		case 73:
			goto st3
		case 78:
			goto st9
		case 81:
			goto st11
		case 83:
			goto st12
		case 105:
			goto st3
		case 110:
			goto st9
		case 113:
			goto st11
		case 115:
			goto st12
		}
		goto st0
	st3:
		if p++; p == pe {
			goto _test_eof3
		}
	st_case_3:
		switch data[p] {
		case 78:
			goto st4
		case 110:
			goto st4
		}
		goto st0
	st4:
		if p++; p == pe {
			goto _test_eof4
		}
	st_case_4:
		switch data[p] {
		case 70:
			goto tr7
		case 102:
			goto tr7
		}
		goto st0
	tr7:
//line number.rl:36
		return Inf, sign
		goto st15
	st15:
		if p++; p == pe {
			goto _test_eof15
		}
	st_case_15:
//line number.go:172
		switch data[p] {
		case 73:
			goto st5
		case 105:
			goto st5
		}
		goto st0
	st5:
		if p++; p == pe {
			goto _test_eof5
		}
	st_case_5:
		switch data[p] {
		case 78:
			goto st6
		case 110:
			goto st6
		}
		goto st0
	st6:
		if p++; p == pe {
			goto _test_eof6
		}
	st_case_6:
		switch data[p] {
		case 73:
			goto st7
		case 105:
			goto st7
		}
		goto st0
	st7:
		if p++; p == pe {
			goto _test_eof7
		}
	st_case_7:
		switch data[p] {
		case 84:
			goto st8
		case 116:
			goto st8
		}
		goto st0
	st8:
		if p++; p == pe {
			goto _test_eof8
		}
	st_case_8:
		switch data[p] {
		case 89:
			goto tr11
		case 121:
			goto tr11
		}
		goto st0
	tr11:
//line number.rl:36
		return Inf, sign
		goto st16
	tr13:
//line number.rl:37
		return QNaN, sign
		goto st16
	tr16:
//line number.rl:38
		return SNaN, sign
		goto st16
	st16:
		if p++; p == pe {
			goto _test_eof16
		}
	st_case_16:
//line number.go:245
		goto st0
	st9:
		if p++; p == pe {
			goto _test_eof9
		}
	st_case_9:
		switch data[p] {
		case 65:
			goto st10
		case 97:
			goto st10
		}
		goto st0
	st10:
		if p++; p == pe {
			goto _test_eof10
		}
	st_case_10:
		switch data[p] {
		case 78:
			goto tr13
		case 110:
			goto tr13
		}
		goto st0
	st11:
		if p++; p == pe {
			goto _test_eof11
		}
	st_case_11:
		switch data[p] {
		case 78:
			goto st9
		case 110:
			goto st9
		}
		goto st0
	st12:
		if p++; p == pe {
			goto _test_eof12
		}
	st_case_12:
		switch data[p] {
		case 78:
			goto st13
		case 110:
			goto st13
		}
		goto st0
	st13:
		if p++; p == pe {
			goto _test_eof13
		}
	st_case_13:
		switch data[p] {
		case 65:
			goto st14
		case 97:
			goto st14
		}
		goto st0
	st14:
		if p++; p == pe {
			goto _test_eof14
		}
	st_case_14:
		switch data[p] {
		case 78:
			goto tr16
		case 110:
			goto tr16
		}
		goto st0
	st_out:
	_test_eof2:
		cs = 2
		goto _test_eof
	_test_eof3:
		cs = 3
		goto _test_eof
	_test_eof4:
		cs = 4
		goto _test_eof
	_test_eof15:
		cs = 15
		goto _test_eof
	_test_eof5:
		cs = 5
		goto _test_eof
	_test_eof6:
		cs = 6
		goto _test_eof
	_test_eof7:
		cs = 7
		goto _test_eof
	_test_eof8:
		cs = 8
		goto _test_eof
	_test_eof16:
		cs = 16
		goto _test_eof
	_test_eof9:
		cs = 9
		goto _test_eof
	_test_eof10:
		cs = 10
		goto _test_eof
	_test_eof11:
		cs = 11
		goto _test_eof
	_test_eof12:
		cs = 12
		goto _test_eof
	_test_eof13:
		cs = 13
		goto _test_eof
	_test_eof14:
		cs = 14
		goto _test_eof

	_test_eof:
		{
		}
	_out:
		{
		}
	}

//line number.rl:44
	return Invalid, sign
}
