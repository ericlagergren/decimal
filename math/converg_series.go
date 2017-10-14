package math

import "github.com/ericlagergren/decimal"

// TODO: generic taylor series function?

// expTaylor sets z to exp(x).
func expTaylor(z, x *decimal.Big) *decimal.Big {
	return nil
}

func expTaylorCompact(z *decimal.Big, x int64) *decimal.Big {
	var (
		sum  = int64(1)
		pow  = x
		fac  = int64(1)
		prev int64
	)
	for i := int64(2); sum != prev; i++ {
		pow *= x
		fac *= i
		term := pow / fac
		prev = sum
		sum += term
	}
	return nil
}
