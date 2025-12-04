package math

type RoundingMode int

const (
	// RoundDown rounding mode to round towards zero.
	RoundDown RoundingMode = iota
	// RoundUp rounding mode to round away from zero.
	RoundUp
	// RoundCeiling rounding mode to round towards positive infinity.
	RoundCeiling
	// RoundHalfUp rounding mode to round towards "nearest neighbor" unless both neighbors are equidistant, in which case round up.
	RoundHalfUp
	// RoundHalfDown rounding mode to round towards "nearest neighbor" unless both neighbors are equidistant, in which case round down.
	RoundHalfDown
	// RoundHalfEven rounding mode to round towards the "nearest neighbor" unless both neighbors are equidistant, in which case, round towards the even neighbor.
	// Alias: Banker's rounding.
	RoundHalfEven
	// RoundUnnecessary rounding mode to assert that the requested operation has an exact result, hence no rounding is necessary.
	RoundUnnecessary
)
