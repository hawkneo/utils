package bigint

import (
	"fmt"
	"testing"

	"github.com/hawkneo/utils/math"
	"github.com/stretchr/testify/assert"
)

func TestNewFromString(t *testing.T) {
	tests := []struct {
		input string
		want  BigInt
	}{
		{
			input: "0",
			want:  NewFromInt(0),
		},
		{
			input: "-0",
			want:  NewFromInt(0),
		},
		{
			input: "10",
			want:  NewFromInt(10),
		},
		{
			input: "0b1",
			want:  NewFromInt(1),
		},
		{
			input: "0b11",
			want:  NewFromInt(3),
		},
		{
			input: "0B1",
			want:  NewFromInt(1),
		},
		{
			input: "0x1",
			want:  NewFromInt(1),
		},
		{
			input: "0X1",
			want:  NewFromInt(1),
		},
		{
			input: "0x10",
			want:  NewFromInt(16),
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("expect %s to equal %s", test.input, test.want.String()), func(t *testing.T) {
			got, ok := NewFromString(test.input)
			if !ok {
				t.Fatalf("got %v, want %v", got, test.want)
			}
			if !got.Equal(test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestBigInt_QuoDown(t *testing.T) {
	tests := []struct {
		x    BigInt
		y    BigInt
		want BigInt
	}{
		{
			x:    NewFromInt(0),
			y:    NewFromInt(1),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(1),
			y:    NewFromInt(1),
			want: NewFromInt(1),
		},
		{
			x:    NewFromInt(1230),
			y:    NewFromInt(10),
			want: NewFromInt(123),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(10),
			want: NewFromInt(123),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(100),
			want: NewFromInt(12),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(1000),
			want: NewFromInt(1),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(10000),
			want: NewFromInt(0),
		},

		{
			x:    NewFromInt(-0),
			y:    NewFromInt(1),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(-1),
			y:    NewFromInt(1),
			want: NewFromInt(-1),
		},
		{
			x:    NewFromInt(-1230),
			y:    NewFromInt(10),
			want: NewFromInt(-123),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(10),
			want: NewFromInt(-123),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(100),
			want: NewFromInt(-12),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(1000),
			want: NewFromInt(-1),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(10000),
			want: NewFromInt(-0),
		},
		{
			x:    NewFromInt(5),
			y:    NewFromInt(2),
			want: NewFromInt(2),
		},
		{
			x:    NewFromInt(-5),
			y:    NewFromInt(-2),
			want: NewFromInt(2),
		},
		{
			x:    NewFromInt(5),
			y:    NewFromInt(-2),
			want: NewFromInt(-2),
		},
		{
			x:    NewFromInt(-5),
			y:    NewFromInt(2),
			want: NewFromInt(-2),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf(
			"expect %s / %s to equal %s", test.x.String(), test.y.String(), test.want.String()),
			func(t *testing.T) {
				got := test.x.QuoDown(test.y)
				if !got.Equal(test.want) {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			},
		)
	}
}

func TestBigInt_QuoUp(t *testing.T) {
	tests := []struct {
		x    BigInt
		y    BigInt
		want BigInt
	}{
		{
			x:    NewFromInt(0),
			y:    NewFromInt(1),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(1),
			y:    NewFromInt(1),
			want: NewFromInt(1),
		},
		{
			x:    NewFromInt(1230),
			y:    NewFromInt(10),
			want: NewFromInt(123),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(10),
			want: NewFromInt(124),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(100),
			want: NewFromInt(13),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(1000),
			want: NewFromInt(2),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(10000),
			want: NewFromInt(1),
		},

		{
			x:    NewFromInt(-0),
			y:    NewFromInt(1),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(-1),
			y:    NewFromInt(1),
			want: NewFromInt(-1),
		},
		{
			x:    NewFromInt(-1230),
			y:    NewFromInt(10),
			want: NewFromInt(-123),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(10),
			want: NewFromInt(-124),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(100),
			want: NewFromInt(-13),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(1000),
			want: NewFromInt(-2),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(10000),
			want: NewFromInt(-1),
		},
		{
			x:    NewFromInt(5),
			y:    NewFromInt(2),
			want: NewFromInt(3),
		},
		{
			x:    NewFromInt(-5),
			y:    NewFromInt(-2),
			want: NewFromInt(3),
		},
		{
			x:    NewFromInt(5),
			y:    NewFromInt(-2),
			want: NewFromInt(-3),
		},
		{
			x:    NewFromInt(-5),
			y:    NewFromInt(2),
			want: NewFromInt(-3),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf(
			"expect %s / %s to equal %s", test.x.String(), test.y.String(), test.want.String()),
			func(t *testing.T) {
				got := test.x.Quo(test.y, math.RoundUp)
				if !got.Equal(test.want) {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			},
		)
	}
}

func TestBigInt_QuoCeiling(t *testing.T) {
	tests := []struct {
		x    BigInt
		y    BigInt
		want BigInt
	}{
		{
			x:    NewFromInt(0),
			y:    NewFromInt(1),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(1),
			y:    NewFromInt(1),
			want: NewFromInt(1),
		},
		{
			x:    NewFromInt(1230),
			y:    NewFromInt(10),
			want: NewFromInt(123),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(10),
			want: NewFromInt(124),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(100),
			want: NewFromInt(13),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(1000),
			want: NewFromInt(2),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(10000),
			want: NewFromInt(1),
		},

		{
			x:    NewFromInt(-0),
			y:    NewFromInt(1),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(-1),
			y:    NewFromInt(1),
			want: NewFromInt(-1),
		},
		{
			x:    NewFromInt(-1230),
			y:    NewFromInt(10),
			want: NewFromInt(-123),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(10),
			want: NewFromInt(-123),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(100),
			want: NewFromInt(-12),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(1000),
			want: NewFromInt(-1),
		},
		{
			x:    NewFromInt(-1234),
			y:    NewFromInt(10000),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(1234),
			y:    NewFromInt(-10000),
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(5),
			y:    NewFromInt(2),
			want: NewFromInt(3),
		},
		{
			x:    NewFromInt(-5),
			y:    NewFromInt(-2),
			want: NewFromInt(3),
		},
		{
			x:    NewFromInt(5),
			y:    NewFromInt(-2),
			want: NewFromInt(-2),
		},
		{
			x:    NewFromInt(-5),
			y:    NewFromInt(2),
			want: NewFromInt(-2),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf(
			"expect %s / %s to equal %s", test.x.String(), test.y.String(), test.want.String()),
			func(t *testing.T) {
				got := test.x.Quo(test.y, math.RoundCeiling)
				if !got.Equal(test.want) {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			},
		)
	}
}

func TestBigInt_Mod(t *testing.T) {
	assert.True(t, NewFromInt(1).Mod(NewFromInt64(2)).Equal(NewFromInt64(1)))
}

func TestBigInt_Power(t *testing.T) {
	tests := []struct {
		x    BigInt
		y    int64
		want BigInt
	}{
		{
			x:    NewFromInt(0),
			y:    1,
			want: NewFromInt(0),
		},
		{
			x:    NewFromInt(1),
			y:    1,
			want: NewFromInt(1),
		},
		{
			x:    NewFromInt(2),
			y:    1,
			want: NewFromInt(2),
		},
		{
			x:    NewFromInt(2),
			y:    2,
			want: NewFromInt(4),
		},
		{
			x:    NewFromInt(3),
			y:    3,
			want: NewFromInt(27),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf(
			"expect %s^%d to equal %s", test.x.String(), test.y, test.want.String()),
			func(t *testing.T) {
				got := test.x.Power(test.y)
				if !got.Equal(test.want) {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			},
		)
	}
}
