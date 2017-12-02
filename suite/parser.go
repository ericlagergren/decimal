
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
    )

    
//line parser.go:20
const parser_start int = 1
const parser_first_final int = 118
const parser_error int = 0

const parser_en_main int = 1


//line parser.go:28
	{
	cs = parser_start
	}

//line parser.go:33
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
	case 22:
		goto st_case_22
	case 23:
		goto st_case_23
	case 24:
		goto st_case_24
	case 25:
		goto st_case_25
	case 26:
		goto st_case_26
	case 27:
		goto st_case_27
	case 28:
		goto st_case_28
	case 118:
		goto st_case_118
	case 119:
		goto st_case_119
	case 120:
		goto st_case_120
	case 29:
		goto st_case_29
	case 30:
		goto st_case_30
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
	case 121:
		goto st_case_121
	case 39:
		goto st_case_39
	case 122:
		goto st_case_122
	case 40:
		goto st_case_40
	case 123:
		goto st_case_123
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
	case 124:
		goto st_case_124
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
	case 125:
		goto st_case_125
	case 54:
		goto st_case_54
	case 126:
		goto st_case_126
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
	case 81:
		goto st_case_81
	case 82:
		goto st_case_82
	case 83:
		goto st_case_83
	case 84:
		goto st_case_84
	case 85:
		goto st_case_85
	case 86:
		goto st_case_86
	case 87:
		goto st_case_87
	case 88:
		goto st_case_88
	case 89:
		goto st_case_89
	case 90:
		goto st_case_90
	case 91:
		goto st_case_91
	case 92:
		goto st_case_92
	case 93:
		goto st_case_93
	case 94:
		goto st_case_94
	case 95:
		goto st_case_95
	case 96:
		goto st_case_96
	case 97:
		goto st_case_97
	case 98:
		goto st_case_98
	case 99:
		goto st_case_99
	case 100:
		goto st_case_100
	case 101:
		goto st_case_101
	case 102:
		goto st_case_102
	case 103:
		goto st_case_103
	case 104:
		goto st_case_104
	case 105:
		goto st_case_105
	case 106:
		goto st_case_106
	case 107:
		goto st_case_107
	case 108:
		goto st_case_108
	case 109:
		goto st_case_109
	case 110:
		goto st_case_110
	case 111:
		goto st_case_111
	case 112:
		goto st_case_112
	case 113:
		goto st_case_113
	case 114:
		goto st_case_114
	case 115:
		goto st_case_115
	case 116:
		goto st_case_116
	case 117:
		goto st_case_117
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
//line parser.rl:19
 mark = p 
	goto st2
	st2:
		if p++; p == pe {
			goto _test_eof2
		}
	st_case_2:
//line parser.go:316
		if 48 <= data[p] && data[p] <= 57 {
			goto tr2
		}
		goto st0
tr2:
//line parser.rl:20
 c.Prefix = string(data[mark:p]) 
//line parser.rl:19
 mark = p 
	goto st3
	st3:
		if p++; p == pe {
			goto _test_eof3
		}
	st_case_3:
//line parser.go:332
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
		case 108:
			goto tr12
		case 113:
			goto tr13
		case 114:
			goto tr14
		case 115:
			goto tr15
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
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st4
	st4:
		if p++; p == pe {
			goto _test_eof4
		}
	st_case_4:
//line parser.go:399
		if data[p] == 32 {
			goto tr16
		}
		goto st0
tr16:
//line parser.rl:26

            if c.Op, ok = valToOp[string(data[mark:p])]; !ok {
                return c, fmt.Errorf("invalid op: %q", data[mark:p])
            }
        
	goto st5
	st5:
		if p++; p == pe {
			goto _test_eof5
		}
	st_case_5:
//line parser.go:417
		switch data[p] {
		case 32:
			goto st5
		case 48:
			goto tr18
		case 61:
			goto tr19
		case 94:
			goto tr18
		}
		if 60 <= data[p] && data[p] <= 62 {
			goto tr18
		}
		goto st0
tr18:
//line parser.rl:19
 mark = p 
	goto st6
	st6:
		if p++; p == pe {
			goto _test_eof6
		}
	st_case_6:
//line parser.go:441
		if data[p] == 32 {
			goto tr20
		}
		goto st0
tr20:
//line parser.rl:31

	    	if c.Mode, ok = valToMode[string(data[mark:p])]; !ok {
				return c, fmt.Errorf("invalid mode: %q", data[mark:p])
	    	}
        
	goto st7
tr21:
//line parser.rl:19
 mark = p 
//line parser.rl:36

			c.Trap = ConditionFromString(string(data[mark:p]))
        
	goto st7
	st7:
		if p++; p == pe {
			goto _test_eof7
		}
	st_case_7:
//line parser.go:467
		switch data[p] {
		case 32:
			goto tr21
		case 43:
			goto tr22
		case 45:
			goto tr22
		case 63:
			goto tr24
		case 70:
			goto tr25
		case 73:
			goto tr26
		case 78:
			goto tr27
		case 81:
			goto tr28
		case 83:
			goto tr29
		case 84:
			goto tr30
		case 90:
			goto tr31
		case 99:
			goto tr24
		case 102:
			goto tr25
		case 105:
			goto tr32
		case 110:
			goto tr33
		case 113:
			goto tr28
		case 115:
			goto tr34
		case 116:
			goto tr35
		}
		switch {
		case data[p] < 109:
			if 48 <= data[p] && data[p] <= 57 {
				goto tr23
			}
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto tr24
			}
		default:
			goto tr24
		}
		goto st0
