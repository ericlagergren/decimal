package decimal

import (
	"math"
	"math/big"
	"sync"
)

const (
	pow10tabLen  = 19
	thresholdLen = 19
)

var (
	pow10tab     [pow10tabLen]int64
	thresholdTab [thresholdLen]int64
	bigPow10Tab  struct {
		sync.RWMutex
		x []big.Int
	}
)

// bigPow10 is returns big.Int(10 ** n)
func bigPow10(n int64) big.Int {
	if n < 0 {
		return big.Int{}
	}

	bigPow10Tab.RLock()
	if n < int64(len(bigPow10Tab.x)) {
		p := bigPow10Tab.x[n]
		bigPow10Tab.RUnlock()
		return p
	}

	// If we don't have a constraint on the length of the big powers table
	// we could very well end up trying to eat up the entire 64-bit address
	// space because our scale will stretch that big.
	// To keep from making silly mistkes like that, cap the array size at
	// something reasonable.
	const maxMemory = 1e5
	if n > maxMemory {
		var b big.Int
		b.Exp(tenInt, big.NewInt(n), nil)
		bigPow10Tab.RUnlock()
		return b
	}

	// Drop the RLock in order to properly Lock it.
	bigPow10Tab.RUnlock()

	// Expand our table to contain the value for 10 ** n.
	bigPow10Tab.Lock()
	tableLen := int64(len(bigPow10Tab.x))
	newLen := tableLen << 1
	for newLen <= n {
		newLen *= 2
	}
	for i := tableLen; i < newLen; i++ {
		prev := bigPow10Tab.x[i-1]
		pow := new(big.Int).Mul(&prev, tenInt)
		bigPow10Tab.x = append(bigPow10Tab.x, *pow)
	}
	p := bigPow10Tab.x[n]
	bigPow10Tab.Unlock()
	return p
}

// pow10 is a wrapper around math.Pow10 so we aren't cluttering
// our code with math.Pow10(int64(x))
func pow10(e int64) float64 {
	// Should be lossless.
	return math.Pow10(int(e))
}

// isOdd returns true if d is odd.
func (z *Decimal) isOdd() bool {
	dec, frac := Modf(z)
	return frac.ez() && dec.and(dec, 1).Equals(one)
}

// integralRoot sets d to the integral root of x and returns z.
func (z *Decimal) integralRoot(x *Decimal, index int64) *Decimal {
	if x.ltz() {
		panic("decimal.integralRoot: x < 0")
	}

	sp := z.Prec() + 1
	i := New(index, 0)
	im1 := New(index-1, 0)
	tol := New(5, sp)
	x0 := new(Decimal).Set(x)

	x.Div(x, i)

	var prev *Decimal
	var xx, xtoi1, xtoi, num, denom Decimal
	for {
		xtoi1.pow(x, im1).SetPrec(sp)
		xtoi.Mul(x, &xtoi1).SetPrec(sp)
		num.Add(x, new(Decimal).Mul(im1, &xtoi)).SetPrec(sp)
		denom.Mul(i, &xtoi1).SetPrec(sp)
		prev = x0
		x0.Div(&num, &denom)
		if xx.Sub(x0, prev).Abs(&xx).LessThanEQ(tol) {
			break
		}
	}
	*z = *x0
	return z
}

// pow sets d to x ** y and returns z.
func (z *Decimal) pow(x, y *Decimal) *Decimal {
	switch {
	case y.ltz(), (x.ez() || y.ez()):
		return z.SetInt64(1)
	case y.Equals(one):
		return z.Set(x)
	case x.ez():
		if x.isOdd() {
			return z.Set(x)
		}
		return z.SetInt64(0)
	}

	x0 := new(Decimal).Set(x).SetPrec(z.Prec())
	y0 := new(Decimal).Set(y)
	ret := New(1, 0)
	for y0.gtz() {
		if y0.isOdd() {
			ret.Mul(ret, x0).SetPrec(z.Prec())
		}
		x0.Mul(x0, x0)
		y0.rsh(y0, 1)
	}
	*z = *ret
	return ret
}

// pow sets d to x ** y and returns z.
func (z *Decimal) powInt(x *Decimal, y int64) *Decimal {
	switch {
	case y < 0, (x.ez() || y == 0):
		return z.SetInt64(1)
	case y == 1:
		return z.Set(x)
	case x.ez():
		if x.isOdd() {
			return z.Set(x)
		}
		return z.SetInt64(0)
	}

	x0 := new(Decimal).Set(x).SetPrec(z.Prec())
	ret := New(1, 0)
	for y > 0 {
		if y&1 == 1 {
			ret.Mul(ret, x0).SetPrec(z.Prec())
		}
		x0.Mul(x0, x0)
		y >>= 1
	}
	*z = *ret
	return ret
}

// pow10int64 returns 10**e if it fits into an int64.
// Otherwise, overflown.
func pow10int64(e int64) int64 {
	if e < 0 {
		return 1 / pow10int64(-e)
	}
	if e < pow10tabLen {
		return pow10tab[e]
	}
	return overflown
}

// thresh returns ...
func thresh(t int64) int64 {
	if t < 0 {
		return 1 / thresh(-t)
	}
	return thresholdTab[t]
}

func init() {
	pow10tab[0] = 0
	pow10tab[1] = 10
	for i := 2; i < pow10tabLen; i++ {
		m := i / 2
		pow10tab[i] = pow10tab[m] * pow10tab[i-m]
	}

	thresholdTab[0] = math.MaxInt64
	for i := int64(1); i < thresholdLen; i++ {
		thresholdTab[i] = math.MaxInt64 / pow10int64(i)
	}

	bigPow10Tab.x = make([]big.Int, 16)
	bigPow10Tab.x[0] = big.Int{}
	for i := int64(1); i < 16; i++ {
		bigPow10Tab.x[i] = *big.NewInt(pow10int64(i))
	}
}
