// Package postgres provides a simple wrapper around a decimal.Big type, allowing
// it to be used in PostgreSQL queries. It ensures the decimal fits inside the
// limits of the DECIMAL type.
package postgres

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/ericlagergren/decimal"
)

const (
	MaxIntegralDigits   = 131072 // max digits before the decimal point
	MaxFractionalDigits = 16383  // max digits after the decimal point
)

// LengthError is returned from Decimal.Value when either its integral (digits
// before the decimal point) or fractional (digits after the decimal point)
// parts are too long for PostgresSQL.
type LengthError struct {
	Part string // "integral" or "fractional"
	N    int    // length of invalid part
	max  int
}

func (e LengthError) Error() string {
	return fmt.Sprintf("%s (%d digits) is too long (%d max)", e.Part, e.N, e.max)
}

// Decimal is a PostgreSQL DECIMAL. Its zero value is valid for use with both
// Value and Scan.
type Decimal struct {
	V     *decimal.Big
	Round bool // round if the decimal exceeds the bounds for DECIMAL
	Zero  bool // return "0" if V == nil
}

// Value implements driver.Valuer.
func (d *Decimal) Value() (driver.Value, error) {
	if d.V == nil {
		if d.Zero {
			return "0", nil
		}
		return nil, nil
	}
	v := d.V
	if v.IsNaN(0) {
		return "NaN", nil
	}
	if v.IsInf(0) {
		return nil, errors.New("Decimal.Value: DECIMAL does not accept Infinities")
	}

	dl := v.Precision()  // length of d
	sl := int(v.Scale()) // length of fractional part

	if il := dl - sl; il > MaxIntegralDigits {
		if !d.Round {
			return nil, &LengthError{Part: "integral", N: il, max: MaxIntegralDigits}
		}
		// Rounding down the integral part automatically chops off the fractional
		// part.
		return v.Round(MaxIntegralDigits).String(), nil
	}
	if sl > MaxFractionalDigits {
		if !d.Round {
			return nil, &LengthError{Part: "fractional", N: sl, max: MaxFractionalDigits}
		}
		v.Round(dl - (sl - MaxFractionalDigits))
	}
	return v.String(), nil
}

// Scan implements sql.Scanner.
func (d *Decimal) Scan(val interface{}) error {
	if d.V == nil {
		d.V = new(decimal.Big)
	}
	switch t := val.(type) {
	case string:
		if _, ok := d.V.SetString(t); !ok {
			if err := d.V.Context.Err(); err != nil {
				return err
			}
			return fmt.Errorf("Decimal.Scan: invalid syntax: %q", t)
		}
		return nil
	case []byte:
		return d.V.UnmarshalText(t)
	default:
		return fmt.Errorf("Decimal.Scan: unknown value: %#v", val)
	}
}
