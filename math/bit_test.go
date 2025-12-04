package math

import (
	"math/big"
	"testing"
)

func TestMostSignificantBit(t *testing.T) {
	t.Run("x is 0", func(t *testing.T) {
		if MostSignificantBit(big.NewInt(0)) != 0 {
			t.Errorf("MostSignificantBit(0) = %d, want 0", MostSignificantBit(big.NewInt(0)))
		}
	})
	t.Run("x is negative value", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Panic was expected")
			}
		}()
		MostSignificantBit(big.NewInt(-1))
	})

	tests := []struct {
		value *big.Int
		want  uint
		name  string
	}{
		{big.NewInt(1), 0, "msb is zero"},
		{big.NewInt(2), 1, "msb is one"},
		{big.NewInt(3), 1, "msb is one"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := MostSignificantBit(test.value)
			if got != test.want {
				t.Errorf("LeastSignificantBit(%v) = %v, want %v", test.value, got, test.want)
			}
		})
	}

	t.Run("powers of 2 from 0-255", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			x := new(big.Int).Lsh(big.NewInt(1), uint(i))
			got := MostSignificantBit(x)
			want := uint(i)
			if got != want {
				t.Errorf("MostSignificantBit(%v) = %v, want %v", x, got, want)
			}
		}
	})
}

func TestLeastSignificantBit(t *testing.T) {
	t.Run("x is 0", func(t *testing.T) {
		if LeastSignificantBit(big.NewInt(0)) != 0 {
			t.Errorf("LeastSignificantBit(0) = %d, want 0", LeastSignificantBit(big.NewInt(0)))
		}
	})
	t.Run("x is negative value", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Panic was expected")
			}
		}()
		LeastSignificantBit(big.NewInt(-1))
	})

	tests := []struct {
		value *big.Int
		want  uint
		name  string
	}{
		{big.NewInt(1), 0, "lsb is zero"},
		{big.NewInt(2), 1, "lsb is one"},
		{big.NewInt(3), 0, "lsb is zero"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := LeastSignificantBit(test.value)
			if got != test.want {
				t.Errorf("LeastSignificantBit(%v) = %v, want %v", test.value, got, test.want)
			}
		})
	}

	t.Run("powers of 2 from 0-255", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			x := new(big.Int).Lsh(big.NewInt(1), uint(i))
			got := LeastSignificantBit(x)
			want := uint(i)
			if got != want {
				t.Errorf("LeastSignificantBit(%v) = %v, want %v", x, got, want)
			}
		}
	})
}
