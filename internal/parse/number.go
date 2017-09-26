
//line number.rl:1
package parse

import "fmt"

type Special uint8

const (
	Invalid Special = iota
	QNaN 
	SNaN 
	PInf
	NInf
)

func (s Special) String() string {
	switch s {
	case Invalid: return "invalid"
	case QNaN:    return "qNaN"
	case SNaN:    return "sNaN"
	case PInf:    return "+inf"
	case NInf:    return "-inf"
	default:      panic(fmt.Sprintf("unknown(%d)", s))
	}
}

func ParseSpecial(data string) Special {
	cs, p, pe := 0, 0, len(data)

	
//line number.go:33
const parser_start int = 1
const parser_first_final int = 22
const parser_error int = 0

const parser_en_main int = 1


//line number.go:41
	{
	cs = parser_start
	}

//line number.go:46
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
	case 22:
		goto st_case_22
	case 5:
		goto st_case_5
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 8:
		goto st_case_8
	case 23:
		goto st_case_23
	case 9:
		goto st_case_9
	case 10:
		goto st_case_10
	case 11:
		goto st_case_11
	case 24:
		goto st_case_24
	case 12:
		goto st_case_12
	case 13:
		goto st_case_13
	case 14:
		goto st_case_14
	case 15:
		goto st_case_15
	case 16:
		goto st_case_16
	case 17:
		goto st_case_17
	case 18:
		goto st_case_18
	case 19:
		goto st_case_19
	case 20:
		goto st_case_20
	case 21:
		goto st_case_21
	}
	goto st_out
	st_case_1:
		switch data[p] {
		case 43:
			goto st2
		case 45:
			goto st9
		case 73:
			goto st3
		case 78:
			goto st16
		case 81:
			goto st18
		case 83:
			goto st19
		case 105:
			goto st3
		case 110:
			goto st16
		case 113:
			goto st18
		case 115:
			goto st19
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
		case 105:
			goto st3
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
			goto tr8
		case 102:
			goto tr8
		}
		goto st0
tr8:
//line number.rl:36
 return PInf 
	goto st22
	st22:
		if p++; p == pe {
			goto _test_eof22
		}
	st_case_22:
//line number.go:177
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
			goto tr12
		case 121:
			goto tr12
		}
		goto st0
tr12:
//line number.rl:36
 return PInf 
	goto st23
tr19:
//line number.rl:37
 return NInf 
	goto st23
tr21:
//line number.rl:38
 return QNaN 
	goto st23
tr24:
//line number.rl:39
 return SNaN 
	goto st23
	st23:
		if p++; p == pe {
			goto _test_eof23
		}
	st_case_23:
//line number.go:254
		goto st0
	st9:
		if p++; p == pe {
			goto _test_eof9
		}
	st_case_9:
		switch data[p] {
		case 73:
			goto st10
		case 105:
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
			goto st11
		case 110:
			goto st11
		}
		goto st0
	st11:
		if p++; p == pe {
			goto _test_eof11
		}
	st_case_11:
		switch data[p] {
		case 70:
			goto tr15
		case 102:
			goto tr15
		}
		goto st0
tr15:
//line number.rl:37
 return NInf 
	goto st24
	st24:
		if p++; p == pe {
			goto _test_eof24
		}
	st_case_24:
//line number.go:301
		switch data[p] {
		case 73:
			goto st12
		case 105:
			goto st12
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
		case 73:
			goto st14
		case 105:
			goto st14
		}
		goto st0
	st14:
		if p++; p == pe {
			goto _test_eof14
		}
	st_case_14:
		switch data[p] {
		case 84:
			goto st15
		case 116:
			goto st15
		}
		goto st0
	st15:
		if p++; p == pe {
			goto _test_eof15
		}
	st_case_15:
		switch data[p] {
		case 89:
			goto tr19
		case 121:
			goto tr19
		}
		goto st0
	st16:
		if p++; p == pe {
			goto _test_eof16
		}
	st_case_16:
		switch data[p] {
		case 65:
			goto st17
		case 97:
			goto st17
		}
		goto st0
	st17:
		if p++; p == pe {
			goto _test_eof17
		}
	st_case_17:
		switch data[p] {
		case 78:
			goto tr21
		case 110:
			goto tr21
		}
		goto st0
	st18:
		if p++; p == pe {
			goto _test_eof18
		}
	st_case_18:
		switch data[p] {
		case 78:
			goto st16
		case 110:
			goto st16
		}
		goto st0
	st19:
		if p++; p == pe {
			goto _test_eof19
		}
	st_case_19:
		switch data[p] {
		case 78:
			goto st20
		case 110:
			goto st20
		}
		goto st0
	st20:
		if p++; p == pe {
			goto _test_eof20
		}
	st_case_20:
		switch data[p] {
		case 65:
			goto st21
		case 97:
			goto st21
		}
		goto st0
	st21:
		if p++; p == pe {
			goto _test_eof21
		}
	st_case_21:
		switch data[p] {
		case 78:
			goto tr24
		case 110:
			goto tr24
		}
		goto st0
	st_out:
	_test_eof2: cs = 2; goto _test_eof
	_test_eof3: cs = 3; goto _test_eof
	_test_eof4: cs = 4; goto _test_eof
	_test_eof22: cs = 22; goto _test_eof
	_test_eof5: cs = 5; goto _test_eof
	_test_eof6: cs = 6; goto _test_eof
	_test_eof7: cs = 7; goto _test_eof
	_test_eof8: cs = 8; goto _test_eof
	_test_eof23: cs = 23; goto _test_eof
	_test_eof9: cs = 9; goto _test_eof
	_test_eof10: cs = 10; goto _test_eof
	_test_eof11: cs = 11; goto _test_eof
	_test_eof24: cs = 24; goto _test_eof
	_test_eof12: cs = 12; goto _test_eof
	_test_eof13: cs = 13; goto _test_eof
	_test_eof14: cs = 14; goto _test_eof
	_test_eof15: cs = 15; goto _test_eof
	_test_eof16: cs = 16; goto _test_eof
	_test_eof17: cs = 17; goto _test_eof
	_test_eof18: cs = 18; goto _test_eof
	_test_eof19: cs = 19; goto _test_eof
	_test_eof20: cs = 20; goto _test_eof
	_test_eof21: cs = 21; goto _test_eof

	_test_eof: {}
	_out: {}
	}

//line number.rl:45

	return Invalid
}
