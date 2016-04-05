package decimal

import (
	"math/big"

	"github.com/EricLagergren/decimal/internal/arith/pow"
	"github.com/EricLagergren/decimal/internal/c"
)

// Modf returns the decomposed integral and fractional parts of the
// value of x such that int + frac == x.
// Neither int nor frac will alias x's mantissa.
func (x *Big) Modf() (int *Big, frac *Big) {
	int = new(Big)
	frac = new(Big)

	if x.form == inf {
		int.form = inf
		frac.form = nan
		return int, frac
	}

	if x.form == nan {
		int.form = nan
		frac.form = nan
		return int, frac
	}

	// scale == 0
	int.ctx = x.ctx
	int.form = finite

	// Needs proper scale.
	frac.scale = x.scale
	frac.ctx = x.ctx
	frac.form = finite

	if x.IsInt() {
		if x.isCompact() {
			int.compact = x.compact
		} else {
			// int and frac cannot alias.
			// This should be faster than any other method...
			m := make([]big.Word, len(x.mantissa.Bits()))
			copy(m, x.mantissa.Bits())
			int.mantissa.SetBits(m)
			int.compact = c.Inflated

			// Bits sets |x| so manually correct negative values.
			if x.SignBit() {
				int.mantissa.Neg(&int.mantissa)
			}
		}
		return int, new(Big)
	}

	if x.isCompact() {
		i, f, ok := mod(x.compact, x.scale)
		if ok {
			int.compact, frac.compact = i, f
			return int, frac
		}
	}

	m := &x.mantissa
	// Possible fallthrough.
	if x.isCompact() {
		m = big.NewInt(x.compact)
	}
	i, f := modbig(m, x.scale)
	int.compact = c.Inflated
	frac.compact = c.Inflated
	int.mantissa.Set(i)
	frac.mantissa.Set(f)
	return int, frac
}

// mod splits f, a scaled decimal, into its integeral and fractional parts.
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

// modbig splits f, a scaled decimal, into its integeral and fractional parts.
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
