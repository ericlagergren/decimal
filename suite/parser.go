
//line parser.rl:1
package suite

import (
    "fmt"
    "strconv"
)

func ParseCase(data []byte) (c Case, err error) {
    cs, p, pe, eof := 0, 0, len(data), len(data)

    var (
        ok   bool // for mode and op
        mark int
        exc  Exception
    )

    
//line parser.go:21
const parser_start int = 1
const parser_first_final int = 81
const parser_error int = 0

const parser_en_main int = 1


//line parser.go:29
	{
	cs = parser_start
	}

//line parser.go:34
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
	case 5:
		goto st_case_5
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 8:
		goto st_case_8
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
	case 81:
		goto st_case_81
	case 14:
		goto st_case_14
	case 82:
		goto st_case_82
	case 15:
		goto st_case_15
	case 83:
		goto st_case_83
	case 16:
		goto st_case_16
	case 84:
		goto st_case_84
	case 17:
		goto st_case_17
	case 85:
		goto st_case_85
	case 18:
		goto st_case_18
	case 19:
		goto st_case_19
	case 20:
		goto st_case_20
	case 21:
		goto st_case_21
	case 22:
		goto st_case_22
	case 23:
		goto st_case_23
	case 24:
		goto st_case_24
	case 86:
		goto st_case_86
	case 25:
		goto st_case_25
	case 26:
		goto st_case_26
	case 27:
		goto st_case_27
	case 28:
		goto st_case_28
	case 29:
		goto st_case_29
	case 30:
		goto st_case_30
	case 87:
		goto st_case_87
	case 31:
		goto st_case_31
	case 32:
		goto st_case_32
	case 33:
		goto st_case_33
	case 34:
		goto st_case_34
	case 35:
		goto st_case_35
	case 36:
		goto st_case_36
	case 37:
		goto st_case_37
	case 38:
		goto st_case_38
	case 39:
		goto st_case_39
	case 40:
		goto st_case_40
	case 41:
		goto st_case_41
	case 42:
		goto st_case_42
	case 43:
		goto st_case_43
	case 44:
		goto st_case_44
	case 45:
		goto st_case_45
	case 46:
		goto st_case_46
	case 47:
		goto st_case_47
	case 48:
		goto st_case_48
	case 49:
		goto st_case_49
	case 50:
		goto st_case_50
	case 51:
		goto st_case_51
	case 52:
		goto st_case_52
	case 53:
		goto st_case_53
	case 54:
		goto st_case_54
	case 55:
		goto st_case_55
	case 56:
		goto st_case_56
	case 57:
		goto st_case_57
	case 58:
		goto st_case_58
	case 59:
		goto st_case_59
	case 60:
		goto st_case_60
	case 61:
		goto st_case_61
	case 62:
		goto st_case_62
	case 63:
		goto st_case_63
	case 64:
		goto st_case_64
	case 65:
		goto st_case_65
	case 66:
		goto st_case_66
	case 67:
		goto st_case_67
	case 68:
		goto st_case_68
	case 69:
		goto st_case_69
	case 70:
		goto st_case_70
	case 71:
		goto st_case_71
	case 72:
		goto st_case_72
	case 73:
		goto st_case_73
	case 74:
		goto st_case_74
	case 75:
		goto st_case_75
	case 76:
		goto st_case_76
	case 77:
		goto st_case_77
	case 78:
		goto st_case_78
	case 79:
		goto st_case_79
	case 80:
		goto st_case_80
	}
	goto st_out
	st_case_1:
		switch data[p] {
		case 98:
			goto tr0
		case 100:
			goto tr0
		}
		goto st0
st_case_0:
	st0:
		cs = 0
		goto _out
tr0:
//line parser.rl:20
 mark = p 
	goto st2
	st2:
		if p++; p == pe {
			goto _test_eof2
		}
	st_case_2:
//line parser.go:239
		if 48 <= data[p] && data[p] <= 57 {
			goto tr2
		}
		goto st0
tr2:
//line parser.rl:21
 c.Prefix = string(data[mark:p]) 
//line parser.rl:20
 mark = p 
	goto st3
	st3:
		if p++; p == pe {
			goto _test_eof3
		}
	st_case_3:
//line parser.go:255
		switch data[p] {
		case 37:
			goto tr3
		case 42:
			goto tr4
		case 43:
			goto tr3
		case 45:
			goto tr3
		case 47:
			goto tr3
		case 61:
			goto tr7
		case 63:
			goto tr8
		case 76:
			goto tr3
		case 78:
			goto tr9
		case 83:
			goto tr3
		case 86:
			goto tr3
		case 99:
			goto tr10
		case 101:
			goto tr11
		case 113:
			goto tr12
		case 114:
			goto tr13
		case 115:
			goto tr14
		case 126:
			goto tr3
		}
		switch {
		case data[p] < 60:
			if 48 <= data[p] && data[p] <= 57 {
				goto st3
			}
		case data[p] > 62:
			if 64 <= data[p] && data[p] <= 65 {
				goto tr3
			}
		default:
			goto tr6
		}
		goto st0
tr3:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st4
	st4:
		if p++; p == pe {
			goto _test_eof4
		}
	st_case_4:
//line parser.go:320
		if data[p] == 32 {
			goto tr15
		}
		goto st0
tr15:
//line parser.rl:27

            if c.Op, ok = valToOp[string(data[mark:p])]; !ok {
                return c, fmt.Errorf("invalid op: %q", data[mark:p])
            }
        
	goto st5
	st5:
		if p++; p == pe {
			goto _test_eof5
		}
	st_case_5:
//line parser.go:338
		switch data[p] {
		case 48:
			goto tr16
		case 61:
			goto tr17
		case 94:
			goto tr16
		}
		if 60 <= data[p] && data[p] <= 62 {
			goto tr16
		}
		goto st0
tr16:
//line parser.rl:20
 mark = p 
	goto st6
	st6:
		if p++; p == pe {
			goto _test_eof6
		}
	st_case_6:
//line parser.go:360
		if data[p] == 32 {
			goto tr18
		}
		goto st0
tr18:
//line parser.rl:32

	    	if c.Mode, ok = valToMode[string(data[mark:p])]; !ok {
				return c, fmt.Errorf("invalid mode: %q", data[mark:p])
	    	}
        
	goto st7
	st7:
		if p++; p == pe {
			goto _test_eof7
		}
	st_case_7:
//line parser.go:378
		switch data[p] {
		case 43:
			goto tr19
		case 45:
			goto tr19
		case 70:
			goto tr21
		case 73:
			goto tr22
		case 78:
			goto tr23
		case 81:
			goto tr24
		case 83:
			goto tr24
		case 84:
			goto tr25
		case 102:
			goto tr21
		case 105:
			goto tr26
		case 110:
			goto tr23
		case 111:
			goto tr27
		case 113:
			goto tr24
		case 115:
			goto tr24
		case 116:
			goto tr25
		case 122:
			goto tr27
		}
		switch {
		case data[p] > 57:
			if 117 <= data[p] && data[p] <= 120 {
				goto tr27
			}
		case data[p] >= 48:
			goto tr20
		}
		goto st0
tr19:
//line parser.rl:20
 mark = p 
	goto st8
	st8:
		if p++; p == pe {
			goto _test_eof8
		}
	st_case_8:
//line parser.go:431
		switch data[p] {
		case 70:
			goto st33
		case 73:
			goto st38
		case 78:
			goto st45
		case 81:
			goto st47
		case 83:
			goto st47
		case 84:
			goto st48
		case 102:
			goto st33
		case 105:
			goto st38
		case 110:
			goto st45
		case 113:
			goto st47
		case 115:
			goto st47
		case 116:
			goto st48
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st9
		}
		goto st0
tr20:
//line parser.rl:20
 mark = p 
	goto st9
	st9:
		if p++; p == pe {
			goto _test_eof9
		}
	st_case_9:
//line parser.go:471
		switch data[p] {
		case 32:
			goto tr34
		case 43:
			goto st50
		case 45:
			goto st50
		case 46:
			goto st52
		case 69:
			goto st54
		case 101:
			goto st54
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st9
		}
		goto st0
tr34:
//line parser.rl:44
 c.Inputs = append(c.Inputs, Data(data[mark:p])) 
	goto st10
	st10:
		if p++; p == pe {
			goto _test_eof10
		}
	st_case_10:
//line parser.go:499
		switch data[p] {
		case 43:
			goto tr19
		case 45:
			goto tr38
		case 70:
			goto tr21
		case 73:
			goto tr22
		case 78:
			goto tr23
		case 81:
			goto tr24
		case 83:
			goto tr24
		case 84:
			goto tr25
		case 102:
			goto tr21
		case 105:
			goto tr22
		case 110:
			goto tr23
		case 113:
			goto tr24
		case 115:
			goto tr24
		case 116:
			goto tr25
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto tr20
		}
		goto st0
