package decimal

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/gridexswap/utils/math"
	"github.com/gridexswap/utils/math/bigint"
)

const (
	MaxPrecision = 128

	// max number of iterations in Sqrt, Log2 function
	maxIterations = 300
)

var (
	precisionMultipliers []*big.Int

	Zero = NewDecimal(0)
	// Deprecated: use Zero instead
	ZeroDecimal = Zero
	One         = NewDecimal(1)
	// Deprecated: use One instead
	OneDecimal = One
	Ten        = NewDecimal(10)
	// Deprecated: use Ten instead
	TenDecimal = Ten
)

// Decimal
// CONTRACT: prec <= MaxPrecision
type Decimal struct {
	i    *big.Int
	prec int
}

func init() {
	precisionMultipliers = make([]*big.Int, MaxPrecision+1)
	precisionMultipliers[0] = big.NewInt(1)
	for i := 0; i <= MaxPrecision; i++ {
		precisionMultipliers[i] = new(big.Int).Exp(tenInt, big.NewInt(int64(i)), nil)
	}
}

func New(value int64) Decimal {
	return NewFromBigInt(big.NewInt(value))
}

// NewDecimal create a new Decimal from int64
//
// Deprecated: use New instead
func NewDecimal(value int64) Decimal {
	return NewFromBigInt(big.NewInt(value))
}

func NewWithPrec(value int64, prec int) Decimal {
	return NewFromBigIntWithPrec(big.NewInt(value), prec)
}

// NewDecimalWithPrec create a new Decimal from int64
//
// Deprecated: use NewWithPrec instead
func NewDecimalWithPrec(value int64, prec int) Decimal {
	return NewFromBigIntWithPrec(big.NewInt(value), prec)
}

func NewFromFloat64(value float64) Decimal {
	return MustFromString(strconv.FormatFloat(value, 'f', -1, 64))
}

// NewDecimalFromFloat64 create a new Decimal from float64
//
// Deprecated: use NewFromFloat64 instead
func NewDecimalFromFloat64(value float64) Decimal {
	return MustFromString(strconv.FormatFloat(value, 'f', -1, 64))
}

func NewWithAppendPrec(value int64, prec int) Decimal {
	return NewFromBigIntWithPrec(
		new(big.Int).Exp(tenInt, big.NewInt(int64(prec)), nil),
		prec,
	).Mul(New(value), math.RoundUnnecessary)
}

// NewDecimalWithAppendPrec create a new Decimal from value, and append number of zeros to make it fit the required precision
// If `value` is 1, `prec` is 2, then return 1.00.
// If `value` is 1, `prec` is 18, then return 1.000000000000000000
//
// Deprecated: use NewWithAppendPrec instead
func NewDecimalWithAppendPrec(value int64, prec int) Decimal {
	return NewWithAppendPrec(value, prec)
}

func NewFromUintWithAppendPrec(value uint64, prec int) Decimal {
	return NewFromBigIntWithPrec(
		new(big.Int).Exp(tenInt, big.NewInt(int64(prec)), nil),
		prec,
	).Mul(NewFromUint64(value, 0), math.RoundUnnecessary)
}

// Deprecated: use NewFromUintWithAppendPrec instead
func NewDecimalFromUintWithAppendPrec(value uint64, prec int) Decimal {
	return NewFromUintWithAppendPrec(value, prec)
}

func NewFromBigInt(value *big.Int) Decimal {
	return NewFromBigIntWithPrec(value, 0)
}

// NewDecimalFromBigInt create a new Decimal from big integer assuming whole numbers
//
// Deprecated: use NewFromBigInt instead
func NewDecimalFromBigInt(value *big.Int) Decimal {
	return NewFromBigInt(value)
}

func NewFromBigIntWithPrec(value *big.Int, precision int) Decimal {
	requirePrecision(precision)
	return Decimal{
		i:    new(big.Int).Mul(value, big.NewInt(1)),
		prec: precision,
	}
}

// NewDecimalFromBigIntWithPrec create a new Decimal from big integer assuming whole numbers
// CONTRACT: prec <= MaxPrecision
//
// Deprecated: use NewFromBigIntWithPrec instead
func NewDecimalFromBigIntWithPrec(value *big.Int, precision int) Decimal {
	return NewFromBigIntWithPrec(value, precision)
}

