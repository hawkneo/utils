package decimal

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/gridexswap/utils/math"
)

func TestNewFromString(t *testing.T) {
	tests := []struct {
		input         string
		wantPrecision int
		wantNeg       bool
		wantError     bool
		want          string
		name          string
	}{
		{"", 0, false, true, "", "input is empty"},
		{"1.", 0, false, true, "", "fraction part is invalid"},
		{".1", 0, false, true, "", "int part is invalid"},
		{"-", 0, false, true, "", "input is invalid"},
		{"-1.", 0, false, true, "", "fraction part is invalid"},
		{"-.1", 0, false, true, "", "int part is invalid"},
		{fmt.Sprintf("0.%0129d", 0), 0, false, true, "", "precision(129) overflow"},
		//{fmt.Sprintf("%s.%0128d", MaxUint256, 0), 0, false, true, "", "bit len overflow"},

		{"1", 0, false, false, "1", "valid"},
		{"0", 0, false, false, "0", "valid"},
		{"-1", 0, true, false, "-1", "valid"},

		{"1.0001", 4, false, false, "1.0001", "valid"},
		{"-1.0001", 4, true, false, "-1.0001", "valid"},
		{"3.7154500000000011e-15", 31, false, false, "0.0000000000000037154500000000011", "valid"},
		{"-3.7154500000000011e-15", 31, true, false, "-0.0000000000000037154500000000011", "valid"},
		{"3.7154e3", 1, false, false, "3715.4", "valid"},
		{"-3.7154e3", 1, true, false, "-3715.4", "valid"},
		{"3.7154e5", 0, false, false, "371540", "valid"},
		{"-3.7154e5", 0, true, false, "-371540", "valid"},
		{fmt.Sprintf("%s.%0128d", MaxUint128, 0), 128, false, false, fmt.Sprintf("%s.%0128d", MaxUint128, 0), "precision(128)"},
		{fmt.Sprintf("-%s.%0128d", MaxUint128, 0), 128, true, false, fmt.Sprintf("-%s.%0128d", MaxUint128, 0), "precision(128)"},
		{fmt.Sprintf("-%s", MaxUint256), 0, true, false, fmt.Sprintf("-%s", MaxUint256), "max uint256"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dec, err := NewFromString(test.input)
			if test.wantError {
				if err == nil {
					t.Errorf("NewFromString(%q) = %v, want error", test.input, dec)
				}
				return
			} else {
				if err != nil {
					t.Errorf("NewFromString(%q) = %v, want %v", test.input, err, test.want)
					return
				}
			}
			if dec.Precision() != test.wantPrecision {
				t.Errorf("NewFromString(%q) = %v, want %v", test.input, dec.Precision(), test.wantPrecision)
				return
			}
			if dec.IsNegative() != test.wantNeg {
				t.Errorf("NewFromString(%q) = %v, want %v", test.input, dec.IsNegative(), test.wantNeg)
				return
			}
			if dec.String() != test.want {
				t.Errorf("NewFromString(%q) = %v, want %v", test.input, dec.String(), test.want)
				return
			}
		})
	}
}

func TestNewFromFloat64(t *testing.T) {
	tests := []struct {
		value float64
		want  Decimal
		name  string
	}{
		{0, New(0), "zero"},
		{1, New(1), "one"},
		{1.1, NewWithPrec(11, 1), "1.1"},
		{1.01, NewWithPrec(101, 2), "1.01"},
		{1.001, NewWithPrec(1001, 3), "1.001"},
		{1.000000001, NewWithPrec(1000000001, 9), "1.000000001"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := NewFromFloat64(test.value)
			if !val.Equal(test.want) {
				t.Fatalf("NewFromFloat64(%v) = %v, want %v", test.value, val, test.want)
			}
		})
	}
}

func TestNewWithPrec(t *testing.T) {
	val := NewWithPrec(0, 18)
	if val.Precision() != 18 {
		t.Fatalf("NewWithPrec(0, 18) = %v, want %v", val.Precision(), 18)
		return
	}
	if val.Sign() != 0 {
		t.Fatalf("NewWithPrec(0, 18) = %v, want %v", val.Sign(), 0)
		return
	}
	if val.String() != "0.000000000000000000" {
		t.Fatalf("NewWithPrec(0, 18) = %v, want %v", val.String(), "0.000000000000000000")
		return
	}
}