tr22:
//line parser.rl:19
 mark = p 
	goto st8
	st8:
		if p++; p == pe {
			goto _test_eof8
		}
	st_case_8:
//line parser.go:528
		switch data[p] {
		case 43:
			goto st9
		case 45:
			goto st9
		case 70:
			goto st60
		case 73:
			goto st64
		case 78:
			goto st71
		case 81:
			goto st73
		case 83:
			goto st75
		case 84:
			goto st76
		case 90:
			goto st78
		case 102:
			goto st60
		case 105:
			goto st64
		case 110:
			goto st74
		case 113:
			goto st73
		case 115:
			goto st73
		case 116:
			goto st76
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st21
		}
		goto st0
	st9:
		if p++; p == pe {
			goto _test_eof9
		}
	st_case_9:
		if data[p] == 83 {
			goto st10
		}
		goto st0
	st10:
		if p++; p == pe {
			goto _test_eof10
		}
	st_case_10:
		if data[p] == 117 {
			goto st11
		}
		goto st0
	st11:
		if p++; p == pe {
			goto _test_eof11
		}
	st_case_11:
		if data[p] == 98 {
			goto st12
		}
		goto st0
	st12:
		if p++; p == pe {
			goto _test_eof12
		}
	st_case_12:
		if data[p] == 110 {
			goto st13
		}
		goto st0
	st13:
		if p++; p == pe {
			goto _test_eof13
		}
	st_case_13:
		if data[p] == 111 {
			goto st14
		}
		goto st0
	st14:
		if p++; p == pe {
			goto _test_eof14
		}
	st_case_14:
		if data[p] == 114 {
			goto st15
		}
		goto st0
	st15:
		if p++; p == pe {
			goto _test_eof15
		}
	st_case_15:
		if data[p] == 109 {
			goto st16
		}
		goto st0
	st16:
		if p++; p == pe {
			goto _test_eof16
		}
	st_case_16:
		if data[p] == 97 {
			goto st17
		}
		goto st0
	st17:
		if p++; p == pe {
			goto _test_eof17
		}
	st_case_17:
		if data[p] == 108 {
			goto st18
		}
		goto st0
	st18:
		if p++; p == pe {
			goto _test_eof18
		}
	st_case_18:
		if data[p] == 32 {
			goto tr55
		}
		goto st0
tr55:
//line parser.rl:39
 c.Inputs = append(c.Inputs, Data(data[mark:p])) 
	goto st19
tr128:
//line parser.rl:36

			c.Trap = ConditionFromString(string(data[mark:p]))
        
//line parser.rl:39
 c.Inputs = append(c.Inputs, Data(data[mark:p])) 
	goto st19
	st19:
		if p++; p == pe {
			goto _test_eof19
		}
	st_case_19:
//line parser.go:672
		switch data[p] {
		case 32:
			goto st19
		case 43:
			goto tr22
		case 45:
			goto tr57
		case 70:
			goto tr25
		case 73:
			goto tr26
		case 78:
			goto tr27
		case 81:
			goto tr28
		case 83:
			goto tr29
		case 84:
			goto tr30
		case 90:
			goto tr31
		case 102:
			goto tr25
		case 105:
			goto tr26
		case 110:
			goto tr58
		case 113:
			goto tr28
		case 115:
			goto tr28
		case 116:
			goto tr30
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto tr23
		}
		goto st0
tr57:
//line parser.rl:19
 mark = p 
	goto st20
	st20:
		if p++; p == pe {
			goto _test_eof20
		}
	st_case_20:
//line parser.go:720
		switch data[p] {
		case 43:
			goto st9
		case 45:
			goto st9
		case 62:
			goto st27
		case 70:
			goto st60
		case 73:
			goto st64
		case 78:
			goto st71
		case 81:
			goto st73
		case 83:
			goto st75
		case 84:
			goto st76
		case 90:
			goto st78
		case 102:
			goto st60
		case 105:
			goto st64
		case 110:
			goto st74
		case 113:
			goto st73
		case 115:
			goto st73
		case 116:
			goto st76
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st21
		}
		goto st0
tr23:
//line parser.rl:19
 mark = p 
	goto st21
	st21:
		if p++; p == pe {
			goto _test_eof21
		}
	st_case_21:
//line parser.go:768
		switch data[p] {
		case 32:
			goto tr55
		case 43:
			goto st22
		case 45:
			goto st22
		case 46:
			goto st24
		case 69:
			goto st26
		case 101:
			goto st26
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st21
		}
		goto st0
	st22:
		if p++; p == pe {
			goto _test_eof22
		}
	st_case_22:
		if 48 <= data[p] && data[p] <= 57 {
			goto st23
		}
		goto st0
	st23:
		if p++; p == pe {
			goto _test_eof23
		}
	st_case_23:
		if data[p] == 32 {
			goto tr55
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st23
		}
		goto st0
	st24:
		if p++; p == pe {
			goto _test_eof24
		}
	st_case_24:
		if 48 <= data[p] && data[p] <= 57 {
			goto st25
		}
		goto st0
	st25:
		if p++; p == pe {
			goto _test_eof25
		}
	st_case_25:
		switch data[p] {
		case 32:
			goto tr55
		case 43:
			goto st22
		case 45:
			goto st22
		case 69:
			goto st26
		case 101:
			goto st26
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st25
		}
		goto st0
	st26:
		if p++; p == pe {
			goto _test_eof26
		}
	st_case_26:
		switch data[p] {
		case 43:
			goto st22
		case 45:
			goto st22
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st23
		}
		goto st0
	st27:
		if p++; p == pe {
			goto _test_eof27
		}
	st_case_27:
		if data[p] == 32 {
			goto st28
		}
		goto st0
	st28:
		if p++; p == pe {
			goto _test_eof28
		}
	st_case_28:
		switch data[p] {
		case 32:
			goto st28
		case 35:
			goto tr66
		case 43:
			goto tr67
		case 45:
			goto tr67
		case 70:
			goto tr69
		case 73:
			goto tr70
		case 78:
			goto tr71
		case 81:
			goto tr72
		case 83:
			goto tr73
		case 84:
			goto tr74
		case 90:
			goto tr75
		case 102:
			goto tr69
		case 105:
			goto tr70
		case 110:
			goto tr76
		case 113:
			goto tr72
		case 115:
			goto tr72
		case 116:
			goto tr74
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto tr68
		}
		goto st0
tr66:
//line parser.rl:19
 mark = p 
	goto st118
	st118:
		if p++; p == pe {
			goto _test_eof118
		}
	st_case_118:
//line parser.go:916
		if data[p] == 32 {
			goto tr150
		}
		goto st0
tr150:
//line parser.rl:40
 c.Output = Data(data[mark:p]) 
	goto st119
	st119:
		if p++; p == pe {
			goto _test_eof119
		}
	st_case_119:
//line parser.go:930
		switch data[p] {
		case 32:
			goto st119
		case 63:
			goto tr152
		case 99:
			goto tr152
		case 105:
			goto tr152
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto tr152
			}
		case data[p] >= 109:
			goto tr152
		}
		goto st0
tr152:
//line parser.rl:19
 mark = p 
	goto st120
	st120:
		if p++; p == pe {
			goto _test_eof120
		}
	st_case_120:
//line parser.go:959
		switch data[p] {
		case 63:
			goto st120
		case 99:
			goto st120
		case 105:
			goto st120
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st120
			}
		case data[p] >= 109:
			goto st120
		}
		goto st0
tr67:
//line parser.rl:19
 mark = p 
	goto st29
	st29:
		if p++; p == pe {
			goto _test_eof29
		}
	st_case_29:
//line parser.go:986
		switch data[p] {
		case 43:
			goto st30
		case 45:
			goto st30
		case 70:
			goto st42
		case 73:
			goto st46
		case 78:
			goto st52
		case 81:
			goto st125
		case 83:
			goto st126
		case 84:
			goto st55
		case 90:
			goto st57
		case 102:
			goto st42
		case 105:
			goto st46
		case 110:
			goto st54
		case 113:
			goto st125
		case 115:
			goto st125
		case 116:
			goto st55
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st121
		}
		goto st0
	st30:
		if p++; p == pe {
			goto _test_eof30
		}
	st_case_30:
		if data[p] == 83 {
			goto st31
		}
		goto st0
	st31:
		if p++; p == pe {
			goto _test_eof31
		}
	st_case_31:
		if data[p] == 117 {
			goto st32
		}
		goto st0
	st32:
		if p++; p == pe {
			goto _test_eof32
		}
	st_case_32:
		if data[p] == 98 {
			goto st33
		}
		goto st0
	st33:
		if p++; p == pe {
			goto _test_eof33
		}
	st_case_33:
		if data[p] == 110 {
			goto st34
		}
		goto st0
	st34:
		if p++; p == pe {
			goto _test_eof34
		}
	st_case_34:
		if data[p] == 111 {
			goto st35
		}
		goto st0
	st35:
		if p++; p == pe {
			goto _test_eof35
		}
	st_case_35:
		if data[p] == 114 {
			goto st36
		}
		goto st0
	st36:
		if p++; p == pe {
			goto _test_eof36
		}
	st_case_36:
		if data[p] == 109 {
			goto st37
		}
		goto st0
	st37:
		if p++; p == pe {
			goto _test_eof37
		}
	st_case_37:
		if data[p] == 97 {
			goto st38
		}
		goto st0
	st38:
		if p++; p == pe {
			goto _test_eof38
		}
	st_case_38:
		if data[p] == 108 {
			goto st118
		}
		goto st0
tr68:
//line parser.rl:19
 mark = p 
	goto st121
	st121:
		if p++; p == pe {
			goto _test_eof121
		}
	st_case_121:
//line parser.go:1113
		switch data[p] {
		case 32:
			goto tr150
		case 43:
			goto st39
		case 45:
			goto st39
		case 46:
			goto st40
		case 69:
			goto st41
		case 101:
			goto st41
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st121
		}
		goto st0
	st39:
		if p++; p == pe {
			goto _test_eof39
		}
	st_case_39:
		if 48 <= data[p] && data[p] <= 57 {
			goto st122
		}
		goto st0
	st122:
		if p++; p == pe {
			goto _test_eof122
		}
	st_case_122:
		if data[p] == 32 {
			goto tr150
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st122
		}
		goto st0
	st40:
		if p++; p == pe {
			goto _test_eof40
		}
	st_case_40:
		if 48 <= data[p] && data[p] <= 57 {
			goto st123
		}
		goto st0
	st123:
		if p++; p == pe {
			goto _test_eof123
		}
	st_case_123:
		switch data[p] {
		case 32:
			goto tr150
		case 43:
			goto st39
		case 45:
			goto st39
		case 69:
			goto st41
		case 101:
			goto st41
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st123
		}
		goto st0
	st41:
		if p++; p == pe {
			goto _test_eof41
		}
	st_case_41:
		switch data[p] {
		case 43:
			goto st39
		case 45:
			goto st39
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto st122
		}
		goto st0
tr69:
//line parser.rl:19
 mark = p 
	goto st42
	st42:
		if p++; p == pe {
			goto _test_eof42
		}
	st_case_42:
//line parser.go:1207
		switch data[p] {
		case 65:
			goto st43
		case 97:
			goto st43
		}
		goto st0
	st43:
		if p++; p == pe {
			goto _test_eof43
		}
	st_case_43:
		switch data[p] {
		case 76:
			goto st44
		case 108:
			goto st44
		}
		goto st0
	st44:
		if p++; p == pe {
			goto _test_eof44
		}
	st_case_44:
		switch data[p] {
		case 83:
			goto st45
		case 115:
			goto st45
		}
		goto st0
	st45:
		if p++; p == pe {
			goto _test_eof45
		}
	st_case_45:
		switch data[p] {
		case 69:
			goto st118
		case 101:
			goto st118
		}
		goto st0
tr70:
//line parser.rl:19
 mark = p 
	goto st46
	st46:
		if p++; p == pe {
			goto _test_eof46
		}
	st_case_46:
//line parser.go:1260
		switch data[p] {
		case 78:
			goto st47
		case 110:
			goto st47
		}
		goto st0
	st47:
		if p++; p == pe {
			goto _test_eof47
		}
	st_case_47:
		switch data[p] {
		case 70:
			goto st124
		case 102:
			goto st124
		}
		goto st0
	st124:
		if p++; p == pe {
			goto _test_eof124
		}
	st_case_124:
		switch data[p] {
		case 32:
			goto tr150
		case 73:
			goto st48
		case 105:
			goto st48
		}
		goto st0
	st48:
		if p++; p == pe {
			goto _test_eof48
		}
	st_case_48:
		switch data[p] {
		case 78:
			goto st49
		case 110:
			goto st49
		}
		goto st0
	st49:
		if p++; p == pe {
			goto _test_eof49
		}
	st_case_49:
		switch data[p] {
		case 73:
			goto st50
		case 105:
			goto st50
		}
		goto st0
	st50:
		if p++; p == pe {
			goto _test_eof50
		}
	st_case_50:
		switch data[p] {
		case 84:
			goto st51
		case 116:
			goto st51
		}
		goto st0
	st51:
		if p++; p == pe {
			goto _test_eof51
		}
	st_case_51:
		switch data[p] {
		case 89:
			goto st118
		case 121:
			goto st118
		}
		goto st0
tr71:
//line parser.rl:19
 mark = p 
	goto st52
	st52:
		if p++; p == pe {
			goto _test_eof52
		}
	st_case_52:
//line parser.go:1351
		switch data[p] {
		case 65:
			goto st53
		case 97:
			goto st53
		case 111:
			goto st35
		}
		goto st0
	st53:
		if p++; p == pe {
			goto _test_eof53
		}
	st_case_53:
		switch data[p] {
		case 78:
			goto st118
		case 110:
			goto st118
		}
		goto st0
tr72:
//line parser.rl:19
 mark = p 
	goto st125
	st125:
		if p++; p == pe {
			goto _test_eof125
		}
	st_case_125:
//line parser.go:1382
		switch data[p] {
		case 32:
			goto tr150
		case 78:
			goto st54
		case 110:
			goto st54
		}
		goto st0
tr76:
//line parser.rl:19
 mark = p 
	goto st54
	st54:
		if p++; p == pe {
			goto _test_eof54
		}
	st_case_54:
//line parser.go:1401
		switch data[p] {
		case 65:
			goto st53
		case 97:
			goto st53
		}
		goto st0
tr73:
//line parser.rl:19
 mark = p 
	goto st126
	st126:
		if p++; p == pe {
			goto _test_eof126
		}
	st_case_126:
//line parser.go:1418
		switch data[p] {
		case 32:
			goto tr150
		case 78:
			goto st54
		case 110:
			goto st54
		case 117:
			goto st32
		}
		goto st0
tr74:
//line parser.rl:19
 mark = p 
	goto st55
	st55:
		if p++; p == pe {
			goto _test_eof55
		}
	st_case_55:
//line parser.go:1439
		switch data[p] {
		case 82:
			goto st56
		case 114:
			goto st56
		}
		goto st0
	st56:
		if p++; p == pe {
			goto _test_eof56
		}
	st_case_56:
		switch data[p] {
		case 85:
			goto st45
		case 117:
			goto st45
		}
		goto st0
tr75:
//line parser.rl:19
 mark = p 
	goto st57
	st57:
		if p++; p == pe {
			goto _test_eof57
		}
	st_case_57:
//line parser.go:1468
		if data[p] == 101 {
			goto st58
		}
		goto st0
	st58:
		if p++; p == pe {
			goto _test_eof58
		}
	st_case_58:
		if data[p] == 114 {
			goto st59
		}
		goto st0
	st59:
		if p++; p == pe {
			goto _test_eof59
		}
	st_case_59:
		if data[p] == 111 {
			goto st118
		}
		goto st0
tr25:
//line parser.rl:19
 mark = p 
	goto st60
	st60:
		if p++; p == pe {
			goto _test_eof60
		}
	st_case_60:
//line parser.go:1500
		switch data[p] {
		case 65:
			goto st61
		case 97:
			goto st61
		}
		goto st0
	st61:
		if p++; p == pe {
			goto _test_eof61
		}
	st_case_61:
		switch data[p] {
		case 76:
			goto st62
		case 108:
			goto st62
		}
		goto st0
	st62:
		if p++; p == pe {
			goto _test_eof62
		}
	st_case_62:
		switch data[p] {
		case 83:
			goto st63
		case 115:
			goto st63
		}
		goto st0
	st63:
		if p++; p == pe {
			goto _test_eof63
		}
	st_case_63:
		switch data[p] {
		case 69:
			goto st18
		case 101:
			goto st18
		}
		goto st0
tr26:
//line parser.rl:19
 mark = p 
	goto st64
	st64:
		if p++; p == pe {
			goto _test_eof64
		}
	st_case_64:
//line parser.go:1553
		switch data[p] {
		case 78:
			goto st65
		case 110:
			goto st65
		}
		goto st0
	st65:
		if p++; p == pe {
			goto _test_eof65
		}
	st_case_65:
		switch data[p] {
		case 70:
			goto st66
		case 102:
			goto st66
		}
		goto st0
	st66:
		if p++; p == pe {
			goto _test_eof66
		}
	st_case_66:
		switch data[p] {
		case 32:
			goto tr55
		case 73:
			goto st67
		case 105:
			goto st67
		}
		goto st0
	st67:
		if p++; p == pe {
			goto _test_eof67
		}
	st_case_67:
		switch data[p] {
		case 78:
			goto st68
		case 110:
			goto st68
		}
		goto st0
	st68:
		if p++; p == pe {
			goto _test_eof68
		}
	st_case_68:
		switch data[p] {
		case 73:
			goto st69
		case 105:
			goto st69
		}
		goto st0
	st69:
		if p++; p == pe {
			goto _test_eof69
		}
	st_case_69:
		switch data[p] {
		case 84:
			goto st70
		case 116:
			goto st70
		}
		goto st0
	st70:
		if p++; p == pe {
			goto _test_eof70
		}
	st_case_70:
		switch data[p] {
		case 89:
			goto st18
		case 121:
			goto st18
		}
		goto st0
tr27:
//line parser.rl:19
 mark = p 
	goto st71
	st71:
		if p++; p == pe {
			goto _test_eof71
		}
	st_case_71:
//line parser.go:1644
		switch data[p] {
		case 65:
			goto st72
		case 97:
			goto st72
		case 111:
			goto st14
		}
		goto st0
	st72:
		if p++; p == pe {
			goto _test_eof72
		}
	st_case_72:
		switch data[p] {
		case 78:
			goto st18
		case 110:
			goto st18
		}
		goto st0
tr28:
//line parser.rl:19
 mark = p 
	goto st73
	st73:
		if p++; p == pe {
			goto _test_eof73
		}
	st_case_73:
//line parser.go:1675
		switch data[p] {
		case 32:
			goto tr55
		case 78:
			goto st74
		case 110:
			goto st74
		}
		goto st0
tr58:
//line parser.rl:19
 mark = p 
	goto st74
	st74:
		if p++; p == pe {
			goto _test_eof74
		}
	st_case_74:
//line parser.go:1694
		switch data[p] {
		case 65:
			goto st72
		case 97:
			goto st72
		}
		goto st0
tr29:
//line parser.rl:19
 mark = p 
	goto st75
	st75:
		if p++; p == pe {
			goto _test_eof75
		}
	st_case_75:
//line parser.go:1711
		switch data[p] {
		case 32:
			goto tr55
		case 78:
			goto st74
		case 110:
			goto st74
		case 117:
			goto st11
		}
		goto st0
tr30:
//line parser.rl:19
 mark = p 
	goto st76
	st76:
		if p++; p == pe {
			goto _test_eof76
		}
	st_case_76:
//line parser.go:1732
		switch data[p] {
		case 82:
			goto st77
		case 114:
			goto st77
		}
		goto st0
	st77:
		if p++; p == pe {
			goto _test_eof77
		}
	st_case_77:
		switch data[p] {
		case 85:
			goto st63
		case 117:
			goto st63
		}
		goto st0