func NewFromInt64(value int64, precision int) Decimal {
	requirePrecision(precision)
	return Decimal{
		i:    new(big.Int).SetInt64(value),
		prec: precision,
	}
}

// Deprecated: use NewFromInt64 instead
func NewDecimalFromInt64(value int64, precision int) Decimal {
	return NewFromInt64(value, precision)
}

func NewFromUint64(value uint64, precision int) Decimal {
	requirePrecision(precision)
	return Decimal{
		i:    new(big.Int).SetUint64(value),
		prec: precision,
	}
}

// NewDecimalFromUint64 create a new Decimal from uint64 value.
// CONTRACT: prec <= MaxPrecision
//
// Deprecated: use NewFromUint64 instead
func NewDecimalFromUint64(value uint64, precision int) Decimal {
	return NewFromUint64(value, precision)
}

func NewFromString(str string) (d Decimal, err error) {
	return NewDecimalFromString(str)
}

// NewDecimalFromString create a new Decimal from decimal string.
// valid must come in the form:
//
//	(-) whole integers (.) decimal integers
//
// examples of acceptable input include:
//
//	-123.456
//	456.7890
//	345
//	-456789
//
// NOTE - An error will return if more decimal places
// are provided in the string than the constant Precision.
//
// CONTRACT - This function does not mutate the input str.
//
// Deprecated: use NewFromString instead
func NewDecimalFromString(str string) (d Decimal, err error) {
	var precision = 0
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return Decimal{}, errors.New("decimal string cannot be empty")
	}

	// Check if number is using scientific notation
	eIndex := strings.IndexAny(str, "Ee")
	if eIndex != -1 {
		expInt, err := strconv.ParseInt(str[eIndex+1:], 10, 32)
		if err != nil {
			if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
				return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", str)
			}
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric", str)
		}
		str = str[:eIndex]
		precision -= int(expInt)
	}

	// first extract any negative symbol
	neg := false
	if str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return Decimal{}, errors.New("decimal string cannot be empty")
	}

	strs := strings.Split(str, ".")
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		precision += len(strs[1])
		if precision == 0 || len(combinedStr) == 0 {
			return Decimal{}, errors.New("invalid decimal string")
		}
		combinedStr += strs[1]
	} else if len(strs) > 2 {
		return Decimal{}, errors.New("invalid decimal string")
	}

	if precision > MaxPrecision {
		return Decimal{}, fmt.Errorf("invalid precision; max: %d, got: %d", MaxPrecision, precision)
	}

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return Decimal{}, fmt.Errorf("failed to set decimal string: %s", combinedStr)
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}

	for precision < 0 {
		combined = new(big.Int).Mul(combined, big.NewInt(10))
		precision++
	}

	return Decimal{
		i:    combined,
		prec: precision,
	}, nil
}

func MustFromString(str string) Decimal {
	return MustDecimalFromString(str)
}

// Deprecated: use MustFromString instead
func MustDecimalFromString(str string) Decimal {
	d, err := NewDecimalFromString(str)
	if err != nil {
		panic(err)
	}
	return d
}

func (d Decimal) Add(d2 Decimal) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Add(d1.i, d2.i),
		prec: maxPrec,
	}
}

func (d Decimal) SafeAdd(d2 Decimal) Decimal {
	return d.Add(d2).requireNonNegative()
}

func (d Decimal) AddRaw(i int64) Decimal {
	return Decimal{
		i:    new(big.Int).Add(d.i, big.NewInt(i)),
		prec: d.prec,
	}
}

func (d Decimal) UnsignedAdd(d2 Decimal, bitLen *BitLen) Decimal {
	result := d.Add(d2)
	result.i = bitLen.limit(result.i)
	return result
}

func (d Decimal) UnsignedAddOverflow(d2 Decimal, bitLen *BitLen) (result Decimal, overflow bool) {
	result = d.Add(d2)
	overflow = result.BitLen() > bitLen.bitLen
	result.i = bitLen.limit(result.i)
	return result, overflow
}

func (d Decimal) Sub(d2 Decimal) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Sub(d1.i, d2.i),
		prec: maxPrec,
	}
}

func (d Decimal) SafeSub(d2 Decimal) Decimal {
	return d.Sub(d2).requireNonNegative()
}