func TestDecimal_Add(t *testing.T) {
	tests := []struct {
		value1     uint64
		precision1 int
		value2     uint64
		precision2 int
		want       Decimal
		name       string
	}{
		{1, 0, 50, 1, NewFromUint64(60, 1), "1+5=6"},
		{10, 1, 5, 1, NewFromUint64(15, 1), "1+0.5=1.5"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d1 := NewFromUint64(test.value1, test.precision1)
			d2 := NewFromUint64(test.value2, test.precision2)
			result := d1.Add(d2)
			if !result.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, result)
			}
		})
	}
}

func TestDecimal_SafeAdd(t *testing.T) {
	tests := []struct {
		value1     int64
		precision1 int
		value2     int64
		precision2 int
		wantPanic  bool
		want       Decimal
		name       string
	}{
		{1, 0, 50, 1, false, NewFromUint64(60, 1), "1+5=6"},
		{1, 0, -10, 1, false, NewFromUint64(0, 0), "1+(-1)=0"},
		{1, 0, -20, 1, true, New(0), "1+(-2)=panic"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if test.wantPanic && r == nil {
					t.Fatalf("expected panic, got nil")
				}
				if !test.wantPanic && r != nil {
					t.Fatalf("expected nil, got %v", r)
				}
			}()

			d1 := NewFromInt64(test.value1, test.precision1)
			d2 := NewFromInt64(test.value2, test.precision2)
			result := d1.SafeAdd(d2)
			if !test.wantPanic && !result.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, result)
			}
		})
	}

	t.Run("max uint256 + 1 should not overflow", func(t *testing.T) {
		maxUint256Dec := NewFromBigInt(MaxUint256)
		maxUint256Dec = maxUint256Dec.AddRaw(1)
		if maxUint256Dec.IsNegative() {
			t.Errorf("max uint256 + 1 should not overflow")
		}
	})
}

func TestDecimal_UnsignedAdd(t *testing.T) {
	tests := []struct {
		value1     int64
		precision1 int
		value2     int64
		precision2 int
		want       Decimal
		name       string
	}{
		{1, 0, -1, 0, MustFromString("0"), "1+(-1)"},
		{1, 0, 0, 0, MustFromString("1"), "1+0"},
		{0, 0, -5, 1, MustFromString("11579208923731619542357098500868790785326998466564056403945758400791312963993.1"), "0+(-0.5)"},
		{1, 0, -2, 0, MustFromString("115792089237316195423570985008687907853269984665640564039457584007913129639935"), "1+(-2)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d1 := NewFromInt64(test.value1, test.precision1)
			d2 := NewFromInt64(test.value2, test.precision2)
			result := d1.UnsignedAdd(d2, Uint256BitLen)
			if !result.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, result)
			}
		})
	}

	t.Run("uint256 overflow", func(t *testing.T) {
		max := NewFromBigInt(new(big.Int).Set(MaxUint256))
		overflow := max.UnsignedAdd(New(1), Uint256BitLen)
		if !overflow.Equal(New(0)) {
			t.Fatalf("expected 0, got %s", overflow)
		}
	})

	t.Run("uint256 underflow", func(t *testing.T) {
		overflow := New(0).UnsignedAdd(New(-1), Uint256BitLen)
		if !overflow.Equal(MustFromString("115792089237316195423570985008687907853269984665640564039457584007913129639935")) {
			t.Fatalf("expected 115792089237316195423570985008687907853269984665640564039457584007913129639935, got %s", overflow)
		}
	})
}

func TestDecimal_SafeSub(t *testing.T) {
	tests := []struct {
		value1     int64
		precision1 int
		value2     int64
		precision2 int
		wantPanic  bool
		want       Decimal
		name       string
	}{
		{1, 0, 50, 2, false, NewFromUint64(5, 1), "1-0.5=0.5"},
		{1, 0, -10, 1, false, NewFromUint64(2, 0), "1-(-1)=2"},
		{1, 0, 20, 1, true, New(0), "1-2=panic"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if test.wantPanic && r == nil {
					t.Fatalf("expected panic, got nil")
				}
				if !test.wantPanic && r != nil {
					t.Fatalf("expected nil, got %v", r)
				}
			}()

			d1 := NewFromInt64(test.value1, test.precision1)
			d2 := NewFromInt64(test.value2, test.precision2)
			result := d1.SafeSub(d2)
			if !test.wantPanic && !result.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, result)
			}
		})
	}
}

