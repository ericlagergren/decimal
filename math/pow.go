package math

import (
	"fmt"
	"math/big"

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

	if powOverflows(z, x, y) {
		return z
	}

	return powToBig(z, x, y)
	if y.IsInt() {
		if v, ok := y.Uint64(); ok {
			fmt.Println("A")
			return powToCompact(z, x, v)
		}
		fmt.Println("B")
		return powToInflated(z, x, y.Int(nil))
	}
	fmt.Println("C")
	panic("k")
}

func powOverflows(z, x, y *decimal.Big) bool {
	return false
}

func powToCompact(z, x *decimal.Big, y uint64) *decimal.Big {
	prec := precision(z)
	ctx := decimal.Context{Precision: prec + 2}

	x0 := new(decimal.Big).Copy(x)
	z.SetUint64(1)
	for y != 0 {
		if y&1 != 0 {
			ctx.Mul(z, z, x0)
			if !z.IsFinite() || z.Sign() == 0 {
				break
			}
		}
		y >>= 1
		ctx.Mul(x0, x0, x0)
		if x0.IsNaN(0) {
			z.Context.Conditions |= x0.Context.Conditions
			return z.SetNaN(false)
		}
	}
	ctx.Precision = prec
	return ctx.Round(z)
}

func powToInflated(z, x *decimal.Big, y *big.Int) *decimal.Big {
	prec := precision(z)
	ctx := decimal.Context{Precision: prec + 2}

	y0 := new(big.Int).Abs(y)
	var x0 *decimal.Big
	if y.Sign() < 0 {
		ctx.Precision++
		x0 = ctx.Quo(new(decimal.Big), one, x)
	} else {
		x0 = new(decimal.Big).Copy(x)
	}

	z.SetUint64(1)
	for y0.Sign() != 0 {
		if y0.Bit(0) != 0 {
			ctx.Mul(z, z, x0)
			if !z.IsFinite() || z.Sign() == 0 {
				break
			}
		}
		y0.Rsh(y0, 1)
		ctx.Mul(x0, x0, x0)
		if x0.IsNaN(0) {
			z.Context.Conditions |= x0.Context.Conditions
			return z.SetNaN(false)
		}
	}
	ctx.Precision = prec
	return ctx.Round(z)
}

func powToBig(z, x, y *decimal.Big) *decimal.Big {
	if z == y {
		y = new(decimal.Big).Copy(y)
	}
	neg := x.Signbit()
	if neg {
		x = misc.CopyAbs(new(decimal.Big), x)
	}
	Log(z, x)
	z.Mul(y, z)
	Exp(z, z)
	if neg && z.IsFinite() {
		misc.CopyNeg(z, z)
	}
	return z
}
