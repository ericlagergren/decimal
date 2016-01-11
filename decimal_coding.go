package decimal

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
)

// Per version 1 we
const gobVersion byte = 1

// GobEncode implements the gob.GobEncoder interface.
func (d *Decimal) GobEncode() ([]byte, error) {
	if d == nil {
		return nil, nil
	}
	gob, err := d.mantissa.GobEncode()
	if err != nil {
		return nil, err
	}
	// 9 == gobVersion + sizeof(int64)
	b := make([]byte, len(gob)+9)
	b[0] = gobVersion
	binary.BigEndian.PutUint64(b[1:], uint64(d.scale))
	copy(b[9:], gob)
	return b, nil
}

// GobDecode implements the gob.GobDecoder interface.
func (d *Decimal) GobDecode(buf []byte) error {
	if len(buf) == 0 {
		// Other side sent a nil or default value.
		*d = Decimal{}
		return nil
	}
	if buf[0] != gobVersion {
		return fmt.Errorf("decimal.GobDecode: encoding version %d not supported", buf[0])
	}
	d.scale = int64(binary.BigEndian.Uint64(buf))
	return d.mantissa.GobDecode(buf[9:])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decimal) UnmarshalJSON(decimalBytes []byte) error {
	str, err := unquoteIfQuoted(decimalBytes)
	if err != nil {
		return fmt.Errorf("decimal.UnmarshalJSON: error decoding string %q: %s", decimalBytes, err)
	}

	decimal, err := NewFromString(str)
	if err != nil {
		return fmt.Errorf("decimal.UnmarshalJSON: error decoding string %q: %s", str, err)
	}
	*d = *decimal
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d *Decimal) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (d *Decimal) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("decimal.Scan: expected []byte, not %T", value)
	}

	str, err := unquoteIfQuoted(bytes)
	if err != nil {
		return err
	}
	d2, err := NewFromString(str)
	if err != nil {
		return err
	}
	*d = *d2
	return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (d *Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for XML
// deserialization.
func (d *Decimal) UnmarshalText(text []byte) error {
	dec, err := NewFromString(string(text))
	if err != nil {
		return fmt.Errorf("decimal.UnmarshalText: error decoding %q: %s", text, err)
	}
	*d = *dec
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface for XML
// serialization.
func (d *Decimal) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

func unquoteIfQuoted(bytes []byte) (string, error) {
	// If the amount is quoted, strip the quotes.
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}
