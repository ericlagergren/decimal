package math

import (
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/misc"
)

// TODO(eric): Pow(z, x, y, m *decimal.Big) *decimal.Big

// Pow sets z to x**y and returns z.
func Pow(z, x, y *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, y) {
		return z
	}

	if xs := x.Sign(); xs <= 0 {
		if xs == 0 {
			if y.Sign() == 0 {
				// 0 ** 0 is undefined
				z.Context.Conditions |= decimal.InvalidOperation
				return z.SetNaN(false)
			}
			if y.Signbit() {
				// 0 ** -y = +Inf
				return z.SetInf(true)
			}
			// 0 ** y = 0
			return z.SetUint64(0)
		}
		if xs < 0 && (!y.IsInt() || y.IsInf(0)) {
			// -x ** y.vvv is undefined
			// -x ** ±Inf is undefined
			z.Context.Conditions |= decimal.InvalidOperation
			return z.SetNaN(false)
		}
	}

	if x.IsInf(0) {
		switch y.Sign() {
		case +1:
			// ±Inf ** y = +Inf
			return z.SetInf(false)
		case -1:
			// ±Inf ** -y = 0
			return z.SetUint64(0)
		case 0:
			// ±Inf ** 0 = 1
			return z.SetUint64(1)
		}
	}

	if y.Sign() == 0 {
		// x ** 0 = 1
		return z.SetUint64(1)
	}

	if x.Cmp(ptFive) == 0 {
		// x ** 0.5 = sqrt(x)
		return Sqrt(z, x)
	}

	if y.IsInt() {
		return powInt(z, x, y)
	}
	return powDec(z, x, y)
}

func powInt(z, x, y *decimal.Big) *decimal.Big {
	prec := precision(z)
	ctx := decimal.Context{Precision: prec - y.Scale() + y.Precision() + 2}

	var x0 decimal.Big
	if y.Signbit() {
		ctx.Precision++
		ctx.Quo(&x0, one, x)
	} else {
		x0.Copy(x)
	}
	z.SetUint64(1)
	sign := x.Signbit()

	// TODO(eric): the Uint64 branch only covers unsigned integers. Set it up
	// to handle signed as well.

	if yy, ok := y.Uint64(); ok {
		sign = sign && yy&1 == 1
		for yy != 0 {
			if yy&1 != 0 {
				ctx.Mul(z, z, &x0)
				if !z.IsFinite() || z.Sign() == 0 ||
					z.Context.Conditions&decimal.Clamped != 0 {
					z.Context.Conditions |= decimal.Underflow | decimal.Subnormal
					sign = false
					break
				}
			}
			yy >>= 1
			ctx.Mul(&x0, &x0, &x0)
			if x0.IsNaN(0) {
				z.Context.Conditions |= x0.Context.Conditions
				return z.SetNaN(false)
			}
		}
	} else {
		y0 := y.Int(nil)
		sign = sign && y0.Bit(0) == 1
		y0.Abs(y0)
		for y0.Sign() != 0 {
			if y0.Bit(0) != 0 {
				ctx.Mul(z, z, &x0)
				if !z.IsFinite() || z.Sign() == 0 ||
					z.Context.Conditions&decimal.Clamped != 0 {
					z.Context.Conditions |= decimal.Underflow | decimal.Subnormal
					sign = false
					break
				}
			}
			y0.Rsh(y0, 1)
			ctx.Mul(&x0, &x0, &x0)
			if x0.IsNaN(0) {
				z.Context.Conditions |= x0.Context.Conditions
				return z.SetNaN(false)
			}
		}
	}

	misc.SetSignbit(z, sign)
	ctx.Precision = prec
	return ctx.Round(z)
}

func powDec(z, x, y *decimal.Big) *decimal.Big {
	if z == y {
		y = new(decimal.Big).Copy(y)
	}
	neg := x.Signbit()
	if neg {
		x = misc.CopyAbs(new(decimal.Big), x)
	}

	oc := z.Context
	z.Context = decimal.Context{
		Precision: max(x.Precision(), precision(z)) + 4 + 19,
	}
	Exp(z, z.Mul(y, Log(z, x)))
	if neg && z.IsFinite() {
		misc.CopyNeg(z, z)
	}
	z.Context = oc
	return oc.Round(z)
}

// fastPowUint sets z to x ** y. It clobbers both z and x.
func fastPowUint(ctx decimal.Context, z, x *decimal.Big, y uint64) {
	z.SetUint64(1)
	for y != 0 {
		if y&1 != 0 {
			ctx.Mul(z, z, x)
			if !z.IsFinite() || z.Sign() == 0 ||
				z.Context.Conditions&decimal.Clamped != 0 {
				z.Context.Conditions |= decimal.Underflow | decimal.Subnormal
				break
			}
		}
		y >>= 1
		ctx.Mul(x, x, x)
	}
}
