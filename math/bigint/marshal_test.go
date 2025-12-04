package bigint

import (
	"encoding/json"
	"testing"
)

func TestBigInt_UnmarshalJSON(t *testing.T) {
	val := NewFromInt(0)
	testCase := []struct {
		input  string
		wanted string
	}{
		{input: "null", wanted: "0"},
		{input: `123456789`, wanted: "123456789"},
		{input: `"123456789"`, wanted: "123456789"},
		{input: `"-123456789"`, wanted: "-123456789"},
		{input: `-123456789`, wanted: "-123456789"},
		{input: `999999999999999999999999999999999999999999`, wanted: "999999999999999999999999999999999999999999"},
		{input: `"999999999999999999999999999999999999999999"`, wanted: "999999999999999999999999999999999999999999"},
	}
	for i := 0; i < len(testCase); i++ {
		err := json.Unmarshal([]byte(testCase[i].input), &val)
		if err != nil {
			t.Error(err)
		}
		if testCase[i].wanted != val.String() {
			t.Errorf("unmarshal bigint error, wanted: %s, got: %s", testCase[i].wanted, val.String())
		}
	}
}
