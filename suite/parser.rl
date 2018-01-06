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

    %%{
        machine parser;

        action mark { mark = fpc }
        action set_prefix { c.Prefix = string(data[mark:fpc]) }
        action set_prec   {
            if c.Prec, err = strconv.Atoi(string(data[mark:fpc])); err != nil {
                return c, err
            }
        }
        action set_op {
            if c.Op, ok = valToOp[string(data[mark:fpc])]; !ok {
                return c, fmt.Errorf("invalid op: %q", data[mark:fpc])
            }
        }
        action set_mode {
	    	if c.Mode, ok = valToMode[string(data[mark:fpc])]; !ok {
				return c, fmt.Errorf("invalid mode: %q", data[mark:fpc])
	    	}
        }
        action set_trap {
			c.Trap = ConditionFromString(string(data[mark:fpc]))
        }
        action add_input  { c.Inputs = append(c.Inputs, Data(data[mark:fpc])) }
        action set_output { c.Output = Data(data[mark:fpc]) }
        action set_excep  {
			c.Excep = ConditionFromString(string(data[mark:fpc]))
        }

       prefix = ('d' | 'b') >mark %set_prefix;
       prec = digit+ >mark %set_prec;
       op = (
              '+'      # Add
            | '-'      # Sub
            | '*'      # Mul
            | '/'      # Div
            | '*-'     # FMA
            | 'V'      # Sqrt
            | '%'      # Rem
            | 'rfi'    # RFI
            | 'cff'    # CFF
            | 'cfi'    # CFI
            | 'cif'    # CIF
            | 'cfd'    # CFD
            | 'cdf'    # CDF
            | 'qC'     # QuietCmp
            | 'sC'     # SigCmp
            | 'cp'     # Copy
            | '~'      # Neg
            | 'A'      # Abs
            | '@'      # CopySign
            | 'S'      # Scalb
            | 'L'      # Logb
            | 'Na'     # NextAfter
            | '?'      # Class
            | '?-'     # IsSigned
            | '?n'     # IsNormal
            | '?f'     # IsFinite
            | '?0'     # IsZero
            | '?s'     # IsSubnormal
            | '?i'     # IsInf
            | '?N'     # IsNaN
            | '?sN'    # IsSignaling
            | '<C'     # MinNum
            | '>C'     # MaxNum
            | '<A'     # MinNumMag
            | '>A'     # MaxNumMag
            | '=quant' # SameQuantum
            | 'quant'  # Quantize
            | 'Nu'     # NextUp
            | 'Nd'     # NextDown
            | 'eq'     # Equiv
			
			# Custom
			| 'rat'     # ToRat
			| 'sign'    # Sign
			| 'signbit' # Signbit
			| 'exp'     # Exponential function
			| 'log'     # Natural logarithm
			| 'log10'   # Common logarithm
			| 'pow'     # Power
            | '//'      # Int Div
			| 'norm'    # Normalize (reduce)
			| 'rtie'    # RoundToInt (exact)
			| 'shift'   # Shift
        ) >mark %set_op; 
        mode = (
              '>'  # ToPositiveInf
            | '<'  # ToNegativeInf
            | '0'  # ToZero
            | '=0' # ToNearestEven
            | '=^' # ToNearestAway
			| '^'  # AwayFromZero
        ) >mark %set_mode;
        condition = (
              'x' # Inexact
            | 'u' # Underflow
            | 'v' # Underflow
            | 'w' # Underflow
            | 'o' # Overflow
            | 'z' # DivisionByZero
            | 'i' # InvalidOperation

			# Custom
			| 'c' # Clamped
			| 'r' # Rounded
			| 'y' # ConversionSyntax
			| 'm' # DivisionImpossible
			| 'n' # DivisionUndefined
			| 't' # InsufficientStorage
			| '?' # InvalidContext
			| 's' # Subnormal
        );
        trap = condition* >mark %set_trap;
        excep = condition* >mark %set_excep;

		sign       = '+' | '-';
		indicator  = 'e' | 'E';
		exponent   = indicator? sign? digit+;
        number     = digit+ ('.' digit+)? exponent?;
		nan_prefix = [sSqQ];
		nan        = (nan_prefix | nan_prefix? 'nan'i);
		class      = (nan_prefix? 'nan'i | (sign? 
				  'Subnormal'
				| 'Normal'
				| 'Zero'
				| 'Infinity')
		);
        numeric_string = sign? (
			  ('true'i | 'false'i) # True, truE, false, FalSe, ...
			| nan                  # S, Q, NaN, sNaN, ...
            | ('inf'i 'inity'i?)   # +inf, -inf, ...
            | number               # 10, 10.1, +0e-392, ...
			| class                # +Normal, -Zero, ...
        );
        input = numeric_string >mark %add_input;
        output = (numeric_string | '#') >mark %set_output;

        main := (
            prefix . prec . (op ' '+) # Prefix, prec, and op are one 'chunk'
            (mode ' '+)               # Mode is its own chunk
            (trap ' '+)?              # Trap is its own chunk, maybe exists
            (input ' '+)+             # Input is one or more chunks
            '->' ' '+                 #
            output                    # Output is its own chunk 
            (' '+ excep)?             # Excep is its own chunk, maybe exists
        );

        write data;
        write init;
        write exec;
    }%%
    return c, nil
}
