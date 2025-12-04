package bigint

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/gridexswap/utils/marshal"
	"math/big"
)

// String implements the fmt.Stringer interface
func (b BigInt) String() string {
	if b.i == nil {
		return "<nil>"
	}
	return b.i.String()
}

// MarshalJSON implements the json.Marshaler interface
func (b BigInt) MarshalJSON() ([]byte, error) {
	if b.i == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(b.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (b *BigInt) UnmarshalJSON(bz []byte) error {
	if len(bz) == len("null") && string(bz) == "null" {
		return nil
	}

	if b.i == nil {
		b.i = new(big.Int)
	}

	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			bTemp, ok := NewFromString(string(bz))
			if ok {
				*b = bTemp
				return nil
			}
		}
		return err
	}

	bTemp, ok := NewFromString(text)
	if !ok {
		return fmt.Errorf("invalid string: %s", text)
	}
	*b = bTemp
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface
func (b BigInt) MarshalYAML() (any, error) {
	return b.String(), nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (b BigInt) MarshalBinary() (data []byte, err error) {
	if b.i == nil {
		return nil, nil
	}
	return b.i.GobEncode()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (b *BigInt) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		// Other side sent a nil or default value.
		b.i = new(big.Int)
		return nil
	}
	if b.i == nil {
		b.i = new(big.Int)
	}
	return b.i.GobDecode(data)
}

// Value implements the driver.Valuer interface
func (b BigInt) Value() (driver.Value, error) {
	return b.String(), nil
}

// Scan implements the sql.Scanner interface
func (b *BigInt) Scan(value any) error {
	switch v := value.(type) {
	case int64:
		b.i = big.NewInt(v)
	case uint64:
		b.i = new(big.Int).SetUint64(v)
	default:
		text, err := marshal.UnquoteIfQuoted(v)
		if err != nil {
			return err
		}

		bTemp, ok := NewFromString(text)
		if !ok {
			return fmt.Errorf("invalid string: %s", text)
		}
		*b = bTemp
	}
	return nil
}

// Marshal implements the gogo proto custom type interface
func (b BigInt) Marshal() ([]byte, error) {
	return b.MarshalBinary()
}

// MarshalTo implements the gogo proto custom type interface
func (b BigInt) MarshalTo(data []byte) (n int, err error) {
	bz, err := b.MarshalBinary()
	if err != nil {
		return
	}
	n = copy(data, bz)
	return
}

// Unmarshal implements the gogo proto custom type interface
func (b *BigInt) Unmarshal(data []byte) error {
	return b.UnmarshalBinary(data)
}

// Size implements the gogo proto custom type interface
func (b BigInt) Size() int {
	bz, _ := b.Marshal()
	return len(bz)
}
