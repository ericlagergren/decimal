package decimal

import (
	"fmt"
	"math"
	"math/big"
)

// logNewton sets z to the natural logarithm of x
// using the Newtonian method and returns z.
func (z *Big) logNewton(x *Big) *Big {
	sp := z.ctx.prec() + 1
	x0 := new(Big).Set(x)
	tol := New(5, sp)
	var term, etx Big
	term.ctx.precision = sp
	etx.ctx.precision = sp
	for {
		etx.exp(x0)
		term.Sub(&etx, x)
		term.Quo(&term, &etx)
		x0.Sub(x0, &term)
		if term.Cmp(tol) < 0 {
			break
		}
	}
	*z = *x0
	return z
}

// exp sets z to e ** x and returns z.
func (z *Big) exp(x *Big) *Big {
	if x.ez() {
		z.ctx.precision = z.ctx.prec()
		return z.SetMantScale(1, 0)
	}

	if x.ltz() {
		x0 := new(Big).Set(x)
		// 1 / (e ** -x)
		return z.Quo(New(1, 0), x0.exp(x0.Neg(x0)))
	}

	dec, frac := z.Modf(x)
	if dec.ez() {
		return z.taylor(x)
	}

	o := New(1, 0)
	o.ctx.precision = z.ctx.prec()
	o.Add(z, frac.Quo(frac, dec))
	o.taylor(o)

	res := New(1, 0)
	for dec.Cmp(max64) >= 0 {
		res.Mul(res, new(Big).powInt(o, math.MaxInt64))
		dec.Sub(dec, max64)
	}
	return z.Mul(res, o.powInt(o, dec.Int64()))
}

// taylor sets z to e ** x using the Taylor series and returns z.
func (z *Big) taylor(x *Big) *Big {
	// Taylor series: Σ x^n/n!

	z.Add(z.Set(x), one) // 1 + x, our sum

	var (
		fac            = New(1, 0) // factorial
		xp, prev, term Big         // x power, previous sum, and current term
		zp             = z.ctx.prec()
	)
	xp.Set(x) // Stack allocation..?
	for i := int64(2); ; i++ {
		// x ^ n / n!
		term.Quo(xp.Mul(&xp, x), fac.Mul(fac, New(i, 0)))
		z.Round(zp)
		prev.Set(z)     // save current sum
		z.Add(z, &term) // sum += current term
		z.Round(zp)

		fmt.Println(z, &prev, zp)

		// Convergence test.
		// http://math.stackexchange.com/a/281971/153292
		if z.Cmp(&prev) == 0 {
			break
		}
	}
	return z
}

var _ = fmt.Println

// pow sets d to x ** y and returns z.
func (z *Big) pow(x *Big, y *big.Int) *Big {
	switch {
	case y.Sign() < 0, (x.ez() || y.Sign() == 0):
		return z.SetMantScale(1, 0)
	case y.Cmp(oneInt) == 0:
		return z.Set(x)
	case x.ez():
		if x.isOdd() {
			return z.Set(x)
		}
		z.form = zero
		return z
	}

	x0 := new(Big).Set(x)
	y0 := new(big.Int).Set(y)
	ret := New(1, 0)
	var odd big.Int
	for y0.Sign() > 0 {
		if odd.And(y0, oneInt).Sign() != 0 {
			ret.Mul(ret, x0)
		}
		y0.Rsh(y0, 1)
		x0.Mul(x0, x0)
	}
	*z = *ret
	return ret
}

// integralRoot sets d to the integral root of x and returns z.
func (z *Big) integralRoot(x *Big, index int64) *Big {
	if x.ltz() {
		panic(ErrNaN{"integralRoot: x < 0"})
	}

	sp := z.ctx.prec() + 1
	i := New(index, 0)
	im1 := New(index-1, 0)
	tol := New(5, sp)
	x0 := new(Big).Set(x)

	x.Quo(x, i)

	var prev *Big
	var xx, xtoi1, xtoi, num, denom Big
	for {
		xtoi1.powInt(x, index-1)
		xtoi.Mul(x, &xtoi1)
		num.Add(x, new(Big).Mul(im1, &xtoi))
		denom.Mul(i, &xtoi1)
		prev = x0
		x0.Quo(&num, &denom)
		if xx.Sub(x0, prev).Abs(&xx).Cmp(tol) <= 0 {
			break
		}
	}
	*z = *x0
	return z
}

// pow sets d to x ** y and returns z.
func (z *Big) powInt(x *Big, y int64) *Big {
	switch {
	case y < 0, (x.ez() || y == 0):
		return z.SetMantScale(1, 0)
	case y == 1:
		return z.Set(x)
	case x.ez():
		if x.isOdd() {
			return z.Set(x)
		}
		z.form = zero
		return z
	}

	x0 := new(Big).Set(x)
	ret := New(1, 0)
	for y > 0 {
		if y&1 == 1 {
			ret.Mul(ret, x0)
		}
		x0.Mul(x0, x0)
		y >>= 1
	}
	return z.Set(ret)
}

// isOdd returns true if d is odd.
func (x *Big) isOdd() (odd bool) {
	if !x.IsInt() {
		return false
	}
	dec, frac := new(Big).Modf(x)
	if dec.isCompact() {
		odd = x.compact&1 != 0
	} else {
		odd = new(big.Int).And(&x.mantissa, oneInt).Cmp(oneInt) == 0
	}
	return frac.ez() && odd
}
