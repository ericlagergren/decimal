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

	%%{
		machine parser;

		infinity = 'inf'i 'inity'i?;
		nan      = 'nan'i digit*;

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
