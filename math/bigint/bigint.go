package bigint

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/gridexswap/utils/math"
)

var (
	Zero = NewFromInt(0)
	One  = NewFromInt(1)
	Ten  = NewFromInt(10)
)

// BigInt is a wrapper around big.Int that provides some convenience methods
//
// Note: BigInt is immutable, so all methods return a new BigInt
type BigInt struct {
	i *big.Int
}

func NewFromInt(i int) BigInt {
	return BigInt{i: big.NewInt(int64(i))}
}

func NewFromInt64(i int64) BigInt {
	return BigInt{i: new(big.Int).SetInt64(i)}
}

func NewFromUint(i uint) BigInt {
	return BigInt{i: new(big.Int).SetUint64(uint64(i))}
}

func NewFromUint64(i uint64) BigInt {
	return BigInt{i: new(big.Int).SetUint64(i)}
}

func NewFromBigInt(i *big.Int) BigInt {
	return BigInt{i: i}
}

// NewFromString returns a BigInt from a string.
//
// If the string starts with 0x or 0X, it is interpreted as a hex string.
// If the string starts with 0b or 0B, it is interpreted as a binary string.
// Otherwise, it is interpreted as a decimal string.
func NewFromString(s string) (BigInt, bool) {
	var base = 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		base = 16
		s = s[2:]
	} else if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		base = 2
		s = s[2:]
	}
	i, ok := new(big.Int).SetString(s, base)
	if !ok {
		return BigInt{}, false
	}
	return BigInt{i: i}, true
}

func MustNewFromString(s string) BigInt {
	b, ok := NewFromString(s)
	if !ok {
		panic(fmt.Errorf("invalid string %s", s))
	}
	return b
}

// Add returns the sum of b and b2
func (b BigInt) Add(b2 BigInt) BigInt {
	return NewFromBigInt(new(big.Int).Add(b.i, b2.i))
}

// Sub returns the difference of b and b2
func (b BigInt) Sub(b2 BigInt) BigInt {
	return NewFromBigInt(new(big.Int).Sub(b.i, b2.i))
}

// Mul returns the product of b and b2
func (b BigInt) Mul(b2 BigInt) BigInt {
	return NewFromBigInt(new(big.Int).Mul(b.i, b2.i))
}

// QuoDown returns the quotient of b and b2
func (b BigInt) QuoDown(b2 BigInt) BigInt {
	return b.quo(b2, math.RoundDown)
}

// Quo returns the quotient of b and b2
//
// Rounding mode is only support for math.RoundDown, math.RoundUp, math.RoundCeiling
// and math.RoundUnnecessary
func (b BigInt) Quo(b2 BigInt, roundingMode math.RoundingMode) BigInt {
	return b.quo(b2, roundingMode)
}

// Mod returns the modulus b % b2
func (b BigInt) Mod(b2 BigInt) BigInt {
	return NewFromBigInt(new(big.Int).Mod(b.i, b2.i))
}

// Power returns a result of raising to integer power.
func (b BigInt) Power(power int64) BigInt {
	return NewFromBigInt(new(big.Int).Exp(b.i, big.NewInt(power), nil))
}

// Sqrt returns the square root of b
func (b BigInt) Sqrt() BigInt {
	return NewFromBigInt(new(big.Int).Sqrt(b.i))
}

// ShiftLeft returns b shifted left by n bits
func (b BigInt) ShiftLeft(n uint) BigInt {
	return NewFromBigInt(new(big.Int).Lsh(b.i, n))
}

// ShiftRight returns b shifted right by n bits
func (b BigInt) ShiftRight(n uint) BigInt {
	return NewFromBigInt(new(big.Int).Rsh(b.i, n))
}

func (b BigInt) Cmp(b2 BigInt) int {
	return b.i.Cmp(b2.i)
}

func (b BigInt) Equal(b2 BigInt) bool {
	return b.i.Cmp(b2.i) == 0
}

func (b BigInt) GT(b2 BigInt) bool {
	return b.i.Cmp(b2.i) > 0
}

func (b BigInt) GTE(b2 BigInt) bool {
	return b.i.Cmp(b2.i) >= 0
}

func (b BigInt) LT(b2 BigInt) bool {
	return b.i.Cmp(b2.i) < 0
}

func (b BigInt) LTE(b2 BigInt) bool {
	return b.i.Cmp(b2.i) <= 0
}

func (b BigInt) Sign() int {
	return b.i.Sign()
}

func (b BigInt) IsNil() bool {
	return b.i == nil
}

func (b BigInt) IsNegative() bool {
	return b.Sign() < 0
}

func (b BigInt) IsZero() bool {
	return b.Sign() == 0
}

func (b BigInt) IsPositive() bool {
	return b.Sign() > 0
}

func (b BigInt) Neg() BigInt {
	return NewFromBigInt(new(big.Int).Neg(b.i))
}

func (b BigInt) Abs() BigInt {
	if b.IsNegative() {
		return b.Neg()
	}
	// We can return b directly, because there is no way to modify the value of b.i
	return b
}

func (b BigInt) BitLen() int {
	return b.i.BitLen()
}

// BigInt returns a copy of the underlying big.Int.
func (b BigInt) BigInt() *big.Int {
	return new(big.Int).Set(b.i)
}

// GetInt64 returns the int64 representation of x. If x cannot be represented in
// an int64, the result is undefined.
func (b BigInt) GetInt64() int64 {
	return b.i.Int64()
}
