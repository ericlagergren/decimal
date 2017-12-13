package math

import (
	"math/big"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/c"
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

	if y.IsInt() {
		if y.IsInt() {
			if v, ok := y.Uint64(); ok {
				return powToCompact(z, x, v)
			}
		}
		return powToInflated(z, x, y.Int(nil))
	}
	return powToBig(z, x, y)
}

const maxint = int(^uint(0) >> 1)

func powOverflows(z, x, y *decimal.Big) bool {
	absx := abs(x)

	Θ := adjusted(y)
	Z := lowerBoundZ(absx)
	if Z == maxint {
		z.Context.Conditions |= decimal.InsufficientStorage
		return true
	}

	neg := x.Signbit() && y.IsInt() && y.Int(nil).Bit(0) != 0
	if adjusted(absx) < 0 != y.Signbit() {
		const Ω = 10 // arith.Length(decimal.MaxScale)
		if Ω < Z+Θ {
			z.SetMantScale(1, c.MaxScaleInf)
			if neg {
				misc.CopyNeg(z, z)
			}
			// decimal.Test(z)
			return true
		}
	} else {
		tiny := etiny(z)
		if Ω := arith.Length(arith.Abs(int64(tiny))); Ω < Z+Θ {
			z.SetMantScale(0, -(tiny - 1))
			if neg {
				misc.CopyNeg(z, z)
			}
			// decimal.Test(z)
			return true
		}
	}
	return false
}

func lowerBoundZ(x *decimal.Big) int {
	t := adjusted(x)
	if t > 0 {
		// x >= 10 = floor(log10(floor(abs(log10(x)))))
		return arith.Length(uint64(t)) - 1
	}
	if t < -1 {
		// x < 1/10 = floor(log10(floor(abs(log10(x)))))
		return arith.Length(uint64(-(t + 1))) - 1
	}
	var tmp decimal.Big
	tmp.Sub(x, one)
	if !tmp.IsFinite() {
		return maxint
	}
	u := adjusted(&tmp)
	if t == 0 {
		return u - 2
	}
	return u - 1
}

func powToCompact(z, x *decimal.Big, y uint64) *decimal.Big {
	x0 := decimal.WithContext(x.Context).Copy(x)
	z.SetUint64(1)
	for y != 0 {
		if y&1 != 0 {
			z.Mul(z, x0)
			if !z.IsFinite() || z.Sign() == 0 {
				break
			}
		}
		y >>= 1
		x0.Mul(x0, x0)
		if x0.IsNaN(0) {
			z.Context.Conditions |= x0.Context.Conditions
			return z.SetNaN(false)
		}
	}
	return z
}

func powToInflated(z, x *decimal.Big, y *big.Int) *decimal.Big {
	prec := precision(z) + arith.BigLength(y) + 2

	y0 := new(big.Int).Abs(y)
	var x0 *decimal.Big
	if y.Sign() < 0 {
		prec++
		x0 = decimal.WithPrecision(prec).Quo(one, x)
	} else {
		x0 = decimal.WithPrecision(prec).Copy(x)
	}

	oldm := z.Context.RoundingMode
	oldp := z.Context.Precision
	z.Context.RoundingMode = decimal.ToNearestEven
	z.Context.Precision = prec
	z.SetUint64(1)
	for y0.Sign() != 0 {
		if y0.Bit(0) != 0 {
			z.Mul(z, x0)
			if !z.IsFinite() || z.Sign() == 0 {
				break
			}
		}
		y0.Rsh(y0, 1)
		x0.Mul(x0, x0)
		if x0.IsNaN(0) {
			z.Context.Conditions |= x0.Context.Conditions
			z.Context.RoundingMode = oldm
			z.Context.Precision = oldp
			return z.SetNaN(false)
		}
	}
	z.Context.RoundingMode = oldm
	z.Context.Precision = oldp
	return decimal.ToNearestEven.Round(z, oldp)
}

func powToBig(z, x, y *decimal.Big) *decimal.Big {
	z0 := decimal.WithPrecision(precision(z))
	return Exp(z, z.Mul(y, Log(z0, x)))
}