tr38:
//line parser.rl:20
 mark = p 
	goto st11
	st11:
		if p++; p == pe {
			goto _test_eof11
		}
	st_case_11:
//line parser.go:543
		switch data[p] {
		case 62:
			goto st12
		case 70:
			goto st33
		case 73:
			goto st38
		case 78:
			goto st45
		case 81:
			goto st47
		case 83:
			goto st47
		case 84:
			goto st48
		case 102:
			goto st33
		case 105:
			goto st38
		case 110:
			goto st45
		case 113:
			goto st47
		case 115:
			goto st47
		case 116:
			goto st48
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st9
		}
		goto st0
	st12:
		if p++; p == pe {
			goto _test_eof12
		}
	st_case_12:
		if data[p] == 32 {
			goto st13
		}
		goto st0
	st13:
		if p++; p == pe {
			goto _test_eof13
		}
	st_case_13:
		switch data[p] {
		case 35:
			goto tr41
		case 43:
			goto tr42
		case 45:
			goto tr42
		case 70:
			goto tr44
		case 73:
			goto tr45
		case 78:
			goto tr46
		case 81:
			goto tr47
		case 83:
			goto tr47
		case 84:
			goto tr48
		case 102:
			goto tr44
		case 105:
			goto tr45
		case 110:
			goto tr46
		case 113:
			goto tr47
		case 115:
			goto tr47
		case 116:
			goto tr48
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto tr43
		}
		goto st0
tr41:
//line parser.rl:20
 mark = p 
	goto st81
	st81:
		if p++; p == pe {
			goto _test_eof81
		}
	st_case_81:
//line parser.go:635
		if data[p] == 32 {
			goto tr100
		}
		goto st0
tr100:
//line parser.rl:45
 c.Output = Data(data[mark:p]) 
	goto st14
	st14:
		if p++; p == pe {
			goto _test_eof14
		}
	st_case_14:
//line parser.go:649
		switch data[p] {
		case 105:
			goto tr49
		case 111:
			goto tr49
		case 122:
			goto tr49
		}
		if 117 <= data[p] && data[p] <= 120 {
			goto tr49
		}
		goto st0
tr49:
//line parser.rl:20
 mark = p 
	goto st82
tr101:
//line parser.rl:46

	    	if exc, ok = valToException[string(data[mark:p])]; !ok {
				return c, fmt.Errorf("invalid result exception: %q", data[mark:p])
	    	}
            c.Excep |= exc
        
//line parser.rl:20
 mark = p 
	goto st82
	st82:
		if p++; p == pe {
			goto _test_eof82
		}
	st_case_82:
//line parser.go:682
		switch data[p] {
		case 105:
			goto tr101
		case 111:
			goto tr101
		case 122:
			goto tr101
		}
		if 117 <= data[p] && data[p] <= 120 {
			goto tr101
		}
		goto st0
tr42:
//line parser.rl:20
 mark = p 
	goto st15
	st15:
		if p++; p == pe {
			goto _test_eof15
		}
	st_case_15:
//line parser.go:704
		switch data[p] {
		case 70:
			goto st19
		case 73:
			goto st23
		case 78:
			goto st29
		case 81:
			goto st87
		case 83:
			goto st87
		case 84:
			goto st31
		case 102:
			goto st19
		case 105:
			goto st23
		case 110:
			goto st29
		case 113:
			goto st87
		case 115:
			goto st87
		case 116:
			goto st31
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st83
		}
		goto st0
tr43:
//line parser.rl:20
 mark = p 
	goto st83
	st83:
		if p++; p == pe {
			goto _test_eof83
		}
	st_case_83:
//line parser.go:744
		switch data[p] {
		case 32:
			goto tr100
		case 43:
			goto st16
		case 45:
			goto st16
		case 46:
			goto st17
		case 69:
			goto st18
		case 101:
			goto st18
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st83
		}
		goto st0
	st16:
		if p++; p == pe {
			goto _test_eof16
		}
	st_case_16:
		if 48 <= data[p] && data[p] <= 57 {
			goto st84
		}
		goto st0
	st84:
		if p++; p == pe {
			goto _test_eof84
		}
	st_case_84:
		if data[p] == 32 {
			goto tr100
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st84
		}
		goto st0
	st17:
		if p++; p == pe {
			goto _test_eof17
		}
	st_case_17:
		if 48 <= data[p] && data[p] <= 57 {
			goto st85
		}
		goto st0
	st85:
		if p++; p == pe {
			goto _test_eof85
		}
	st_case_85:
		switch data[p] {
		case 32:
			goto tr100
		case 43:
			goto st16
		case 45:
			goto st16
		case 69:
			goto st18
		case 101:
			goto st18
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st85
		}
		goto st0
	st18:
		if p++; p == pe {
			goto _test_eof18
		}
	st_case_18:
		switch data[p] {
		case 43:
			goto st16
		case 45:
			goto st16
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st84
		}
		goto st0
tr44:
//line parser.rl:20
 mark = p 
	goto st19
	st19:
		if p++; p == pe {
			goto _test_eof19
		}
	st_case_19:
//line parser.go:838
		switch data[p] {
		case 65:
			goto st20
		case 97:
			goto st20
		}
		goto st0
	st20:
		if p++; p == pe {
			goto _test_eof20
		}
	st_case_20:
		switch data[p] {
		case 76:
			goto st21
		case 108:
			goto st21
		}
		goto st0
	st21:
		if p++; p == pe {
			goto _test_eof21
		}
	st_case_21:
		switch data[p] {
		case 83:
			goto st22
		case 115:
			goto st22
		}
		goto st0
	st22:
		if p++; p == pe {
			goto _test_eof22
		}
	st_case_22:
		switch data[p] {
		case 69:
			goto st81
		case 101:
			goto st81
		}
		goto st0
tr45:
//line parser.rl:20
 mark = p 
	goto st23
	st23:
		if p++; p == pe {
			goto _test_eof23
		}
	st_case_23:
//line parser.go:891
		switch data[p] {
		case 78:
			goto st24
		case 110:
			goto st24
		}
		goto st0
	st24:
		if p++; p == pe {
			goto _test_eof24
		}
	st_case_24:
		switch data[p] {
		case 70:
			goto st86
		case 102:
			goto st86
		}
		goto st0
	st86:
		if p++; p == pe {
			goto _test_eof86
		}
	st_case_86:
		switch data[p] {
		case 32:
			goto tr100
		case 73:
			goto st25
		case 105:
			goto st25
		}
		goto st0
	st25:
		if p++; p == pe {
			goto _test_eof25
		}
	st_case_25:
		switch data[p] {
		case 78:
			goto st26
		case 110:
			goto st26
		}
		goto st0
	st26:
		if p++; p == pe {
			goto _test_eof26
		}
	st_case_26:
		switch data[p] {
		case 73:
			goto st27
		case 105:
			goto st27
		}
		goto st0
	st27:
		if p++; p == pe {
			goto _test_eof27
		}
	st_case_27:
		switch data[p] {
		case 84:
			goto st28
		case 116:
			goto st28
		}
		goto st0
	st28:
		if p++; p == pe {
			goto _test_eof28
		}
	st_case_28:
		switch data[p] {
		case 89:
			goto st81
		case 121:
			goto st81
		}
		goto st0
tr46:
//line parser.rl:20
 mark = p 
	goto st29
	st29:
		if p++; p == pe {
			goto _test_eof29
		}
	st_case_29:
//line parser.go:982
		switch data[p] {
		case 65:
			goto st30
		case 97:
			goto st30
		}
		goto st0
	st30:
		if p++; p == pe {
			goto _test_eof30
		}
	st_case_30:
		switch data[p] {
		case 78:
			goto st81
		case 110:
			goto st81
		}
		goto st0
tr47:
//line parser.rl:20
 mark = p 
	goto st87
	st87:
		if p++; p == pe {
			goto _test_eof87
		}
	st_case_87:
//line parser.go:1011
		switch data[p] {
		case 32:
			goto tr100
		case 78:
			goto st29
		case 110:
			goto st29
		}
		goto st0
tr48:
//line parser.rl:20
 mark = p 
	goto st31
	st31:
		if p++; p == pe {
			goto _test_eof31
		}
	st_case_31:
//line parser.go:1030
		switch data[p] {
		case 82:
			goto st32
		case 114:
			goto st32
		}
		goto st0
	st32:
		if p++; p == pe {
			goto _test_eof32
		}
	st_case_32:
		switch data[p] {
		case 85:
			goto st22
		case 117:
			goto st22
		}
		goto st0
