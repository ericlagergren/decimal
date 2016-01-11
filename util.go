package decimal

import (
	"math"
	"math/big"
)

// The default rounding should be unbiased rounding.
// It takes marginally longer than simply
//
// 		if f < 0 { return math.Ceil(f - 0.5) }
// 		return math.Floor(f + 0.5)
//
// But returns more accurate results.
func round(f float64) float64 {
	d, frac := math.Modf(f)
	if f > 0.0 && (frac > 0.5 || (frac == 0.5 && uint64(d)%2 != 0)) {
		return d + 1.0
	}
	if f < 0.0 && (frac < -0.5 || (frac == -0.5 && uint64(d)%2 != 0)) {
		return d - 1.0
	}
	return d
}

// Because of rounding issues.
// float64(8470964534836491162) == 8470964534836491264
// https://play.golang.org/p/mNPhjMhN_I
func abs(x int64) int64 {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0
	}
	return x
}

func prod(x, y int64) int64 {
	if x == 0 || y == 0 {
		return x * y
	}

	ax := abs(x)
	ay := abs(y)
	p := x * y
	if (ax|ay)>>31 == 0 || p/y == x {
		return p
	}
	return overflown
}

func sub(x, y int64) int64 {
	return sum(x, -y)
}

func sum(x, y int64) int64 {
	sum := x + y
	// Algorith from "Hacker's Delight" 2-12
	if (sum^x)&(sum^y) < 0 {
		return overflown
	}
	return sum
}

func safeScale(x, y, a int64) int64 {
	if a == overflown {
		if x != 0 {
			if y > 0 {
				panic("decimal: rescale underflow")
			}
			panic("decimal: rescale overflow")
		}
	}
	return a
}

func safeScale2(y, a int64) int64 {
	if a == overflown {
		if y > 0 {
			panic("decimal: rescale underflow")
		}
		panic("decimal: rescale overflow")
	}
	return a
}

// mulPow10 returns 10 * x ^ n
func mulPow10(x int64, n int64) int64 {
	if x == 0 || n <= 0 {
		return x
	}
	if x == overflown {
		return x
	}
	if n < pow10tabLen && n < thresholdLen {
		if x == 1 {
			return pow10int64(n)
		}
		if abs(n) < thresh(n) {
			return prod(x, pow10int64(n))
		}
	}
	return overflown
}

// mulBigPow10 returns 10 * x ^ n
func mulBigPow10(x *big.Int, n int64) *big.Int {
	if x.Sign() == 0 || n <= 0 {
		return x
	}
	b := bigPow10(n)
	return new(big.Int).Mul(x, &b)
}

// modi splits f, a scaled decimal, into its integral
// and fractional parts.
// It does not check for overflows.
func modi(f int64, scale int64) (dec int64, frac int64) {
	if f < 0 {
		dec, frac = modi(-f, scale)
		return -dec, -frac
	}
	exp := pow10int64(scale)
	if exp == overflown {
		return overflown, overflown
	}
	if exp == 0 {
		return f, 0
	}
	dec = f / exp
	frac = f - (dec * exp)
	return dec, frac
}

func modbig(b *big.Int, scale int64) (dec *big.Int, frac *big.Int) {
	if b.Sign() < 0 {
		dec, frac = modbig(new(big.Int).Neg(b), scale)
		dec.Neg(dec)
		frac.Neg(frac)
		return dec, frac
	}
	exp := bigPow10(scale)
	if exp.Cmp(zeroInt) == 0 {
		return b, new(big.Int)
	}
	dec = new(big.Int).Quo(b, &exp)
	frac = new(big.Int).Sub(b, new(big.Int).Mul(dec, &exp))
	return dec, frac
}

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func cmpBigAbs(x, y big.Int) int {
	// SetBits sets the absolute value, thus causing an absolute comparison.
	x0 := new(big.Int).SetBits(x.Bits())
	y0 := new(big.Int).SetBits(y.Bits())
	return x0.Cmp(y0)
}

func cmpAbs(x, y int64) int {
	if x < 0 {
		x = -x
	}
	if y < 0 {
		y = -y
	}

	switch {
	case x > y:
		return +1
	case x < y:
		return -1
	default:
		return 0
	}
}

func clz(x int64) (n int64)
func ctz(x int64) (n int64)