func TestDecimal_UnsignedSub(t *testing.T) {
	tests := []struct {
		value1     int64
		precision1 int
		value2     int64
		precision2 int
		want       Decimal
		name       string
	}{
		{1, 0, 1, 0, MustFromString("0"), "1-1"},
		{1, 0, 0, 0, MustFromString("1"), "1-0"},
		{0, 0, 5, 1, MustFromString("11579208923731619542357098500868790785326998466564056403945758400791312963993.1"), "0-0.5"},
		{1, 0, 2, 0, MustFromString("115792089237316195423570985008687907853269984665640564039457584007913129639935"), "1-2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d1 := NewFromInt64(test.value1, test.precision1)
			d2 := NewFromInt64(test.value2, test.precision2)
			result := d1.UnsignedSub(d2, Uint256BitLen)
			if !result.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, result)
			}
		})
	}
}

func TestDecimal_MulDown(t *testing.T) {
	tests := []struct {
		value1        Decimal
		value2        Decimal
		wantPrecision int
		want          Decimal
		name          string
	}{
		{MustFromString("1"), MustFromString("0"), 0, MustFromString("0"), "1x0=0"},
		{MustFromString("1.000"), MustFromString("0"), 3, MustFromString("0"), "1.000x0=0.000"},
		{MustFromString("1.000"), MustFromString("1.000"), 3, MustFromString("1"), "1.000x1.000=1.000"},
		{MustFromString("1.111"), MustFromString("1.111"), 3, MustFromString("1.234"), "1.111x1.111=1.234"},
		{MustFromString("1.333"), MustFromString("1.333"), 3, MustFromString("1.776"), "1.333x1.333=1.776"},
		{MustFromString("1." + strings.Repeat("3", 128)), MustFromString("1." + strings.Repeat("3", 128)), 128, MustFromString("1." + strings.Repeat("7", 127) + "6"), "1.3(128)x1.3(128)=1.7(127)6"},

		{MustFromString("-1"), MustFromString("0"), 0, MustFromString("0"), "-1x0=0"},
		{MustFromString("-1.000"), MustFromString("0"), 3, MustFromString("0"), "-1.000x0=0.000"},
		{MustFromString("-1.000"), MustFromString("1.000"), 3, MustFromString("-1"), "-1.000x1.000=-1.000"},
		{MustFromString("-1.111"), MustFromString("1.111"), 3, MustFromString("-1.234"), "-1.111x1.111=-1.234"},
		{MustFromString("-1.333"), MustFromString("1.333"), 3, MustFromString("-1.776"), "-1.333x1.333=-1.776"},

		{MustFromString("-1.000"), MustFromString("-1.000"), 3, MustFromString("1"), "-1.000x-1.000=1.000"},
		{MustFromString("-1.111"), MustFromString("-1.111"), 3, MustFromString("1.234"), "-1.111x-1.111=1.234"},
		{MustFromString("-1.333"), MustFromString("-1.333"), 3, MustFromString("1.776"), "-1.333x-1.333=1.776"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := test.value1.MulDown(test.value2)
			if !val.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, val)
				return
			}
			if val.Precision() != test.wantPrecision {
				t.Fatalf("expected %d, got %d", test.wantPrecision, val.Precision())
				return
			}
		})
	}
}

func TestDecimal_MulRoundUp(t *testing.T) {
	tests := []struct {
		value1        Decimal
		value2        Decimal
		wantPrecision int
		want          Decimal
		name          string
	}{
		{MustFromString("1"), MustFromString("0"), 0, MustFromString("0"), "1x0=0"},
		{MustFromString("1.000"), MustFromString("0"), 3, MustFromString("0"), "1.000x0=0.000"},
		{MustFromString("1.000"), MustFromString("1.000"), 3, MustFromString("1"), "1.000x1.000=1.000"},
		{MustFromString("1.111"), MustFromString("1.111"), 3, MustFromString("1.235"), "1.111x1.111=1.235"},
		{MustFromString("1.333"), MustFromString("1.333"), 3, MustFromString("1.777"), "1.333x1.333=1.777"},
		{MustFromString("1." + strings.Repeat("3", 128)), MustFromString("1." + strings.Repeat("3", 128)), 128, MustFromString("1." + strings.Repeat("7", 128)), "1.3(128)x1.3(128)=1.7(128)"},

		{MustFromString("-1"), MustFromString("0"), 0, MustFromString("0"), "-1x0=0"},
		{MustFromString("-1.000"), MustFromString("0"), 3, MustFromString("0"), "-1.000x0=0.000"},
		{MustFromString("-1.000"), MustFromString("1.000"), 3, MustFromString("-1"), "-1.000x1.000=-1.000"},
		{MustFromString("-1.111"), MustFromString("1.111"), 3, MustFromString("-1.235"), "-1.111x1.111=-1.235"},
		{MustFromString("-1.333"), MustFromString("1.333"), 3, MustFromString("-1.777"), "-1.333x1.333=-1.777"},

		{MustFromString("-1.000"), MustFromString("-1.000"), 3, MustFromString("1"), "-1.000x-1.000=1.000"},
		{MustFromString("-1.111"), MustFromString("-1.111"), 3, MustFromString("1.235"), "-1.111x-1.111=1.235"},
		{MustFromString("-1.333"), MustFromString("-1.333"), 3, MustFromString("1.777"), "-1.333x-1.333=1.777"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := test.value1.Mul(test.value2, math.RoundUp)
			if !val.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, val)
				return
			}
			if val.Precision() != test.wantPrecision {
				t.Fatalf("expected %d, got %d", test.wantPrecision, val.Precision())
				return
			}
		})
	}
}

