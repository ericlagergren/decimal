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

func ParseSpecial(data []byte) Special {
	cs, p, pe := 0, 0, len(data)

	%%{
		machine parser;
		
		sign           = [-+];
		indicator      = [eE];
		decimal_part   = digit+ '.' digit* | '.'? digit+;
		exponent_part  = indicator sign? digit+;
		infinity       = 'inf'i 'inity'i?;
		nan            = [sSqQ]? 'nan'i digit*;
		numeric_value  = decimal_part exponent_part? | infinity;
		numeric_string = sign? numeric_value | sign? nan;

		main := (
			  '+'?  infinity @{ return PInf }
			| '-'   infinity @{ return NInf }
			| 'q'i? nan      @{ return QNaN }
			| 's'i  nan      @{ return SNaN }
		);

		write data;
		write init;
		write exec;
	}%%
	return Invalid
}
