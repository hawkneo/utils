package decimal

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gridexswap/utils/marshal"
	"math/big"
)

const (
	PrecisionFixedSize = 4
)

// same as cosmos-sdk Decimal
func (d Decimal) String() string {
	if d.i == nil {
		return "<nil>"
	}
	if d.prec == 0 {
		return d.i.String()
	}

	isNeg := d.IsNegative()

	if isNeg {
		d = d.Neg()
	}

	bzInt, err := d.i.MarshalText()
	if err != nil {
		return ""
	}
	inputSize := len(bzInt)

	var bzStr []byte

	// TODO: Remove trailing zeros
	// case 1, purely decimal
	if inputSize <= d.prec {
		bzStr = make([]byte, d.prec+2)

		// 0. prefix
		bzStr[0] = byte('0')
		bzStr[1] = byte('.')

		// set relevant digits to 0
		for i := 0; i < d.prec-inputSize; i++ {
			bzStr[i+2] = byte('0')
		}

		// set final digits
		copy(bzStr[2+(d.prec-inputSize):], bzInt)
	} else {
		// inputSize + 1 to account for the decimal point that is being added
		bzStr = make([]byte, inputSize+1)
		decPointPlace := inputSize - d.prec

		copy(bzStr, bzInt[:decPointPlace])                   // pre-decimal digits
		bzStr[decPointPlace] = byte('.')                     // decimal point
		copy(bzStr[decPointPlace+1:], bzInt[decPointPlace:]) // post-decimal digits
	}

	if isNeg {
		return "-" + string(bzStr)
	}

	return string(bzStr)
}

// MarshalJSON implements json.Marshaler
func (d Decimal) MarshalJSON() ([]byte, error) {
	if d.i == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Decimal) UnmarshalJSON(bz []byte) error {
	if len(bz) == len("null") && string(bz) == "null" {
		return nil
	}

	if d.i == nil {
		d.i = new(big.Int)
	}

	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			dTemp, err := NewFromString(string(bz))
			if err == nil {
				*d = dTemp
				return nil
			}
		}
		return err
	}

	newDec, err := NewDecimalFromString(text)
	if err != nil {
		return err
	}

	*d = newDec
	return nil
}

// MarshalYAML implements yaml.Marshaler
func (d Decimal) MarshalYAML() (any, error) {
	return d.String(), nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface
func (d Decimal) MarshalBinary() (data []byte, err error) {
	if d.i == nil {
		return nil, nil
	}
	// Write the precision as fixed-width bytes.
	precBytes := make([]byte, PrecisionFixedSize)
	binary.BigEndian.PutUint32(precBytes, uint32(d.prec))

	var intBytes []byte
	if intBytes, err = d.i.GobEncode(); err != nil {
		return nil, err
	}

	data = append(precBytes, intBytes...)
	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface.
func (d *Decimal) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		d.i = &big.Int{}
		return nil
	}

	if len(data) < PrecisionFixedSize {
		return fmt.Errorf("error decoding binary %v: expected at least %d bytes, got %v",
			data, PrecisionFixedSize, len(data))
	}

	// Read the precision as fixed-width bytes.
	d.prec = int(binary.BigEndian.Uint32(data[:PrecisionFixedSize]))

	// Read the big.Int.
	d.i = new(big.Int)
	return d.i.GobDecode(data[PrecisionFixedSize:])
}

// Value implements driver.Valuer interface for database serialization.
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// Scan implements sql.Scanner interface for database deserialization.
func (d *Decimal) Scan(value any) error {
	// first try to see if the data is stored in database as a Numeric datatype
	switch v := value.(type) {

	case float32:
		*d = NewDecimalFromFloat64(float64(v))
		return nil

	case float64:
		// numeric in sqlite3 sends us float64
		*d = NewDecimalFromFloat64(v)
		return nil

	case int64:
		// at least in sqlite3 when the value is 0 in db, the data is sent
		// to us as an int64 instead of a float64 ...
		*d = NewDecimal(v)
		return nil

	default:
		// default is trying to interpret value stored as string
		text, err := marshal.UnquoteIfQuoted(v)
		if err != nil {
			return err
		}
		bTemp, err := NewDecimalFromString(text)
		if err != nil {
			return err
		}
		*d = bTemp
		return nil
	}
}

// Marshal implements the gogo proto custom type interface.
func (d Decimal) Marshal() ([]byte, error) {
	return d.MarshalBinary()
}

// MarshalTo implements the gogo proto custom type interface.
func (d Decimal) MarshalTo(data []byte) (n int, err error) {
	bz, err := d.MarshalBinary()
	if err != nil {
		return
	}
	n = copy(data, bz)
	return
}

// Unmarshal implements the gogo proto custom type interface.
func (d *Decimal) Unmarshal(data []byte) error {
	return d.UnmarshalBinary(data)
}

// Size implements the gogo proto custom type interface.
func (d Decimal) Size() int {
	bz, _ := d.Marshal()
	return len(bz)
}

// MarshalAmino Override Amino binary serialization by proxying to protobuf.
func (d Decimal) MarshalAmino() ([]byte, error) {
	return d.Marshal()
}

// UnmarshalAmino Override Amino binary serialization by proxying to protobuf.
func (d Decimal) UnmarshalAmino(bz []byte) error {
	return d.Unmarshal(bz)
}
