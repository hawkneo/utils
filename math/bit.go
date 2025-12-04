package math

import (
	"math/big"
)

// MostSignificantBit returns the index of the most significant bit of the number.
// The function satisfies the property:
// x >= power(2, MostSignificantBit(x)) and x < power(2, MostSignificantBit(x)+1)
func MostSignificantBit(x *big.Int) uint {
	if x.Sign() < 0 {
		panic("MostSignificantBit of not positive number")
	}
	if x.Sign() == 0 {
		return 0
	}

	var msb uint = 0
	for bitLen := x.BitLen(); bitLen > 0; {
		bitLenHalf := bitLen >> 1
		if bitLenHalf<<1 != bitLen {
			bitLenHalf++
		}
		mask := new(big.Int).Lsh(big.NewInt(1), uint(bitLenHalf))
		mask = mask.Sub(mask, big.NewInt(1))
		if x.Cmp(mask) >= 0 {
			msb += uint(bitLenHalf)
			bitLen -= bitLenHalf
			x = x.Rsh(x, uint(bitLenHalf))
		} else {
			bitLen = bitLenHalf
		}
	}
	return msb - 1
}

// LeastSignificantBit returns the index of the least significant bit of the number.
// The function satisfies the property:
// x & power(2, LeastSignificantBit(x)) != 0 and (x & power(2, LeastSignificantBit(x) - 1)) == 0
func LeastSignificantBit(x *big.Int) (lsb uint) {
	if x.Sign() < 0 {
		panic("LeastSignificantBit of not positive number")
	}
	if x.Sign() == 0 {
		return 0
	}

	mask := big.NewInt(1)
	for i := 0; i < x.BitLen(); i++ {
		if new(big.Int).And(x, mask).Sign() > 0 {
			return uint(i)
		}
		mask = mask.Lsh(mask, 1)
	}
	panic("unreachable")
}
