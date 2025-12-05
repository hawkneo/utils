package decimal

import (
	"github.com/hawkneo/utils/math"
	"math/big"
)

// Deprecated: use math.RoundingMode instead
type RoundingMode = math.RoundingMode

const (
	// RoundDown rounding mode to round towards zero.
	//
	// Deprecated: use math.RoundDown instead.
	RoundDown = math.RoundDown
	// RoundUp rounding mode to round away from zero.
	//
	// Deprecated: use math.RoundUp instead.
	RoundUp
	// RoundCeiling rounding mode to round towards positive infinity.
	//
	// Deprecated: use math.RoundCeiling instead.
	RoundCeiling
	// RoundHalfUp rounding mode to round towards "nearest neighbor" unless both neighbors are equidistant, in which case round up.
	//
	// Deprecated: use math.RoundHalfUp instead.
	RoundHalfUp
	// RoundHalfDown rounding mode to round towards "nearest neighbor" unless both neighbors are equidistant, in which case round down.
	//
	// Deprecated: use math.RoundHalfDown instead.
	RoundHalfDown
	// RoundHalfEven rounding mode to round towards the "nearest neighbor" unless both neighbors are equidistant, in which case, round towards the even neighbor.
	// Alias: Banker's rounding.
	//
	// Deprecated: use math.RoundHalfEven instead.
	RoundHalfEven
	// RoundUnnecessary rounding mode to assert that the requested operation has an exact result, hence no rounding is necessary.
	//
	// Deprecated: use math.RoundUnnecessary instead.
	RoundUnnecessary
)

func (d Decimal) round(mode math.RoundingMode) Decimal {
	switch mode {
	case math.RoundDown:
		return d.roundDown()
	case math.RoundUp:
		return d.roundUp()
	case math.RoundCeiling:
		return d.roundCeiling()
	case math.RoundHalfUp:
		return d.roundHalfUp()
	case math.RoundHalfDown:
		return d.roundHalfDown()
	case math.RoundHalfEven:
		return d.roundHalfEven()
	case math.RoundUnnecessary:
		return d.roundUnnecessary()
	default:
		panic("invalid rounding mode")
	}
}

func (d Decimal) roundDown() Decimal {
	value := new(big.Int).Quo(d.i, precisionMultipliers[d.prec])
	return Decimal{
		i:    value,
		prec: d.prec,
	}
}

func (d Decimal) roundTruncate() Decimal {
	return d.roundDown()
}

func (d Decimal) roundUp() Decimal {
	if d.IsNegative() {
		// Make d positive
		abs := d.Neg()
		abs = abs.roundUp()
		return abs.Neg()
	}

	// Get the truncated quotient and remainder
	quo, rem := new(big.Int).QuoRem(d.i, precisionMultipliers[d.prec], new(big.Int))
	if rem.Sign() == 0 {
		return Decimal{
			i:    quo,
			prec: d.prec,
		}
	}

	return Decimal{
		i:    quo.Add(quo, oneInt),
		prec: d.prec,
	}
}

func (d Decimal) roundCeiling() Decimal {
	if d.IsNegative() {
		return d.roundDown()
	}
	return d.roundUp()
}

func (d Decimal) roundHalfUp() Decimal {
	if d.prec == 0 {
		return d
	}
	if d.IsNegative() {
		// Make a positive
		abs := d.Neg()
		abs = abs.roundHalfUp()
		return abs.Neg()
	}
	quo, rem := new(big.Int).QuoRem(d.i, precisionMultipliers[d.prec], new(big.Int))
	fivePrecision := new(big.Int).Mul(fiveInt, precisionMultipliers[d.prec-1])
	cmp := rem.Cmp(fivePrecision)
	if cmp < 0 {
		return Decimal{
			i:    quo,
			prec: d.prec,
		}
	} else {
		return Decimal{
			i:    quo.Add(quo, oneInt),
			prec: d.prec,
		}
	}
}

func (d Decimal) roundHalfDown() Decimal {
	if d.prec == 0 {
		return d
	}
	if d.IsNegative() {
		// Make a positive
		abs := d.Neg()
		abs = abs.roundHalfDown()
		return abs.Neg()
	}

	quo, rem := new(big.Int).QuoRem(d.i, precisionMultipliers[d.prec], new(big.Int))
	fivePrecision := new(big.Int).Mul(fiveInt, precisionMultipliers[d.prec-1])
	cmp := rem.Cmp(fivePrecision)
	if cmp <= 0 {
		return Decimal{
			i:    quo,
			prec: d.prec,
		}
	} else {
		return Decimal{
			i:    quo.Add(quo, oneInt),
			prec: d.prec,
		}
	}
}

func (d Decimal) roundHalfEven() Decimal {
	if d.prec == 0 {
		return d
	}
	if d.IsNegative() {
		// Make d positive
		abs := d.Neg()
		abs = abs.roundHalfEven()
		return abs.Neg()
	}

	quo, rem := new(big.Int).QuoRem(d.i, precisionMultipliers[d.prec], new(big.Int))
	fivePrecision := new(big.Int).Mul(fiveInt, precisionMultipliers[d.prec-1])
	cmp := rem.Cmp(fivePrecision)
	var resultD Decimal
	if cmp < 0 {
		resultD = Decimal{
			i:    quo,
			prec: d.prec,
		}
	} else if cmp > 0 {
		resultD = Decimal{
			i:    quo.Add(quo, oneInt),
			prec: d.prec,
		}
	} else {
		// Bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			resultD = Decimal{
				i:    quo,
				prec: d.prec,
			}
		} else {
			resultD = Decimal{
				i:    quo.Add(quo, oneInt),
				prec: d.prec,
			}
		}
	}

	return resultD
}

func (d Decimal) roundUnnecessary() Decimal {
	if d.IsNegative() {
		// Make d positive
		abs := d.Neg()
		abs = abs.roundUnnecessary()
		return abs.Neg()
	}

	quo, rem := new(big.Int).QuoRem(d.i, precisionMultipliers[d.prec], new(big.Int))
	if rem.Sign() != 0 {
		panic("expected 0 remainder")
	}
	return Decimal{
		i:    quo,
		prec: d.prec,
	}
}