func (d Decimal) SubRaw(i int64) Decimal {
	return Decimal{
		i:    new(big.Int).Sub(d.i, big.NewInt(i)),
		prec: d.prec,
	}
}

func (d Decimal) UnsignedSub(d2 Decimal, bitLen *BitLen) Decimal {
	result := d.Sub(d2)
	result.i = bitLen.limit(result.i)
	return result
}

func (d Decimal) UnsignedSubOverflow(d2 Decimal, bitLen *BitLen) (result Decimal, overflow bool) {
	result = d.Sub(d2)
	overflow = result.BitLen() > bitLen.bitLen
	result.i = bitLen.limit(result.i)
	return result, overflow
}

func (d Decimal) Mul(d2 Decimal, roundingMode math.RoundingMode) Decimal {
	d1, d2, maxPrec := rescalePair(d, d2)

	return Decimal{
		i:    new(big.Int).Mul(d1.i, d2.i),
		prec: maxPrec,
	}.round(roundingMode)
}

func (d Decimal) MulDown(d2 Decimal) Decimal {
	return d.Mul(d2, math.RoundDown)
}

func (d Decimal) UnsignedMul(d2 Decimal, roundingMode math.RoundingMode, bitLen *BitLen) Decimal {
	result := d.Mul(d2, roundingMode)
	result.i = bitLen.limit(result.i)
	return result
}

func (d Decimal) UnsignedMulDown(d2 Decimal, bitLen *BitLen) Decimal {
	return d.UnsignedMul(d2, math.RoundDown, bitLen)
}

func (d Decimal) UnsignedMulOverflow(d2 Decimal, roundingMode math.RoundingMode, bitLen *BitLen) (result Decimal, overflow bool) {
	result = d.Mul(d2, roundingMode)
	overflow = result.i.BitLen() > bitLen.bitLen
	result.i = bitLen.limit(result.i)
	return result, overflow
}

func (d Decimal) Quo(d2 Decimal, roundingMode math.RoundingMode) Decimal {
	// To adapt to the situation where the precision of both numbers is 0,
	// the precision of both numbers is increased by 1, and the final calculation
	// result is rescaled to 0.
	if d.prec == 0 && d2.prec == 0 {
		d1, d2 := d.RescaleDown(1), d2.RescaleDown(1)
		// multiply precision twice
		d1Twice := new(big.Int).Mul(d1.i, precisionMultipliers[1])
		d1Twice = new(big.Int).Mul(d1Twice, precisionMultipliers[1])

		return Decimal{
			i:    new(big.Int).Quo(d1Twice, d2.i),
			prec: 1 * 2,
		}.Rescale(0, roundingMode)
	}

	d1, d2, maxPrec := rescalePair(d, d2)
	// multiply precision twice
	d1Twice := new(big.Int).Mul(d1.i, precisionMultipliers[maxPrec])
	d1Twice = new(big.Int).Mul(d1Twice, precisionMultipliers[maxPrec])

	return Decimal{
		i:    new(big.Int).Quo(d1Twice, d2.i),
		prec: maxPrec,
	}.round(roundingMode)
}

func (d Decimal) QuoDown(d2 Decimal) Decimal {
	return d.Quo(d2, math.RoundDown)
}

func (d Decimal) UnsignedQuo(d2 Decimal, roundingMode math.RoundingMode, bitLen *BitLen) Decimal {
	result := d.Quo(d2, roundingMode)
	result.i = bitLen.limit(result.i)
	return result
}

func (d Decimal) UnsignedQuoDown(d2 Decimal, bitLen *BitLen) Decimal {
	return d.UnsignedQuo(d2, math.RoundDown, bitLen)
}

func (d Decimal) UnsignedQuoOverflow(d2 Decimal, roundingMode math.RoundingMode, bitLen *BitLen) (result Decimal, overflow bool) {
	result = d.Quo(d2, roundingMode)
	overflow = result.i.BitLen() > bitLen.bitLen
	result.i = bitLen.limit(result.i)
	return result, overflow
}

// IntPart returns integer part.
func (d Decimal) IntPart() *big.Int {
	intPart, _ := d.Remainder()
	return intPart
}

// Remainder returns integer part and fractional part.
func (d Decimal) Remainder() (intPart *big.Int, fractionPart *big.Int) {
	return new(big.Int).QuoRem(d.i, precisionMultipliers[d.prec], new(big.Int))
}

