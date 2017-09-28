package math

import "github.com/ericlagergren/decimal"

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	p0 := new(decimal.Big).Set(p)
	q0 := new(decimal.Big).Set(q)
	if p0.Sign() <= 0 {
		p0.Neg(p0)
	}
	if q0.Sign() <= 0 {
		q0.Neg(q0)
	}
	if p0.Sign() == 0 {
		return z.SetMantScale(0, 0)
	}
	p0.Mul(p0, p0)
	q0.Mul(q0, q0)
	return Sqrt(z, p0.Add(p0, q0))
}

// Sqrt sets z to the square root of x and returns z.
// The precision of Sqrt is determined by z's Context.
// Sqrt will panic on negative values since decimal.Big cannot
// represent imaginary numbers.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	return z.Sqrt(x)
}
