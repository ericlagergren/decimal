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
	case Invalid: return "invalid"
	case QNaN:    return "qNaN"
	case SNaN:    return "sNaN"
	case Inf:     return "+inf"
	default:      panic(fmt.Sprintf("unknown(%d)", s))
	}
}

func ParseSpecial(data string) (s Special, sign bool) {
	sign = len(data) > 0 && data[0] == '-'
	cs, p, pe := 0, 0, len(data)

	%%{
		machine parser;

		sign     = ('-' | '+');
		infinity = 'inf'i 'inity'i?;
		nan      = 'nan'i;

		main := (
			  sign? infinity   @{ return Inf,  sign }
			| sign? 'q'i? nan  @{ return QNaN, sign }
			| sign? 's'i  nan  @{ return SNaN, sign }
		);

		write data;
		write init;
		write exec;
	}%%
	return Invalid, sign 
}