tr21:
//line parser.rl:20
 mark = p 
	goto st33
	st33:
		if p++; p == pe {
			goto _test_eof33
		}
	st_case_33:
//line parser.go:1059
		switch data[p] {
		case 65:
			goto st34
		case 97:
			goto st34
		}
		goto st0
	st34:
		if p++; p == pe {
			goto _test_eof34
		}
	st_case_34:
		switch data[p] {
		case 76:
			goto st35
		case 108:
			goto st35
		}
		goto st0
	st35:
		if p++; p == pe {
			goto _test_eof35
		}
	st_case_35:
		switch data[p] {
		case 83:
			goto st36
		case 115:
			goto st36
		}
		goto st0
	st36:
		if p++; p == pe {
			goto _test_eof36
		}
	st_case_36:
		switch data[p] {
		case 69:
			goto st37
		case 101:
			goto st37
		}
		goto st0
	st37:
		if p++; p == pe {
			goto _test_eof37
		}
	st_case_37:
		if data[p] == 32 {
			goto tr34
		}
		goto st0
tr22:
//line parser.rl:20
 mark = p 
	goto st38
	st38:
		if p++; p == pe {
			goto _test_eof38
		}
	st_case_38:
//line parser.go:1121
		switch data[p] {
		case 78:
			goto st39
		case 110:
			goto st39
		}
		goto st0
	st39:
		if p++; p == pe {
			goto _test_eof39
		}
	st_case_39:
		switch data[p] {
		case 70:
			goto st40
		case 102:
			goto st40
		}
		goto st0
	st40:
		if p++; p == pe {
			goto _test_eof40
		}
	st_case_40:
		switch data[p] {
		case 32:
			goto tr34
		case 73:
			goto st41
		case 105:
			goto st41
		}
		goto st0
	st41:
		if p++; p == pe {
			goto _test_eof41
		}
	st_case_41:
		switch data[p] {
		case 78:
			goto st42
		case 110:
			goto st42
		}
		goto st0
	st42:
		if p++; p == pe {
			goto _test_eof42
		}
	st_case_42:
		switch data[p] {
		case 73:
			goto st43
		case 105:
			goto st43
		}
		goto st0
	st43:
		if p++; p == pe {
			goto _test_eof43
		}
	st_case_43:
		switch data[p] {
		case 84:
			goto st44
		case 116:
			goto st44
		}
		goto st0
	st44:
		if p++; p == pe {
			goto _test_eof44
		}
	st_case_44:
		switch data[p] {
		case 89:
			goto st37
		case 121:
			goto st37
		}
		goto st0
tr23:
//line parser.rl:20
 mark = p 
	goto st45
	st45:
		if p++; p == pe {
			goto _test_eof45
		}
	st_case_45:
//line parser.go:1212
		switch data[p] {
		case 65:
			goto st46
		case 97:
			goto st46
		}
		goto st0
	st46:
		if p++; p == pe {
			goto _test_eof46
		}
	st_case_46:
		switch data[p] {
		case 78:
			goto st37
		case 110:
			goto st37
		}
		goto st0
tr24:
//line parser.rl:20
 mark = p 
	goto st47
	st47:
		if p++; p == pe {
			goto _test_eof47
		}
	st_case_47:
//line parser.go:1241
		switch data[p] {
		case 32:
			goto tr34
		case 78:
			goto st45
		case 110:
			goto st45
		}
		goto st0
tr25:
//line parser.rl:20
 mark = p 
	goto st48
	st48:
		if p++; p == pe {
			goto _test_eof48
		}
	st_case_48:
//line parser.go:1260
		switch data[p] {
		case 82:
			goto st49
		case 114:
			goto st49
		}
		goto st0
	st49:
		if p++; p == pe {
			goto _test_eof49
		}
	st_case_49:
		switch data[p] {
		case 85:
			goto st36
		case 117:
			goto st36
		}
		goto st0
	st50:
		if p++; p == pe {
			goto _test_eof50
		}
	st_case_50:
		if 48 <= data[p] && data[p] <= 57 {
			goto st51
		}
		goto st0
	st51:
		if p++; p == pe {
			goto _test_eof51
		}
	st_case_51:
		if data[p] == 32 {
			goto tr34
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st51
		}
		goto st0
	st52:
		if p++; p == pe {
			goto _test_eof52
		}
	st_case_52:
		if 48 <= data[p] && data[p] <= 57 {
			goto st53
		}
		goto st0
	st53:
		if p++; p == pe {
			goto _test_eof53
		}
	st_case_53:
		switch data[p] {
		case 32:
			goto tr34
		case 43:
			goto st50
		case 45:
			goto st50
		case 69:
			goto st54
		case 101:
			goto st54
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st53
		}
		goto st0
	st54:
		if p++; p == pe {
			goto _test_eof54
		}
	st_case_54:
		switch data[p] {
		case 43:
			goto st50
		case 45:
			goto st50
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st51
		}
		goto st0