func TestDecimal_QuoRoundUp(t *testing.T) {
	tests := []struct {
		value1        Decimal
		value2        Decimal
		wantPrecision int
		want          Decimal
		name          string
	}{
		{MustFromString("5"), MustFromString("2"), 0, MustFromString("3"), "5/2=3"},
		{MustFromString("-5"), MustFromString("-2"), 0, MustFromString("3"), "-5/-2=3"},
		{MustFromString("5"), MustFromString("-2"), 0, MustFromString("-3"), "5/-2=-3"},
		{MustFromString("-5"), MustFromString("2"), 0, MustFromString("-3"), "-5/2=-3"},
		{MustFromString("55"), MustFromString("100.0"), 1, MustFromString("0.6"), "55/100.0=0.6"},
		{MustFromString("25"), MustFromString("100.0"), 1, MustFromString("0.3"), "25/100.0=0.3"},
		{MustFromString("16"), MustFromString("100.0"), 1, MustFromString("0.2"), "16/100.0=0.2"},
		{MustFromString("11"), MustFromString("100.0"), 1, MustFromString("0.2"), "11/100.0=0.2"},
		{MustFromString("10"), MustFromString("100.0"), 1, MustFromString("0.1"), "10/100.0=0.1"},

		{MustFromString("-55"), MustFromString("100.0"), 1, MustFromString("-0.6"), "-55/100.0=-0.6"},
		{MustFromString("-25"), MustFromString("100.0"), 1, MustFromString("-0.3"), "-25/100.0=-0.3"},
		{MustFromString("-16"), MustFromString("100.0"), 1, MustFromString("-0.2"), "-16/100.0=-0.2"},
		{MustFromString("-11"), MustFromString("100.0"), 1, MustFromString("-0.2"), "-11/100.0=-0.2"},
		{MustFromString("-10"), MustFromString("100.0"), 1, MustFromString("-0.1"), "-10/100.0=-0.1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := test.value1.Quo(test.value2, math.RoundUp)
			if !val.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, val)
				return
			}
			if val.Precision() != test.wantPrecision {
				t.Fatalf("expected %d, got %d", test.wantPrecision, val.Precision())
				return
			}
		})
	}
}

func TestDecimal_Power(t *testing.T) {
	tests := []struct {
		value         Decimal
		power         int64
		wantPrecision int
		want          Decimal
		name          string
	}{
		{MustFromString("2"), 0, 0, MustFromString("1"), "2^0=1"},
		{MustFromString("2"), 1, 0, MustFromString("2"), "2^1=2"},
		{MustFromString("2"), 2, 0, MustFromString("4"), "2^2=4"},
		{MustFromString("2"), 3, 0, MustFromString("8"), "2^3=8"},
		{MustFromString("2"), 4, 0, MustFromString("16"), "2^4=16"},
		{MustFromString("2"), 5, 0, MustFromString("32"), "2^5=32"},
		{MustFromString("2"), 6, 0, MustFromString("64"), "2^6=64"},
		{MustFromString("2"), 7, 0, MustFromString("128"), "2^7=128"},
		{MustFromString("2"), 8, 0, MustFromString("256"), "2^8=256"},

		{MustFromString("-2"), 1, 0, MustFromString("-2"), "-2^1=-2"},
		{MustFromString("-2"), 2, 0, MustFromString("4"), "-2^2=4"},
		{MustFromString("-2"), 3, 0, MustFromString("-8"), "-2^3=-8"},

		{NewWithPrec(1414213562373095049, 18), 2, 18, MustFromString("2.000000000000000001"), "1.414213562373095049^2=2.000000000000000001"},

		{NewWithAppendPrec(2, 18), -1, 18, MustFromString("0.50000000000000000"), "2.000000000000000000^-1=0.50000000000000000"},
		{NewWithAppendPrec(2, 18), -2, 18, MustFromString("0.25000000000000000"), "2.000000000000000000^-2=0.25000000000000000"},
		{NewWithAppendPrec(2, 1), -2, 1, MustFromString("0.3"), "2.0^-2=0.3"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := test.value.Power(test.power)
			if !val.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, val)
				return
			}
			if val.Precision() != test.wantPrecision {
				t.Fatalf("expected %d, got %d", test.wantPrecision, val.Precision())
				return
			}
		})
	}
}

