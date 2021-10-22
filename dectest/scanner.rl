package dectest

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	. "github.com/ericlagergren/decimal"
)

type Scanner struct {
	precision   int
	maxExponent int
	minExponent int
	clamp       int
	rounding    RoundingMode
	extended    int
	c           *Case
	s           *bufio.Scanner
	err         error
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{s: bufio.NewScanner(r)}
}

func (s *Scanner) Scan() bool {
	s.c = nil
	for s.c == nil {
		if !s.s.Scan() {
			return false
		}
		if s.err = s.parse(s.s.Bytes()); s.err != nil {
			return false
		}
		if s.c != nil {
			return true
		}
	}
	return false
}

func (s *Scanner) Case() *Case {
	s.c.Clamp = s.clamp == 1
	s.c.Prec = s.precision
	s.c.Mode = s.rounding
	s.c.MinScale = s.minExponent
	s.c.MaxScale = s.maxExponent
	return s.c
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) parse(data []byte) (err error) {
	cs, p, pe, eof := 0, 0, len(data), len(data)
	var (
		mark int
		ok   bool
		cond Condition
	)

	%%{
		machine parser;

		action mark { mark = fpc }

		action set_id {
			s.c = &Case{ID: string(data[mark:fpc])}
		}

		action set_clamp {
			if s.clamp, err = strconv.Atoi(string(data[mark:fpc])); err != nil {
				return err
			}
		}

		action set_op {
			if s.c.Op, ok = operations[strings.ToLower(string(data[mark:fpc]))]; !ok {
				return fmt.Errorf("dectest: invalid op: %q", data[mark:fpc])
			}
		}

		action set_precision {
			s.precision, err = strconv.Atoi(string(data[mark:fpc]))
			if err != nil {
				return err
			}
		}

		action set_max_exponent {
			s.maxExponent, err = strconv.Atoi(string(data[mark:fpc]))
			if err != nil {
				return err
			}
		}

		action set_min_exponent {
			s.minExponent, err = strconv.Atoi(string(data[mark:fpc]))
			if err != nil {
				return err
			}
		}

		action set_rounding {
			if s.rounding, ok = roundingModes[strings.ToLower(string(data[mark:fpc]))]; !ok {
				return fmt.Errorf("unknown rounding mode: %q", data[mark:fpc])
			}
		}

		action add_input  { s.c.Inputs = append(s.c.Inputs, Data(data[mark:fpc])) }
		action set_output { s.c.Output = Data(data[mark:fpc]) }
		action add_condition {
			cond, ok = conditions[strings.ToLower(string(data[mark:fpc]))]
			if !ok {
				return fmt.Errorf("unknown condition: %q", data[mark:fpc])
			}
			s.c.Conditions |= cond
		}

		sign = '+' | '-';

		precision = ( digit+ ) >mark %set_precision;
		clamp = ( digit+ ) >mark %set_clamp;
		max_exponent = ( sign? digit+ ) >mark %set_max_exponent;
		min_exponent = ( sign? digit+ ) >mark %set_min_exponent;

		rounding = (
			  'ceiling'i
			| 'down'i
			| 'floor'i
			| 'half_down'i
			| 'half_even'i
			| 'half_up'i
			| 'up'i
			| '05up'i
		) >mark %set_rounding;

		id = ( alpha{3,} digit{3,} ) >mark %set_id;

		op = (
			  'abs'i
			| 'add'i
			| 'and'i
			| 'apply'i
			| 'canonical'i
			| 'class'i
			| 'compare'i
			| 'comparesig'i
			| 'comparetotal'i
			| 'comparetotmag'i
			| 'copy'i
			| 'copyabs'i
			| 'copynegate'i
			| 'copysign'i
			| 'divide'i
			| 'divideint'i
			| 'exp'i
			| 'fma'i
			| 'invert'i
			| 'ln'i
			| 'log10'i
			| 'logb'i
			| 'max'i
			| 'min'i
			| 'maxmag'i
			| 'minmag'i
			| 'minus'i
			| 'multiply'i
			| 'nextminus'i
			| 'nextplus'i
			| 'nexttoward'i
			| 'or'i
			| 'plus'i
			| 'power'i
			| 'quantize'i
			| 'reduce'i
			| 'remainder'i
			| 'remaindernear'i
			| 'rescale'i
			| 'rotate'i
			| 'samequantum'i
			| 'scaleb'i
			| 'shift'i
			| 'squareroot'i
			| 'subtract'i
			| 'toEng'i
			| 'tointegral'i
			| 'tointegralx'i
			| 'toSci'i
			| 'trim'i
			| 'xor'i
		) >mark %set_op;

		quote      = '\'' | '"';
		indicator  = 'e' | 'E';
		exponent   = indicator? sign? digit+;
		number	   = (digit+ '.' digit* | '.' digit+ | digit+ | digit) exponent?;
		nan_prefix = [sSqQ];
		nan		   = (nan_prefix? 'nan'i digit* | '?');
		class	   = (nan_prefix? 'nan'i | sign? (
				  'Subnormal'
				| 'Normal'
				| 'Zero'
				| 'Infinity')
		);
		numeric_string = sign? (
			  nan				   # S, Q, NaN, sNaN, ...
			| ('inf'i 'inity'i?)   # +inf, -inf, ...
			| number			   # 10, 10.1, +0e-392, ...
		);
		input  = ( numeric_string | '#') >mark %add_input;
		output = ( numeric_string | class | '#') >mark %set_output;

		condition = (
			  'Clamped'i
			| 'Conversion_syntax'i
			| 'Division_by_zero'i
			| 'Division_impossible'i
			| 'Division_undefined'i
			| 'Inexact'i
			| 'Insufficient_storage'i
			| 'Invalid_context'i
			| 'Invalid_operation'i
			| 'Lost_digits'i
			| 'Overflow'i
			| 'Rounded'i
			| 'Subnormal'i
			| 'Underflow'i
		) >mark %add_condition;

		main := (
			  ('precision:'i space+ precision space* any*)
			| ('clamp:'i space+ clamp space* any*)
			| ('maxexponent:'i space+ max_exponent space* any*)
			| ('minexponent:'i space+ min_exponent space* any*)
			| ('rounding:'i space+ rounding space* any*)
			| (id space+ op space+ (quote? input quote? space+)+ '->' space+ quote? output quote? (space+ condition)*)
		);

		write data;
		write init;
		write exec;
	}%%

	return nil
}
