package bigint

import (
	"github.com/gridexswap/utils/math"
	"math/big"
)

func (b BigInt) quo(b2 BigInt, mode math.RoundingMode) BigInt {
	switch mode {
	case math.RoundDown:
		return b.quoDown(b2)
	case math.RoundUp:
		return b.quoUp(b2)
	case math.RoundCeiling:
		return b.quoCeiling(b2)
	case math.RoundUnnecessary:
		return b.quoUnnecessary(b2)
	default:
		panic("invalid rounding mode")
	}
}

func (b BigInt) quoDown(b2 BigInt) BigInt {
	return NewFromBigInt(new(big.Int).Quo(b.i, b2.i))
}

func (b BigInt) quoUp(b2 BigInt) BigInt {
	quo, rem := new(big.Int).QuoRem(b.i, b2.i, new(big.Int))
	if rem.Sign() == 0 {
		return NewFromBigInt(quo)
	}
	isNegative := b.Sign() != b2.Sign()
	abs := quo
	if isNegative {
		abs = new(big.Int).Abs(quo)
	}
	i := new(big.Int).Add(abs, big.NewInt(1))
	if isNegative {
		i = new(big.Int).Neg(i)
	}
	return NewFromBigInt(i)
}

func (b BigInt) quoCeiling(b2 BigInt) BigInt {
	quo, rem := new(big.Int).QuoRem(b.i, b2.i, new(big.Int))
	if rem.Sign() == 0 {
		return NewFromBigInt(quo)
	}
	if b.Sign() == b2.Sign() {
		return b.quoUp(b2)
	} else {
		return b.quoDown(b2)
	}
}

func (b BigInt) quoUnnecessary(b2 BigInt) BigInt {
	quo, rem := new(big.Int).QuoRem(b.i, b2.i, new(big.Int))
	if rem.Sign() != 0 {
		panic("expected 0 remainder")
	}
	return NewFromBigInt(quo)
}