tr26:
//line parser.rl:20
 mark = p 
	goto st55
	st55:
		if p++; p == pe {
			goto _test_eof55
		}
	st_case_55:
//line parser.go:1355
		switch data[p] {
		case 32:
			goto tr84
		case 78:
			goto st39
		case 105:
			goto tr85
		case 110:
			goto st39
		case 111:
			goto tr85
		case 122:
			goto tr85
		}
		if 117 <= data[p] && data[p] <= 120 {
			goto tr85
		}
		goto st0
tr84:
//line parser.rl:37

	    exc, ok = valToException[string(data[mark:p])]
	    if !ok {
                return c, fmt.Errorf("invalid trap exception: %q", data[mark:p])
	    }
            c.Trap |= exc
        
	goto st56
	st56:
		if p++; p == pe {
			goto _test_eof56
		}
	st_case_56:
//line parser.go:1389
		switch data[p] {
		case 43:
			goto tr19
		case 45:
			goto tr19
		case 70:
			goto tr21
		case 73:
			goto tr22
		case 78:
			goto tr23
		case 81:
			goto tr24
		case 83:
			goto tr24
		case 84:
			goto tr25
		case 102:
			goto tr21
		case 105:
			goto tr22
		case 110:
			goto tr23
		case 113:
			goto tr24
		case 115:
			goto tr24
		case 116:
			goto tr25
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto tr20
		}
		goto st0
tr27:
//line parser.rl:20
 mark = p 
	goto st57
tr85:
//line parser.rl:37

	    exc, ok = valToException[string(data[mark:p])]
	    if !ok {
                return c, fmt.Errorf("invalid trap exception: %q", data[mark:p])
	    }
            c.Trap |= exc
        
//line parser.rl:20
 mark = p 
	goto st57
	st57:
		if p++; p == pe {
			goto _test_eof57
		}
	st_case_57:
//line parser.go:1445
		switch data[p] {
		case 32:
			goto tr84
		case 105:
			goto tr85
		case 111:
			goto tr85
		case 122:
			goto tr85
		}
		if 117 <= data[p] && data[p] <= 120 {
			goto tr85
		}
		goto st0
tr17:
//line parser.rl:20
 mark = p 
	goto st58
	st58:
		if p++; p == pe {
			goto _test_eof58
		}
	st_case_58:
//line parser.go:1469
		switch data[p] {
		case 48:
			goto st6
		case 94:
			goto st6
		}
		goto st0
tr4:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st59
	st59:
		if p++; p == pe {
			goto _test_eof59
		}
	st_case_59:
//line parser.go:1492
		switch data[p] {
		case 32:
			goto tr15
		case 45:
			goto st4
		}
		goto st0
tr6:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st60
	st60:
		if p++; p == pe {
			goto _test_eof60
		}
	st_case_60:
//line parser.go:1515
		switch data[p] {
		case 65:
			goto st4
		case 67:
			goto st4
		}
		goto st0
tr7:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st61
	st61:
		if p++; p == pe {
			goto _test_eof61
		}
	st_case_61:
//line parser.go:1538
		if data[p] == 113 {
			goto st62
		}
		goto st0
	st62:
		if p++; p == pe {
			goto _test_eof62
		}
	st_case_62:
		if data[p] == 117 {
			goto st63
		}
		goto st0
	st63:
		if p++; p == pe {
			goto _test_eof63
		}
	st_case_63:
		if data[p] == 97 {
			goto st64
		}
		goto st0
	st64:
		if p++; p == pe {
			goto _test_eof64
		}
	st_case_64:
		if data[p] == 110 {
			goto st65
		}
		goto st0
	st65:
		if p++; p == pe {
			goto _test_eof65
		}
	st_case_65:
		if data[p] == 116 {
			goto st4
		}
		goto st0
