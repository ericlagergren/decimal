package math

import "github.com/ericlagergren/decimal"

// Max returns the greater of x and y, or x if x == y.
func Max(x, y *decimal.Big) *decimal.Big {
	if x.Cmp(y) >= 0 {
		return x
	}
	return y
}

// Min returns the lesser of x and y, or x if x == y.
func Min(x, y *decimal.Big) *decimal.Big {
	if x.Cmp(y) <= 0 {
		return x
	}
	return y
}
