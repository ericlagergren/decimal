package decimal

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith/pow"
)

// Modf decomposes x into its integral and fractional parts such that int +
// frac == x, sets z to the integral part (such that z aliases int) and returns
// both parts.
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

	ctx := x.ctx
	scale := x.scale

	// x is an integer—we can just set z to x.
	if x.IsInt() {
		frac.form = zero
		return z.Set(x), frac
	}

	if x.isCompact() {
		i, f, ok := mod(x.compact, scale)
		if ok {
			return z.SetMantScale(i, 0).SetContext(ctx),
				frac.SetMantScale(f, scale).SetContext(ctx)
		}
	}

	m := &x.unscaled
	// Possible fallthrough from 'ok'.
	if x.isCompact() {
		m = big.NewInt(x.compact)
	}
	i, f := modbig(m, scale)
	return z.SetBigMantScale(i, 0).SetContext(ctx),
		frac.SetBigMantScale(f, scale).SetContext(ctx)
}

// mod splits fr, a scaled decimal, into its integeral and fractional parts.
// The returned bool will be false if the scale is too large for 64-bit
// arithmetic.
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
