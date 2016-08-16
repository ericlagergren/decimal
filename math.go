package decimal

import (
	"math/big"

	"github.com/EricLagergren/decimal/internal/arith/pow"
	"github.com/EricLagergren/decimal/internal/c"
)

// Modf decomposes x into its integral and fractional parts such that int +
// frac == x, sets z to the integral part, and returns the integral and
// fractional parts.
func (z *Big) Modf(x *Big) (int *Big, frac *Big) {
	int = z
	frac = new(Big)

	if x.form == zero {
		z.form = zero
		frac.form = zero
		return z, frac
	}

	if x.form == inf {
		z.form = inf
		frac.form = inf
		return z, frac
	}

	z.ctx = x.ctx
	z.form = finite

	// Needs proper scale.
	// Set frac before z in case z aliases x.
	frac.scale = x.scale
	frac.ctx = x.ctx
	frac.form = finite

	if x.IsInt() {
		if x.isCompact() {
			z.compact = x.compact
		} else {
			z.mantissa.Set(&x.mantissa)
		}
		z.scale = 0
		return z, frac
	}

	if x.isCompact() {
		i, f, ok := mod(x.compact, x.scale)
		if ok {
			z.compact, frac.compact = i, f
			z.scale = 0
			return z, frac
		}
	}

	m := &x.mantissa
	// Possible fallthrough.
	if x.isCompact() {
		m = big.NewInt(x.compact)
	}
	i, f := modbig(m, x.scale)
	z.compact = c.Inflated
	frac.compact = c.Inflated
	z.mantissa.Set(i)
	frac.mantissa.Set(f)
	z.scale = 0
	return z, frac
}

// mod splits fr, a scaled decimal, into its integeral and fractional parts.
func mod(fr int64, scale int32) (dec int64, frac int64, ok bool) {
	if fr < 0 {
		dec, frac, ok = mod(-fr, scale)
		return -dec, -frac, ok
	}
	exp, ok := pow.Ten64(int64(scale))
	if !ok {
		return 0, 0, false
	}
	if exp == 0 {
		return fr, 0, true
	}
	dec = fr / exp
	frac = fr - (dec * exp)
	return dec, frac, true
}

// modbig splits b, a scaled decimal, into its integeral and fractional parts.
func modbig(b *big.Int, scale int32) (dec *big.Int, frac *big.Int) {
	if b.Sign() < 0 {
		dec, frac = modbig(new(big.Int).Neg(b), scale)
		dec.Neg(dec)
		frac.Neg(frac)
		return dec, frac
	}
	exp := pow.BigTen(int64(scale))
	if exp.Sign() == 0 {
		return b, new(big.Int)
	}
	dec = new(big.Int).Quo(b, &exp)
	frac = new(big.Int).Mul(dec, &exp)
	frac = frac.Sub(b, frac)
	return dec, frac
}
