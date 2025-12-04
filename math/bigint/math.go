package bigint

func Max(a, b BigInt) BigInt {
	if a.Cmp(b) >= 0 {
		return a
	}
	return b
}

func Min(a, b BigInt) BigInt {
	if a.Cmp(b) <= 0 {
		return a
	}
	return b
}