// Power returns a result of raising to integer power.
func (d Decimal) Power(power int64) Decimal {
	if power == 0 {
		return One.Rescale(d.prec, math.RoundUnnecessary)
	}

	if power < 0 {
		// If power is negative, we will return a round up value
		return One.Quo(d.Power(-power), math.RoundUp)
	}

	tmp, resultD := NewWithAppendPrec(1, d.prec), d
	for i := power; i > 1; {
		if i%2 != 0 {
			tmp = tmp.Mul(resultD, math.RoundHalfEven)
		}
		i /= 2
		resultD = resultD.Mul(resultD, math.RoundHalfEven)
	}
	return resultD.Mul(tmp, math.RoundHalfEven)
}

// Sqrt sets z to ⌊√x⌋, the largest integer such that z² ≤ x, and returns z.
// It returns -(sqrt(abs(d)) if input is negative.
func (d Decimal) Sqrt() (guess Decimal, err error) {
	return d.ApproxRoot(2)
}

func (d Decimal) ApproxRoot(root int64) (guess Decimal, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = errors.New("out of bounds")
			}
		}
	}()

	if d.IsNegative() {
		absRoot, err := d.Neg().ApproxRoot(root)
		return absRoot.Neg(), err
	}

	if root == 1 || d.IsZero() || d.Equal(One) {
		return d, nil
	}

	if root == 0 {
		return One.Rescale(d.prec, math.RoundUnnecessary), nil
	}

	rootInt := big.NewInt(0).SetInt64(root)
	guess = NewWithAppendPrec(1, d.prec)
	delta := guess

	for iter := 0; delta.Abs().i.Cmp(oneInt) > 0 && iter < maxIterations; iter++ {
		prev := guess.Power(root - 1)
		if prev.IsZero() {
			prev = One
		}
		delta = d.Quo(prev, math.RoundHalfEven)
		delta = delta.Sub(guess)

		quo := new(big.Int).Quo(delta.i, rootInt)
		delta = Decimal{i: quo, prec: d.prec}

		guess = guess.Add(delta)
	}
	return
}

// Log2 returns log2.
func (d Decimal) Log2() Decimal {
	if d.Sign() <= 0 {
		panic("value must greater than 0")
	}

	oneDec := NewWithAppendPrec(1, d.prec)
	twoDec := NewWithAppendPrec(2, d.prec)

	lessOne := d.Cmp(oneDec) < 0
	copyD := d
	exp := 4 * d.prec
	if lessOne {
		// Ensure copyD greater than 1
		copyD = copyD.Mul(New(2).Power(int64(exp)), math.RoundHalfEven)
	}

	intPart, _ := copyD.Remainder()
	n := math.MostSignificantBit(intPart)
	resultDec := NewFromUintWithAppendPrec(uint64(n), copyD.prec)

	int64N := int64(n)
	if int64N < 0 {
		panic(fmt.Sprintf("Most Significant Bit %d too larger", n))
	}

	remDec := copyD.Quo(New(2).Power(int64N), math.RoundHalfEven)
	for i := 0; i < maxIterations && remDec.Sign() > 0; i++ {
		if remDec.GTE(twoDec) {
			resultDec = resultDec.Add(oneDec.Quo(twoDec.Power(int64(i)), math.RoundHalfEven))
			remDec = remDec.Quo(twoDec, math.RoundHalfEven)
		}
		remDec = remDec.Power(2)
	}

	if lessOne {
		resultDec = resultDec.Sub(New(int64(exp)))
	}
	return resultDec
}

func (d Decimal) RescaleDown(prec int) Decimal {
	return d.Rescale(prec, math.RoundDown)
}

func (d Decimal) Rescale(prec int, roundingMode math.RoundingMode) Decimal {
	if d.prec == prec {
		return d
	}

	diff := d.prec - prec
	var newI = new(big.Int)
	if diff < 0 {
		// Mul never should round
		newI.Mul(d.i, precisionMultipliers[-diff])
	} else {
		roundedDecimal := Decimal{
			i:    d.i,
			prec: diff,
		}.round(roundingMode)
		return Decimal{
			i:    roundedDecimal.i,
			prec: prec,
		}
	}
	return Decimal{
		i:    newI,
		prec: prec,
	}
}

