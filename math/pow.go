package math

import (
	"fmt"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/misc"
)

// Pow sets z to x**y % m if m != nil and returns z.
func Pow(z, x, y, m *decimal.Big) *decimal.Big {
	// Pass x to the second call to CheckNaNs since the first argument cannot
	// be nil.
	if z.CheckNaNs(x, y) || z.CheckNaNs(x, m) {
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
			// ±Inf ** y = 0
			return z.SetUint64(0)
		case -1:
			// ±Inf ** -y = +Inf
			return z.SetInf(true)
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
		fmt.Println("A")
		return powInt(z, x, y)
	}
	fmt.Println("B")
	return powDec(z, x, y)
}

func powInt(z, x, y *decimal.Big) *decimal.Big {
	prec := precision(z)
	ctx := decimal.Context{Precision: prec + y.Precision() + 2}

	var x0 decimal.Big
	if y.Signbit() {
		ctx.Precision++
		ctx.Quo(&x0, one, x)
	} else {
		x0.Copy(x)
	}
	z.SetUint64(1)
	sign := x.Signbit()

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
	z.Context = decimal.Context{Precision: max(x.Precision(), precision(z)) + 4 + 19}
	Exp(z, z.Mul(y, Log(z, x)))
	if neg && z.IsFinite() {
		misc.CopyNeg(z, z)
	}
	z.Context = oc
	return oc.Round(z)
}