func TestDecimal_Sqrt(t *testing.T) {
	tests := []struct {
		value         Decimal
		wantPrecision int
		want          Decimal
		name          string
	}{
		{MustFromString("4"), 0, MustFromString("1"), "sqrt(4)=1"},
		{MustFromString("4.0000"), 4, MustFromString("2"), "sqrt(4.0000)=2"},
		{NewWithPrec(25, 2), 2, NewWithPrec(5, 1), "sqrt(0.25)=0.50"},
		{NewWithAppendPrec(2, 18), 18, NewWithPrec(1414213562373095049, 18), "sqrt(2)=1.414213562373095049"},
		{NewWithPrec(1, 18), 18, NewWithPrec(1, 9), "sqrt(0.000000000000000001)=0.000000001"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := test.value.Sqrt()
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
				return
			}
			if !val.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, val)
				return
			}
			if val.Precision() != test.wantPrecision {
				t.Fatalf("expected %d, got %d", test.wantPrecision, val.Precision())
				return
			}
		})
	}
}

func TestDecimal_ApproxRoot(t *testing.T) {
	tests := []struct {
		value         Decimal
		wantPrecision int
		want          Decimal
		name          string
	}{
		{MustFromString("3125.0000"), 4, MustFromString("5.0000"), "root(3125.0000)=5"},
		{MustFromString("100000.0000"), 4, MustFromString("10.0000"), "root(100000.0000)=10"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := test.value.ApproxRoot(5)
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
				return
			}
			if !val.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, val)
				return
			}
			if val.Precision() != test.wantPrecision {
				t.Fatalf("expected %d, got %d", test.wantPrecision, val.Precision())
				return
			}
		})
	}
}

func TestDecimal_Logn(t *testing.T) {
	tests := []struct {
		value         Decimal
		wantPrecision int
		want          Decimal
		name          string
	}{
		{NewWithAppendPrec(1, 18), 18, New(0), "log_2(1)=0"},
		{NewWithAppendPrec(2, 18), 18, New(1), "log_2(2)=1"},
		{NewWithAppendPrec(4, 18), 18, New(2), "log_2(4)=2"},
		{NewWithAppendPrec(8, 18), 18, New(3), "log_2(8)=3"},
		{NewWithAppendPrec(16, 18), 18, New(4), "log_2(16)=4"},

		{NewWithAppendPrec(33, 18), 18, MustFromString("5.044394119358453436"), "log_2(33)=5.044394119358453436"},
		{NewWithAppendPrec(63, 18), 18, MustFromString("5.977279923499916469"), "log_2(63)=5.977279923499916469"},

		{MustFromString("2.12345678"), 8, MustFromString("1.08641474"), "log_2(2.12345678)=1.08641474"},
		{MustFromString("1.12345678"), 8, MustFromString("0.16794462"), "log_2(1.12345678)=0.0.16794462"},
		{NewWithPrec(200000000000000000, 18), 18, MustFromString("-2.321928094887362348"), "log_2(0.2)=-2.321928094887362348"},
		{NewWithPrec(2, 18), 18, MustFromString("-58.794705707972522263"), "log_2(0.000000000000000002)=-58.794705707972522263"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := test.value.Log2()
			if !val.Equal(test.want) {
				t.Fatalf("expected %s, got %s", test.want, val)
				return
			}
			if val.Precision() != test.wantPrecision {
				t.Fatalf("expected %d, got %d", test.wantPrecision, val.Precision())
				return
			}
		})
	}
}