tr8:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st66
	st66:
		if p++; p == pe {
			goto _test_eof66
		}
	st_case_66:
//line parser.go:1594
		switch data[p] {
		case 32:
			goto tr15
		case 45:
			goto st4
		case 48:
			goto st4
		case 78:
			goto st4
		case 102:
			goto st4
		case 105:
			goto st4
		case 110:
			goto st4
		case 115:
			goto st67
		}
		goto st0
	st67:
		if p++; p == pe {
			goto _test_eof67
		}
	st_case_67:
		switch data[p] {
		case 32:
			goto tr15
		case 78:
			goto st4
		}
		goto st0
tr9:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st68
	st68:
		if p++; p == pe {
			goto _test_eof68
		}
	st_case_68:
//line parser.go:1641
		switch data[p] {
		case 97:
			goto st4
		case 100:
			goto st4
		case 117:
			goto st4
		}
		goto st0
tr10:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st69
	st69:
		if p++; p == pe {
			goto _test_eof69
		}
	st_case_69:
//line parser.go:1666
		switch data[p] {
		case 100:
			goto st70
		case 102:
			goto st71
		case 105:
			goto st70
		case 112:
			goto st4
		}
		goto st0
	st70:
		if p++; p == pe {
			goto _test_eof70
		}
	st_case_70:
		if data[p] == 102 {
			goto st4
		}
		goto st0
	st71:
		if p++; p == pe {
			goto _test_eof71
		}
	st_case_71:
		switch data[p] {
		case 100:
			goto st4
		case 102:
			goto st4
		case 105:
			goto st4
		}
		goto st0
tr11:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st72
	st72:
		if p++; p == pe {
			goto _test_eof72
		}
	st_case_72:
//line parser.go:1716
		if data[p] == 113 {
			goto st4
		}
		goto st0
tr12:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st73
	st73:
		if p++; p == pe {
			goto _test_eof73
		}
	st_case_73:
//line parser.go:1736
		switch data[p] {
		case 67:
			goto st4
		case 117:
			goto st63
		}
		goto st0
tr13:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st74
	st74:
		if p++; p == pe {
			goto _test_eof74
		}
	st_case_74:
//line parser.go:1759
		switch data[p] {
		case 97:
			goto st65
		case 102:
			goto st75
		}
		goto st0
	st75:
		if p++; p == pe {
			goto _test_eof75
		}
	st_case_75:
		if data[p] == 105 {
			goto st4
		}
		goto st0
tr14:
//line parser.rl:22

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:20
 mark = p 
	goto st76
	st76:
		if p++; p == pe {
			goto _test_eof76
		}
	st_case_76:
