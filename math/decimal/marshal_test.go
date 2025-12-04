package decimal

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

type marshalTest struct {
	Value Decimal
}

func TestString(t *testing.T) {
	d := Decimal{}
	require.Equal(t, "<nil>", d.String())

	d = New(0)
	require.Equal(t, "0", d.String())

	d = NewFromBigIntWithPrec(big.NewInt(1000), 18)
	require.Equal(t, "0.000000000000001000", d.String())
}

func TestMarshalBinary(t *testing.T) {
	d := Decimal{}
	b, err := d.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, []byte(nil), b)

	require.NoError(t, d.UnmarshalBinary([]byte(nil)))
	require.True(t, d.i != nil)
}

func TestJSON(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		val := marshalTest{}
		bz, err := json.Marshal(val)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		if `{"Value":null}` != string(bz) {
			t.Errorf("Expected %v, got %v", "{}", string(bz))
			return
		}
		newVal := marshalTest{}
		err = json.Unmarshal(bz, &newVal)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		if !newVal.Value.IsNil() {
			t.Errorf("Expected %v, got %v", "nil value", newVal.Value.String())
			return
		}
	})

	t.Run("not zero value", func(t *testing.T) {
		val := marshalTest{Value: NewDecimalWithPrec(10001, 4)}
		bz, err := json.Marshal(val)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		if `{"Value":"1.0001"}` != string(bz) {
			t.Errorf("Expected %v, got %v", "{}", string(bz))
			return
		}
		newVal := marshalTest{}
		err = json.Unmarshal(bz, &newVal)
		if err != nil {
			t.Errorf("%v", err)
			return
		}
		if newVal.Value.String() != val.Value.String() {
			t.Errorf("Expected %v, got %v", val.Value.String(), newVal.Value.String())
			return
		}
	})
}

func TestDecimal_UnmarshalJSON(t *testing.T) {
	val := New(0)
	testCase := []struct {
		input  string
		wanted string
	}{
		{input: "null", wanted: "0"},
		{input: `0.123456789`, wanted: "0.123456789"},
		{input: `"0.123456789"`, wanted: "0.123456789"},
		{input: `"-0.123456789"`, wanted: "-0.123456789"},
		{input: `-0.123456789`, wanted: "-0.123456789"},
		{input: `999999999999999999999999999999999999999999`, wanted: "999999999999999999999999999999999999999999"},
		{input: `"999999999999999999999999999999999999999999"`, wanted: "999999999999999999999999999999999999999999"},
	}
	for i := 0; i < len(testCase); i++ {
		err := json.Unmarshal([]byte(testCase[i].input), &val)
		if err != nil {
			t.Error(err)
		}
		if testCase[i].wanted != val.String() {
			t.Errorf("unmarshal decimal error, wanted: %s, got: %s", testCase[i].wanted, val.String())
		}
	}
}