// StripTrailingZeros returns a Decimal which is numerically equal to this one
// but with any trailing zeros removed from the representation.
func (d Decimal) StripTrailingZeros() Decimal {
	if d.prec == 0 {
		return d
	}
	str := d.String()
	splits := strings.Split(str, ".")
	foundNotZero := false
	for i := len(splits[1]) - 1; i >= 0; i-- {
		if splits[1][i] != '0' {
			splits[1] = splits[1][:i+1]
			foundNotZero = true
			break
		}
	}
	if len(splits[1]) == 0 || !foundNotZero {
		return MustFromString(splits[0])
	}
	return MustFromString(fmt.Sprintf("%s.%s", splits[0], splits[1]))
}

// SignificantFigures returns a Decimal with the specified number of significant figures
func (d Decimal) SignificantFigures(figures int, roundingMode math.RoundingMode) Decimal {
	if figures <= 0 {
		panic("figures must be greater than 0")
	}
	if d.prec == 0 || d.prec <= figures {
		return d
	}

	absD := d.Abs()
	str := absD.String()
	splits := strings.Split(str, ".")
	if splits[0] != "0" {
		figures -= len(splits[0])
		if figures < 0 {
			figures = 0
		}
		return d.Rescale(min(figures, d.prec), roundingMode)
	} else {
		for i := 0; i < len(splits[1]); i++ {
			if splits[1][i] != '0' {
				figures += i
				return d.Rescale(min(figures, d.prec), roundingMode)
			}
		}
	}
	return d
}

func (d Decimal) MustNonNegative() Decimal {
	return d.requireNonNegative()
}

func (d Decimal) requireNonNegative() Decimal {
	if d.Sign() < 0 {
		panic("Negative value")
	}
	return d
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (d Decimal) Cmp(d2 Decimal) int {
	d1, d2, _ := rescalePair(d, d2)
	return d1.i.Cmp(d2.i)
}

// Equal returns equal other value
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// GT greater than other value
func (d Decimal) GT(d2 Decimal) bool {
	return d.Cmp(d2) > 0
}

// GTE greater than or equal other value
func (d Decimal) GTE(d2 Decimal) bool {
	return d.Cmp(d2) >= 0
}

// LT less than other value
func (d Decimal) LT(d2 Decimal) bool {
	return d.Cmp(d2) < 0
}

// LTE less than or equal other value
func (d Decimal) LTE(d2 Decimal) bool {
	return d.Cmp(d2) <= 0
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (d Decimal) Sign() int {
	return d.i.Sign()
}

// IsNil is decimal nil
func (d Decimal) IsNil() bool {
	return d.i == nil
}

// IsNegative returns is negative value
func (d Decimal) IsNegative() bool {
	return d.Sign() < 0
}

// IsZero returns is zero value
func (d Decimal) IsZero() bool {
	return d.Sign() == 0
}

// IsPositive returns is positive value
func (d Decimal) IsPositive() bool {
	return d.Sign() > 0
}

// Neg reverse the decimal sign
func (d Decimal) Neg() Decimal {
	return Decimal{new(big.Int).Neg(d.i), d.prec}
}

// Abs returns absolute value
func (d Decimal) Abs() Decimal {
	if d.IsNegative() {
		return d.Neg()
	}
	// We can return d directly, because there is no way to modify the value of d.i
	return d
}

// BigInt returns a copy of the underlying big.Int.
func (d Decimal) BigInt() *big.Int {
	if d.IsNil() {
		return nil
	}

	cp := new(big.Int)
	return cp.Set(d.i)
}

// BigInt2 returns a copy of the underlying big.Int
func (d Decimal) BigInt2() bigint.BigInt {
	return bigint.NewFromBigInt(d.BigInt())
}

func (d Decimal) BitLen() int {
	return d.i.BitLen()
}

func (d Decimal) Precision() int {
	return d.prec
}

func requirePrecision(prec int) {
	if prec > MaxPrecision {
		panic("Precision too high")
	}
}

func rescalePair(d1, d2 Decimal) (rescaledD1, rescaledD2 Decimal, maxPrec int) {
	maxPrec = max(d1.prec, d2.prec)
	rescaledD1 = d1.RescaleDown(maxPrec)
	rescaledD2 = d2.RescaleDown(maxPrec)
	return
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