tr31:
//line parser.rl:19
 mark = p 
	goto st78
	st78:
		if p++; p == pe {
			goto _test_eof78
		}
	st_case_78:
//line parser.go:1761
		if data[p] == 101 {
			goto st79
		}
		goto st0
	st79:
		if p++; p == pe {
			goto _test_eof79
		}
	st_case_79:
		if data[p] == 114 {
			goto st80
		}
		goto st0
	st80:
		if p++; p == pe {
			goto _test_eof80
		}
	st_case_80:
		if data[p] == 111 {
			goto st18
		}
		goto st0
tr24:
//line parser.rl:19
 mark = p 
	goto st81
	st81:
		if p++; p == pe {
			goto _test_eof81
		}
	st_case_81:
//line parser.go:1793
		switch data[p] {
		case 32:
			goto tr124
		case 63:
			goto st81
		case 99:
			goto st81
		case 105:
			goto st81
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
tr124:
//line parser.rl:36

			c.Trap = ConditionFromString(string(data[mark:p]))
        
	goto st82
	st82:
		if p++; p == pe {
			goto _test_eof82
		}
	st_case_82:
//line parser.go:1824
		switch data[p] {
		case 32:
			goto st82
		case 43:
			goto tr22
		case 45:
			goto tr22
		case 70:
			goto tr25
		case 73:
			goto tr26
		case 78:
			goto tr27
		case 81:
			goto tr28
		case 83:
			goto tr29
		case 84:
			goto tr30
		case 90:
			goto tr31
		case 102:
			goto tr25
		case 105:
			goto tr26
		case 110:
			goto tr58
		case 113:
			goto tr28
		case 115:
			goto tr28
		case 116:
			goto tr30
		}
		if 48 <= data[p] && data[p] <= 57 {
			goto tr23
		}
		goto st0
tr32:
//line parser.rl:19
 mark = p 
	goto st83
	st83:
		if p++; p == pe {
			goto _test_eof83
		}
	st_case_83:
//line parser.go:1872
		switch data[p] {
		case 32:
			goto tr124
		case 63:
			goto st81
		case 78:
			goto st65
		case 99:
			goto st81
		case 105:
			goto st81
		case 110:
			goto st84
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
	st84:
		if p++; p == pe {
			goto _test_eof84
		}
	st_case_84:
		switch data[p] {
		case 32:
			goto tr124
		case 63:
			goto st81
		case 70:
			goto st66
		case 99:
			goto st81
		case 102:
			goto st66
		case 105:
			goto st81
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
tr33:
//line parser.rl:19
 mark = p 
	goto st85
	st85:
		if p++; p == pe {
			goto _test_eof85
		}
	st_case_85:
//line parser.go:1933
		switch data[p] {
		case 32:
			goto tr124
		case 63:
			goto st81
		case 65:
			goto st72
		case 97:
			goto st72
		case 99:
			goto st81
		case 105:
			goto st81
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
tr34:
//line parser.rl:19
 mark = p 
	goto st86
	st86:
		if p++; p == pe {
			goto _test_eof86
		}
	st_case_86:
//line parser.go:1966
		switch data[p] {
		case 32:
			goto tr128
		case 63:
			goto st81
		case 78:
			goto st74
		case 99:
			goto st81
		case 105:
			goto st81
		case 110:
			goto st85
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
tr35:
//line parser.rl:19
 mark = p 
	goto st87
	st87:
		if p++; p == pe {
			goto _test_eof87
		}
	st_case_87:
//line parser.go:1999
		switch data[p] {
		case 32:
			goto tr124
		case 63:
			goto st81
		case 82:
			goto st77
		case 99:
			goto st81
		case 105:
			goto st81
		case 114:
			goto st88
		}
		switch {
		case data[p] > 111:
			if 115 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
	st88:
		if p++; p == pe {
			goto _test_eof88
		}
	st_case_88:
		switch data[p] {
		case 32:
			goto tr124
		case 63:
			goto st81
		case 85:
			goto st63
		case 99:
			goto st81
		case 105:
			goto st81
		case 117:
			goto st89
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
	st89:
		if p++; p == pe {
			goto _test_eof89
		}
	st_case_89:
		switch data[p] {
		case 32:
			goto tr124
		case 63:
			goto st81
		case 69:
			goto st18
		case 99:
			goto st81
		case 101:
			goto st18
		case 105:
			goto st81
		}
		switch {
		case data[p] > 111:
			if 114 <= data[p] && data[p] <= 122 {
				goto st81
			}
		case data[p] >= 109:
			goto st81
		}
		goto st0
tr19:
//line parser.rl:19
 mark = p 
	goto st90
	st90:
		if p++; p == pe {
			goto _test_eof90
		}
	st_case_90:
//line parser.go:2088
		switch data[p] {
		case 48:
			goto st6
		case 94:
			goto st6
		}
		goto st0
tr4:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st91
	st91:
		if p++; p == pe {
			goto _test_eof91
		}
	st_case_91:
//line parser.go:2111
		switch data[p] {
		case 32:
			goto tr16
		case 45:
			goto st4
		}
		goto st0
tr6:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st92
	st92:
		if p++; p == pe {
			goto _test_eof92
		}
	st_case_92:
//line parser.go:2134
		switch data[p] {
		case 65:
			goto st4
		case 67:
			goto st4
		}
		goto st0
tr7:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st93
	st93:
		if p++; p == pe {
			goto _test_eof93
		}
	st_case_93:
//line parser.go:2157
		if data[p] == 113 {
			goto st94
		}
		goto st0
	st94:
		if p++; p == pe {
			goto _test_eof94
		}
	st_case_94:
		if data[p] == 117 {
			goto st95
		}
		goto st0
	st95:
		if p++; p == pe {
			goto _test_eof95
		}
	st_case_95:
		if data[p] == 97 {
			goto st96
		}
		goto st0
	st96:
		if p++; p == pe {
			goto _test_eof96
		}
	st_case_96:
		if data[p] == 110 {
			goto st97
		}
		goto st0
	st97:
		if p++; p == pe {
			goto _test_eof97
		}
	st_case_97:
		if data[p] == 116 {
			goto st4
		}
		goto st0
tr8:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st98
	st98:
		if p++; p == pe {
			goto _test_eof98
		}
	st_case_98:
//line parser.go:2213
		switch data[p] {
		case 32:
			goto tr16
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
			goto st99
		}
		goto st0
	st99:
		if p++; p == pe {
			goto _test_eof99
		}
	st_case_99:
		switch data[p] {
		case 32:
			goto tr16
		case 78:
			goto st4
		}
		goto st0
tr9:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st100
	st100:
		if p++; p == pe {
			goto _test_eof100
		}
	st_case_100:
//line parser.go:2260
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
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st101
	st101:
		if p++; p == pe {
			goto _test_eof101
		}
	st_case_101:
//line parser.go:2285
		switch data[p] {
		case 100:
			goto st102
		case 102:
			goto st103
		case 105:
			goto st102
		case 112:
			goto st4
		}
		goto st0
	st102:
		if p++; p == pe {
			goto _test_eof102
		}
	st_case_102:
		if data[p] == 102 {
			goto st4
		}
		goto st0
	st103:
		if p++; p == pe {
			goto _test_eof103
		}
	st_case_103:
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
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st104
	st104:
		if p++; p == pe {
			goto _test_eof104
		}
	st_case_104:
//line parser.go:2335
		switch data[p] {
		case 113:
			goto st4
		case 120:
			goto st105
		}
		goto st0
	st105:
		if p++; p == pe {
			goto _test_eof105
		}
	st_case_105:
		if data[p] == 112 {
			goto st4
		}
		goto st0
tr12:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st106
	st106:
		if p++; p == pe {
			goto _test_eof106
		}
	st_case_106:
//line parser.go:2367
		if data[p] == 111 {
			goto st107
		}
		goto st0
	st107:
		if p++; p == pe {
			goto _test_eof107
		}
	st_case_107:
		if data[p] == 103 {
			goto st108
		}
		goto st0
	st108:
		if p++; p == pe {
			goto _test_eof108
		}
	st_case_108:
		switch data[p] {
		case 32:
			goto tr16
		case 49:
			goto st109
		}
		goto st0
	st109:
		if p++; p == pe {
			goto _test_eof109
		}
	st_case_109:
		if data[p] == 48 {
			goto st4
		}
		goto st0
tr13:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st110
	st110:
		if p++; p == pe {
			goto _test_eof110
		}
	st_case_110:
//line parser.go:2417
		switch data[p] {
		case 67:
			goto st4
		case 117:
			goto st95
		}
		goto st0
tr14:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st111
	st111:
		if p++; p == pe {
			goto _test_eof111
		}
	st_case_111:
//line parser.go:2440
		switch data[p] {
		case 97:
			goto st97
		case 102:
			goto st112
		}
		goto st0
	st112:
		if p++; p == pe {
			goto _test_eof112
		}
	st_case_112:
		if data[p] == 105 {
			goto st4
		}
		goto st0
tr15:
//line parser.rl:21

            if c.Prec, err = strconv.Atoi(string(data[mark:p])); err != nil {
                return c, err
            }
        
//line parser.rl:19
 mark = p 
	goto st113
	st113:
		if p++; p == pe {
			goto _test_eof113
		}
	st_case_113:
//line parser.go:2472
		switch data[p] {
		case 67:
			goto st4
		case 105:
			goto st114
		}
		goto st0
	st114:
		if p++; p == pe {
			goto _test_eof114
		}
	st_case_114:
		if data[p] == 103 {
			goto st115
		}
		goto st0
	st115:
		if p++; p == pe {
			goto _test_eof115
		}
	st_case_115:
		if data[p] == 110 {
			goto st116
		}
		goto st0
	st116:
		if p++; p == pe {
			goto _test_eof116
		}
	st_case_116:
		switch data[p] {
		case 32:
			goto tr16
		case 98:
			goto st117
		}
		goto st0
	st117:
		if p++; p == pe {
			goto _test_eof117
		}
	st_case_117:
		if data[p] == 105 {
			goto st97
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
	_test_eof14: cs = 14; goto _test_eof
	_test_eof15: cs = 15; goto _test_eof
	_test_eof16: cs = 16; goto _test_eof
	_test_eof17: cs = 17; goto _test_eof
	_test_eof18: cs = 18; goto _test_eof
	_test_eof19: cs = 19; goto _test_eof
	_test_eof20: cs = 20; goto _test_eof
	_test_eof21: cs = 21; goto _test_eof
	_test_eof22: cs = 22; goto _test_eof
	_test_eof23: cs = 23; goto _test_eof
	_test_eof24: cs = 24; goto _test_eof
	_test_eof25: cs = 25; goto _test_eof
	_test_eof26: cs = 26; goto _test_eof
	_test_eof27: cs = 27; goto _test_eof
	_test_eof28: cs = 28; goto _test_eof
	_test_eof118: cs = 118; goto _test_eof
	_test_eof119: cs = 119; goto _test_eof
	_test_eof120: cs = 120; goto _test_eof
	_test_eof29: cs = 29; goto _test_eof
	_test_eof30: cs = 30; goto _test_eof
	_test_eof31: cs = 31; goto _test_eof
	_test_eof32: cs = 32; goto _test_eof
	_test_eof33: cs = 33; goto _test_eof
	_test_eof34: cs = 34; goto _test_eof
	_test_eof35: cs = 35; goto _test_eof
	_test_eof36: cs = 36; goto _test_eof
	_test_eof37: cs = 37; goto _test_eof
	_test_eof38: cs = 38; goto _test_eof
	_test_eof121: cs = 121; goto _test_eof
	_test_eof39: cs = 39; goto _test_eof
	_test_eof122: cs = 122; goto _test_eof
	_test_eof40: cs = 40; goto _test_eof
	_test_eof123: cs = 123; goto _test_eof
	_test_eof41: cs = 41; goto _test_eof
	_test_eof42: cs = 42; goto _test_eof
	_test_eof43: cs = 43; goto _test_eof
	_test_eof44: cs = 44; goto _test_eof
	_test_eof45: cs = 45; goto _test_eof
	_test_eof46: cs = 46; goto _test_eof
	_test_eof47: cs = 47; goto _test_eof
	_test_eof124: cs = 124; goto _test_eof
	_test_eof48: cs = 48; goto _test_eof
	_test_eof49: cs = 49; goto _test_eof
	_test_eof50: cs = 50; goto _test_eof
	_test_eof51: cs = 51; goto _test_eof
	_test_eof52: cs = 52; goto _test_eof
	_test_eof53: cs = 53; goto _test_eof
	_test_eof125: cs = 125; goto _test_eof
	_test_eof54: cs = 54; goto _test_eof
	_test_eof126: cs = 126; goto _test_eof
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
	_test_eof81: cs = 81; goto _test_eof
	_test_eof82: cs = 82; goto _test_eof
	_test_eof83: cs = 83; goto _test_eof
	_test_eof84: cs = 84; goto _test_eof
	_test_eof85: cs = 85; goto _test_eof
	_test_eof86: cs = 86; goto _test_eof
	_test_eof87: cs = 87; goto _test_eof
	_test_eof88: cs = 88; goto _test_eof
	_test_eof89: cs = 89; goto _test_eof
	_test_eof90: cs = 90; goto _test_eof
	_test_eof91: cs = 91; goto _test_eof
	_test_eof92: cs = 92; goto _test_eof
	_test_eof93: cs = 93; goto _test_eof
	_test_eof94: cs = 94; goto _test_eof
	_test_eof95: cs = 95; goto _test_eof
	_test_eof96: cs = 96; goto _test_eof
	_test_eof97: cs = 97; goto _test_eof
	_test_eof98: cs = 98; goto _test_eof
	_test_eof99: cs = 99; goto _test_eof
	_test_eof100: cs = 100; goto _test_eof
	_test_eof101: cs = 101; goto _test_eof
	_test_eof102: cs = 102; goto _test_eof
	_test_eof103: cs = 103; goto _test_eof
	_test_eof104: cs = 104; goto _test_eof
	_test_eof105: cs = 105; goto _test_eof
	_test_eof106: cs = 106; goto _test_eof
	_test_eof107: cs = 107; goto _test_eof
	_test_eof108: cs = 108; goto _test_eof
	_test_eof109: cs = 109; goto _test_eof
	_test_eof110: cs = 110; goto _test_eof
	_test_eof111: cs = 111; goto _test_eof
	_test_eof112: cs = 112; goto _test_eof
	_test_eof113: cs = 113; goto _test_eof
	_test_eof114: cs = 114; goto _test_eof
	_test_eof115: cs = 115; goto _test_eof
	_test_eof116: cs = 116; goto _test_eof
	_test_eof117: cs = 117; goto _test_eof

	_test_eof: {}
	if p == eof {
		switch cs {
		case 118, 121, 122, 123, 124, 125, 126:
//line parser.rl:40
 c.Output = Data(data[mark:p]) 
		case 120:
//line parser.rl:41

			c.Excep = ConditionFromString(string(data[mark:p]))
        
		case 119:
//line parser.rl:19
 mark = p 
//line parser.rl:41

			c.Excep = ConditionFromString(string(data[mark:p]))
        
//line parser.go:2664
		}
	}

	_out: {}
	}

//line parser.rl:162

    return c, nil
}
