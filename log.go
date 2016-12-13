package decimal

import (
	"fmt"
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
		//	etx.Exp(x0)
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

var taylorSeq = [...]*Big{
	nil, nil,
	New(2, 0), New(3, 0), New(4, 0), New(5, 0),
	New(6, 0), New(7, 0), New(8, 0), New(9, 0),
	New(10, 0), New(11, 0), New(12, 0), New(13, 0),
	New(14, 0), New(15, 0), New(16, 0), New(17, 0),
}

// taylorFor returns a *Big for the given i value. The first 15 values are
// cached inside taylorSeq. 0 and 1 will return nil. The return value must not
// be modified. This prevents allocs inside the `taylor` method for at least
// the first 15 values.
func taylorFor(i int64) *Big {
	if i < int64(len(taylorSeq)) {
		return taylorSeq[i]
	}
	return New(i, 0)
}

// taylor sets z to e ** x using the Taylor series and returns z.
func (z *Big) taylor(x *Big) *Big {
	// Taylor series: Σ x^n/n!

	zz := alias(z, x).Set(x).Add(x, one) // x + 1, our sum

	var (
		fac        = New(1, 0)       // factorial
		xp         = new(Big).Set(x) // x power
		prev, term Big               // previous sum and current term
		zp         = z.ctx.prec()
	)

	for i := int64(2); ; i++ {
		// x ** n / n!
		term.Quo(xp.Mul(xp, x), fac.Mul(fac, taylorFor(i)))
		zz.Round(zp)
		prev.Set(zz)      // save current sum
		zz.Add(zz, &term) // sum += current term
		zz.Round(zp)

		// Convergence test.
		// http://math.stackexchange.com/a/281971/153292
		if zz.Cmp(&prev) == 0 {
			break
		}
	}
	return z.Set(zz)
}

var _ = fmt.Println

// pow sets d to x ** y and returns z.
func (z *Big) pow(x *Big, y *big.Int) *Big {
	switch {
	// 1 / (x ** -y)
	case y.Sign() < 0:
		return z.Quo(one, z.pow(x, new(big.Int).Neg(y)))
	// x ** 1 == x
	case y.Cmp(oneInt) == 0:
		return z.SetBigMantScale(y, 0)
	// 0 ** y == 0
	case x.form == 0:
		z.form = zero
		return z
	}

	x0 := new(Big).Set(x)
	y0 := new(big.Int).Set(y)
	ret := alias(z, x).SetMantScale(1, 0)
	var odd big.Int
	for y0.Sign() > 0 {
		if odd.And(y0, oneInt).Sign() != 0 {
			ret.Mul(ret, x0)
		}
		y0.Rsh(y0, 1)
		x0.Mul(x0, x0)
	}
	return z.Set(ret)
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

// pow sets z to x ** y and returns z.
func (z *Big) powInt(x *Big, y int64) *Big {
	switch {
	// 1 / (x ** -y)
	case y < 0:
		return z.Quo(one, z.powInt(x, -y))
	// x ** 1 == x
	case y == 1:
		return z.Set(x)
	// 0 ** y == 0
	case x.form == 0:
		z.form = zero
		return z
	}

	x0 := new(Big).Set(x)
	ret := alias(z, x).SetMantScale(1, 0)
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
		odd = new(big.Int).And(&x.unscaled, oneInt).Cmp(oneInt) == 0
	}
	return frac.ez() && odd
}