//line parser.go:1791
		switch data[p] {
		case 67:
			goto st4
		case 105:
			goto st77
		}
		goto st0
	st77:
		if p++; p == pe {
			goto _test_eof77
		}
	st_case_77:
		if data[p] == 103 {
			goto st78
		}
		goto st0
	st78:
		if p++; p == pe {
			goto _test_eof78
		}
	st_case_78:
		if data[p] == 110 {
			goto st79
		}
		goto st0
	st79:
		if p++; p == pe {
			goto _test_eof79
		}
	st_case_79:
		switch data[p] {
		case 32:
			goto tr15
		case 98:
			goto st80
		}
		goto st0
	st80:
		if p++; p == pe {
			goto _test_eof80
		}
	st_case_80:
		if data[p] == 105 {
			goto st65
		}
		goto st0
	st_out:
	_test_eof2: cs = 2; goto _test_eof
	_test_eof3: cs = 3; goto _test_eof
	_test_eof4: cs = 4; goto _test_eof
	_test_eof5: cs = 5; goto _test_eof
	_test_eof6: cs = 6; goto _test_eof
	_test_eof7: cs = 7; goto _test_eof
	_test_eof8: cs = 8; goto _test_eof
	_test_eof9: cs = 9; goto _test_eof
	_test_eof10: cs = 10; goto _test_eof
	_test_eof11: cs = 11; goto _test_eof
	_test_eof12: cs = 12; goto _test_eof
	_test_eof13: cs = 13; goto _test_eof
	_test_eof81: cs = 81; goto _test_eof
	_test_eof14: cs = 14; goto _test_eof
	_test_eof82: cs = 82; goto _test_eof
	_test_eof15: cs = 15; goto _test_eof
	_test_eof83: cs = 83; goto _test_eof
	_test_eof16: cs = 16; goto _test_eof
	_test_eof84: cs = 84; goto _test_eof
	_test_eof17: cs = 17; goto _test_eof
	_test_eof85: cs = 85; goto _test_eof
	_test_eof18: cs = 18; goto _test_eof
	_test_eof19: cs = 19; goto _test_eof
	_test_eof20: cs = 20; goto _test_eof
	_test_eof21: cs = 21; goto _test_eof
	_test_eof22: cs = 22; goto _test_eof
	_test_eof23: cs = 23; goto _test_eof
	_test_eof24: cs = 24; goto _test_eof
	_test_eof86: cs = 86; goto _test_eof
	_test_eof25: cs = 25; goto _test_eof
	_test_eof26: cs = 26; goto _test_eof
	_test_eof27: cs = 27; goto _test_eof
	_test_eof28: cs = 28; goto _test_eof
	_test_eof29: cs = 29; goto _test_eof
	_test_eof30: cs = 30; goto _test_eof
	_test_eof87: cs = 87; goto _test_eof
	_test_eof31: cs = 31; goto _test_eof
	_test_eof32: cs = 32; goto _test_eof
	_test_eof33: cs = 33; goto _test_eof
	_test_eof34: cs = 34; goto _test_eof
	_test_eof35: cs = 35; goto _test_eof
	_test_eof36: cs = 36; goto _test_eof
	_test_eof37: cs = 37; goto _test_eof
	_test_eof38: cs = 38; goto _test_eof
	_test_eof39: cs = 39; goto _test_eof
	_test_eof40: cs = 40; goto _test_eof
	_test_eof41: cs = 41; goto _test_eof
	_test_eof42: cs = 42; goto _test_eof
	_test_eof43: cs = 43; goto _test_eof
	_test_eof44: cs = 44; goto _test_eof
	_test_eof45: cs = 45; goto _test_eof
	_test_eof46: cs = 46; goto _test_eof
	_test_eof47: cs = 47; goto _test_eof
	_test_eof48: cs = 48; goto _test_eof
	_test_eof49: cs = 49; goto _test_eof
	_test_eof50: cs = 50; goto _test_eof
	_test_eof51: cs = 51; goto _test_eof
	_test_eof52: cs = 52; goto _test_eof
	_test_eof53: cs = 53; goto _test_eof
	_test_eof54: cs = 54; goto _test_eof
	_test_eof55: cs = 55; goto _test_eof
	_test_eof56: cs = 56; goto _test_eof
	_test_eof57: cs = 57; goto _test_eof
	_test_eof58: cs = 58; goto _test_eof
	_test_eof59: cs = 59; goto _test_eof
	_test_eof60: cs = 60; goto _test_eof
	_test_eof61: cs = 61; goto _test_eof
	_test_eof62: cs = 62; goto _test_eof
	_test_eof63: cs = 63; goto _test_eof
	_test_eof64: cs = 64; goto _test_eof
	_test_eof65: cs = 65; goto _test_eof
	_test_eof66: cs = 66; goto _test_eof
	_test_eof67: cs = 67; goto _test_eof
	_test_eof68: cs = 68; goto _test_eof
	_test_eof69: cs = 69; goto _test_eof
	_test_eof70: cs = 70; goto _test_eof
	_test_eof71: cs = 71; goto _test_eof
	_test_eof72: cs = 72; goto _test_eof
	_test_eof73: cs = 73; goto _test_eof
	_test_eof74: cs = 74; goto _test_eof
	_test_eof75: cs = 75; goto _test_eof
	_test_eof76: cs = 76; goto _test_eof
	_test_eof77: cs = 77; goto _test_eof
	_test_eof78: cs = 78; goto _test_eof
	_test_eof79: cs = 79; goto _test_eof
	_test_eof80: cs = 80; goto _test_eof

	_test_eof: {}
	if p == eof {
		switch cs {
		case 81, 83, 84, 85, 86, 87:
//line parser.rl:45
 c.Output = Data(data[mark:p]) 
		case 82:
//line parser.rl:46

	    	if exc, ok = valToException[string(data[mark:p])]; !ok {
				return c, fmt.Errorf("invalid result exception: %q", data[mark:p])
	    	}
            c.Excep |= exc
        
//line parser.go:1940
		}
	}

	_out: {}
	}

//line parser.rl:149

    return c, nil
}