func TestDeciaml_Rescale(t *testing.T) {
	tests := []struct {
		value         Decimal
		roundingMode  math.RoundingMode
		rescaleValue  int
		wantPrecision int
		want          Decimal
		name          string
	}{
		{New(1), math.RoundDown, 0, 0, New(1), "1 set precision to and round down"},

		{NewWithPrec(55, 1), math.RoundUp, 0, 0, New(6), "5.5 set precision to 0 and round up"},
		{NewWithPrec(25, 1), math.RoundUp, 0, 0, New(3), "2.5 set precision to 0 and round up"},
		{NewWithPrec(16, 1), math.RoundUp, 0, 0, New(2), "1.6 set precision to 0 and round up"},
		{NewWithPrec(11, 1), math.RoundUp, 0, 0, New(2), "1.1 set precision to 0 and round up"},
		{NewWithPrec(10, 1), math.RoundUp, 0, 0, New(1), "1.0 set precision to 0 and round up"},
		{NewWithPrec(-10, 1), math.RoundUp, 0, 0, New(-1), "-1.0 set precision to 0 and round up"},
		{NewWithPrec(-11, 1), math.RoundUp, 0, 0, New(-2), "-1.1 set precision to 0 and round up"},
		{NewWithPrec(-16, 1), math.RoundUp, 0, 0, New(-2), "-1.6 set precision to 0 and round up"},
		{NewWithPrec(-25, 1), math.RoundUp, 0, 0, New(-3), "-2.5 set precision to 0 and round up"},
		{NewWithPrec(-55, 1), math.RoundUp, 0, 0, New(-6), "-5.5 set precision to 0 and round up"},

		{NewWithPrec(55, 1), math.RoundDown, 0, 0, New(5), "5.5 set precision to 0 and round down"},
		{NewWithPrec(25, 1), math.RoundDown, 0, 0, New(2), "2.5 set precision to 0 and round down"},
		{NewWithPrec(16, 1), math.RoundDown, 0, 0, New(1), "1.6 set precision to 0 and round down"},
		{NewWithPrec(11, 1), math.RoundDown, 0, 0, New(1), "1.1 set precision to 0 and round down"},
		{NewWithPrec(10, 1), math.RoundDown, 0, 0, New(1), "1.0 set precision to 0 and round down"},
		{NewWithPrec(-10, 1), math.RoundDown, 0, 0, New(-1), "-1.0 set precision to 0 and round down"},
		{NewWithPrec(-11, 1), math.RoundDown, 0, 0, New(-1), "-1.1 set precision to 0 and round down"},
		{NewWithPrec(-16, 1), math.RoundDown, 0, 0, New(-1), "-1.6 set precision to 0 and round down"},
		{NewWithPrec(-25, 1), math.RoundDown, 0, 0, New(-2), "-2.5 set precision to 0 and round down"},
		{NewWithPrec(-55, 1), math.RoundDown, 0, 0, New(-5), "-5.5 set precision to 0 and round down"},

		{NewWithPrec(55, 1), math.RoundCeiling, 0, 0, New(6), "5.5 set precision to 0 and round ceiling"},
		{NewWithPrec(25, 1), math.RoundCeiling, 0, 0, New(3), "2.5 set precision to 0 and round ceiling"},
		{NewWithPrec(16, 1), math.RoundCeiling, 0, 0, New(2), "1.6 set precision to 0 and round ceiling"},
		{NewWithPrec(11, 1), math.RoundCeiling, 0, 0, New(2), "1.1 set precision to 0 and round ceiling"},
		{NewWithPrec(10, 1), math.RoundCeiling, 0, 0, New(1), "1.0 set precision to 0 and round ceiling"},
		{NewWithPrec(-10, 1), math.RoundCeiling, 0, 0, New(-1), "-1.0 set precision to 0 and round ceiling"},
		{NewWithPrec(-11, 1), math.RoundCeiling, 0, 0, New(-1), "-1.1 set precision to 0 and round ceiling"},
		{NewWithPrec(-16, 1), math.RoundCeiling, 0, 0, New(-1), "-1.6 set precision to 0 and round ceiling"},
		{NewWithPrec(-25, 1), math.RoundCeiling, 0, 0, New(-2), "-2.5 set precision to 0 and round ceiling"},
		{NewWithPrec(-55, 1), math.RoundCeiling, 0, 0, New(-5), "-5.5 set precision to 0 and round ceiling"},

		{NewWithPrec(55, 1), math.RoundHalfUp, 0, 0, New(6), "5.5 set precision to 0 and round half up"},
		{NewWithPrec(25, 1), math.RoundHalfUp, 0, 0, New(3), "2.5 set precision to 0 and round half up"},
		{NewWithPrec(16, 1), math.RoundHalfUp, 0, 0, New(2), "1.6 set precision to 0 and round half up"},
		{NewWithPrec(11, 1), math.RoundHalfUp, 0, 0, New(1), "1.1 set precision to 0 and round half up"},
		{NewWithPrec(10, 1), math.RoundHalfUp, 0, 0, New(1), "1.0 set precision to 0 and round half up"},
		{NewWithPrec(-10, 1), math.RoundHalfUp, 0, 0, New(-1), "-1.0 set precision to 0 and round half up"},
		{NewWithPrec(-11, 1), math.RoundHalfUp, 0, 0, New(-1), "-1.1 set precision to 0 and round half up"},
		{NewWithPrec(-16, 1), math.RoundHalfUp, 0, 0, New(-2), "-1.6 set precision to 0 and round half up"},
		{NewWithPrec(-25, 1), math.RoundHalfUp, 0, 0, New(-3), "-2.5 set precision to 0 and round half up"},
		{NewWithPrec(-55, 1), math.RoundHalfUp, 0, 0, New(-6), "-5.5 set precision to 0 and round half up"},

		{NewWithPrec(55, 1), math.RoundHalfDown, 0, 0, New(5), "5.5 set precision to 0 and round half down"},
		{NewWithPrec(25, 1), math.RoundHalfDown, 0, 0, New(2), "2.5 set precision to 0 and round half down"},
		{NewWithPrec(16, 1), math.RoundHalfDown, 0, 0, New(2), "1.6 set precision to 0 and round half down"},
		{NewWithPrec(11, 1), math.RoundHalfDown, 0, 0, New(1), "1.1 set precision to 0 and round half down"},
		{NewWithPrec(10, 1), math.RoundHalfDown, 0, 0, New(1), "1.0 set precision to 0 and round half down"},
		{NewWithPrec(-10, 1), math.RoundHalfDown, 0, 0, New(-1), "-1.0 set precision to 0 and round half down"},
		{NewWithPrec(-11, 1), math.RoundHalfDown, 0, 0, New(-1), "-1.1 set precision to 0 and round half down"},
		{NewWithPrec(-16, 1), math.RoundHalfDown, 0, 0, New(-2), "-1.6 set precision to 0 and round half down"},
		{NewWithPrec(-25, 1), math.RoundHalfDown, 0, 0, New(-2), "-2.5 set precision to 0 and round half down"},
		{NewWithPrec(-55, 1), math.RoundHalfDown, 0, 0, New(-5), "-5.5 set precision to 0 and round half down"},

		{NewWithPrec(55, 1), math.RoundHalfEven, 0, 0, New(6), "5.5 set precision to 0 and round half even"},
		{NewWithPrec(25, 1), math.RoundHalfEven, 0, 0, New(2), "2.5 set precision to 0 and round half even"},
		{NewWithPrec(16, 1), math.RoundHalfEven, 0, 0, New(2), "1.6 set precision to 0 and round half even"},
		{NewWithPrec(11, 1), math.RoundHalfEven, 0, 0, New(1), "1.1 set precision to 0 and round half even"},
		{NewWithPrec(10, 1), math.RoundHalfEven, 0, 0, New(1), "1.0 set precision to 0 and round half even"},
		{NewWithPrec(-10, 1), math.RoundHalfEven, 0, 0, New(-1), "-1.0 set precision to 0 and round half even"},
		{NewWithPrec(-11, 1), math.RoundHalfEven, 0, 0, New(-1), "-1.1 set precision to 0 and round half even"},
		{NewWithPrec(-16, 1), math.RoundHalfEven, 0, 0, New(-2), "-1.6 set precision to 0 and round half even"},
		{NewWithPrec(-25, 1), math.RoundHalfEven, 0, 0, New(-2), "-2.5 set precision to 0 and round half even"},
		{NewWithPrec(-55, 1), math.RoundHalfEven, 0, 0, New(-6), "-5.5 set precision to 0 and round half even"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := test.value.Rescale(test.rescaleValue, test.roundingMode)
			if value.Precision() != test.wantPrecision {
				t.Errorf("expected %d, got %d", test.wantPrecision, value.Precision())
				return
			}

			if !value.Equal(test.want) {
				t.Errorf("expected %s, got %s", test.want, value)
				return
			}
		})
	}

	t.Run("should panic if rounding mode is round unnecessary and value is invalid", func(t *testing.T) {
		value := NewWithPrec(10001, 4)
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		value.Rescale(0, math.RoundUnnecessary)
		t.Fail()
	})
	t.Run("should success if rounding mode is round unnecessary and value is valid", func(t *testing.T) {
		value := NewWithPrec(10000, 4)
		if value.Rescale(0, math.RoundUnnecessary).Cmp(New(1)) != 0 {
			t.Errorf("expect %s, got %s", New(1), value)
		}
	})
}

