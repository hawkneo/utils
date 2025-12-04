package decimal

import (
	"math/big"
)

type BitLen struct {
	bitLen int
	limit  func(i *big.Int) *big.Int
}

var (
	zeroInt = big.NewInt(0)
	oneInt  = big.NewInt(1)
	twoInt  = big.NewInt(2)
	fiveInt = big.NewInt(5)
	tenInt  = big.NewInt(10)
)

var (
	MaxUint128 = calcMaxUint(128)
	MaxUint256 = calcMaxUint(256)
)

func calcMaxUint(bitLen uint) *big.Int {
	max := new(big.Int).Lsh(oneInt, bitLen)
	return max.Sub(max, oneInt)
}

var (
	tt128 = new(big.Int).Lsh(big.NewInt(1), 128)
	tt256 = new(big.Int).Lsh(big.NewInt(1), 256)

	Uint128BitLen = &BitLen{128, func(i *big.Int) *big.Int {
		return truncateAdd(i, tt128, MaxUint128)
	}}

	Uint256BitLen = &BitLen{256, func(i *big.Int) *big.Int {
		return truncateAdd(i, tt256, MaxUint256)
	}}
)

func truncateAdd(i, tt, mask *big.Int) *big.Int {
	i = i.Add(i, tt)
	return i.And(i, mask)
}
