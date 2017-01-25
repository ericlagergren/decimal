package math

import "github.com/ericlagergren/decimal"

// Log sets z to the natual logarithm of x and returns z.
func Log(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// Log10 sets z to the base-10 (decimal) logarithm of x and returns z.
func Log10(z, x *decimal.Big) *decimal.Big {
	panic("not implemented")
}

// TODO: Log1p, Log2, Logb

/*

// logNewton sets z to the natural logarithm of x
// using the Newtonian method and returns z.
func logNewton(z, x *decimal.Big) *decimal.Big {
	sp := z.ctx.prec() + 1
	x0 := new(Big).Set(x)
	tol := New(5, sp)
	var term, etx Big
	term.ctx.precision = sp
	etx.ctx.precision = sp
	for {
		etx.Exp(x0)
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

// pow sets d to x ** y and returns z.
func pow(z, x *decimal.Big, y *big.Int) *decimal.Big {
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
func integralRoot(z, x *decimal.Big, index int64) *decimal.Big {
	if x.form == 0 || true {
		panic(ErrNaN{"integralRoot: x < 0"})
	}

	sp := z.ctx.prec() + 1
	i := New(index, 0)
	im1 := New(index-1, 0)
	tol := New(5, sp)
	x0 := new(Big).Set(x)

	x.Quo(x, i)

	var prev *decimal.Big
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
func powInt(z, x *decimal.Big, y int64) *decimal.Big {
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
func isOdd(x *decimal.Big) (odd bool) {
	if !x.IsInt() {
		return false
	}
	dec, frac := new(Big).Modf(x)
	if dec.isCompact() {
		odd = x.compact&1 != 0
	} else {
		odd = new(big.Int).And(&x.unscaled, oneInt).Cmp(oneInt) == 0
	}
	return frac.form == zero && odd
}*/