func TestDecimal_StripTrailingZeros(t *testing.T) {
	tests := []struct {
		value Decimal
		want  string
		name  string
	}{
		{MustFromString("0"), "0", "0"},
		{MustFromString("0.00"), "0", "0.00"},
		{MustFromString("0.10"), "0.1", "0.10"},
		{MustFromString("0.11"), "0.11", "0.11"},
		{MustFromString("0.110000000000"), "0.11", "0.110000000000"},

		{MustFromString("-0"), "0", "-0"},
		{MustFromString("-0.00"), "0", "-0.00"},
		{MustFromString("-0.10"), "-0.1", "-0.10"},
		{MustFromString("-0.11"), "-0.11", "-0.11"},
		{MustFromString("-0.110000000000"), "-0.11", "-0.110000000000"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.value.StripTrailingZeros().String() != test.want {
				t.Errorf("expected %s, got %s", test.want, test.value.StripTrailingZeros().String())
			}
		})
	}
}

func TestDecimal_SignificantFigures(t *testing.T) {
	tests := []struct {
		value        Decimal
		figures      int
		roundingMode math.RoundingMode
		want         string
		name         string
	}{
		{MustFromString("0"), 1, math.RoundUp, "0", "0"},
		{MustFromString("0"), 10, math.RoundUp, "0", "0"},

		{MustFromString("0.001"), 1, math.RoundUp, "0.001", "0.001"},
		{MustFromString("0.001"), 2, math.RoundUp, "0.001", "0.001"},
		{MustFromString("0.001"), 10, math.RoundUp, "0.001", "0.001"},

		{MustFromString("-0.001"), 1, math.RoundUp, "-0.001", "-0.001"},
		{MustFromString("-0.001"), 2, math.RoundUp, "-0.001", "-0.001"},
		{MustFromString("-0.001"), 10, math.RoundUp, "-0.001", "-0.001"},

		{MustFromString("0.001001"), 1, math.RoundUp, "0.002", "0.001001"},
		{MustFromString("0.001001"), 2, math.RoundUp, "0.0011", "0.001001"},
		{MustFromString("0.001001"), 4, math.RoundUp, "0.001001", "0.001001"},

		{MustFromString("-0.001001"), 1, math.RoundUp, "-0.002", "-0.001001"},
		{MustFromString("-0.001001"), 2, math.RoundUp, "-0.0011", "-0.001001"},
		{MustFromString("-0.001001"), 4, math.RoundUp, "-0.001001", "-0.001001"},

		{MustFromString("1.001001"), 1, math.RoundUp, "2", "1.001001"},
		{MustFromString("1.001001"), 2, math.RoundUp, "1.1", "1.001001"},
		{MustFromString("1.001001"), 4, math.RoundUp, "1.002", "1.001001"},

		{MustFromString("-1.001001"), 1, math.RoundUp, "-2", "-1.001001"},
		{MustFromString("-1.001001"), 2, math.RoundUp, "-1.1", "-1.001001"},
		{MustFromString("-1.001001"), 4, math.RoundUp, "-1.002", "-1.001001"},

		{MustFromString("1111.001001"), 1, math.RoundUp, "1112", "1111.001001"},
		{MustFromString("1111.001001"), 3, math.RoundUp, "1112", "1111.001001"},
		{MustFromString("1111.001001"), 4, math.RoundUp, "1112", "1111.001001"},
		{MustFromString("1111.001001"), 5, math.RoundUp, "1111.1", "1111.001001"},

		{MustFromString("-1111.001001"), 1, math.RoundUp, "-1112", "-1111.001001"},
		{MustFromString("-1111.001001"), 3, math.RoundUp, "-1112", "-1111.001001"},
		{MustFromString("-1111.001001"), 4, math.RoundUp, "-1112", "-1111.001001"},
		{MustFromString("-1111.001001"), 5, math.RoundUp, "-1111.1", "-1111.001001"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.value.SignificantFigures(test.figures, test.roundingMode).String() != test.want {
				t.Errorf("expected %s, got %s", test.want, test.value.SignificantFigures(test.figures, test.roundingMode).String())
			}
		})
	}
}
