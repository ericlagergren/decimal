package math

import (
	"math/big"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/internal/arith/pow"
)

// Modf decomposes x into its integral and fractional parts such that int +
// frac == x, sets z to the integral part (such that z aliases int) and returns
// both parts.
func Modf(z, x *decimal.Big) (int *decimal.Big, frac *decimal.Big) {
	if x.Sign() == 0 {
		return z.SetMantScale(0, 0), decimal.New(0, 0)
	}

	if x.IsInf(0) {
		z.SetInf(x.IsInf(+1))
		panic(decimal.ErrNaN{"modf of an infinity"})
	}

	// x is an integer—we can just set z to x.
	if x.IsInt() {
		return z.Set(x), decimal.New(0, 0)
	}

	ctx := x.Context()
	scale := x.Scale()

	int = z.SetContext(ctx)
	frac = new(decimal.Big).SetContext(ctx)

	xc, xu := decimal.Raw(x)
	if !x.IsBig() {
		i, f, ok := mod(xc, scale)
		if ok {
			return z.SetMantScale(i, 0), frac.SetMantScale(f, scale)
		}
		// Fallthrough.
		xu = big.NewInt(xc)
	}
	i, f := modbig(xu, scale)
	return z.SetBigMantScale(i, 0), frac.SetBigMantScale(f, scale)
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
