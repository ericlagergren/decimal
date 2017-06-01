package decimal

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/c"
)

const debug = true

var (
	_ fmt.Stringer  = (*Big)(nil)
	_ fmt.Formatter = (*Big)(nil)
)

// alias returns a if a != b, otherwise it returns a newly-allocated Big. It
// should be used if a *might* be able to be used for storage, but only if it
// doesn't b. The returned Big will have a's Context.
func alias(a, b *Big) *Big {
	if a != b {
		return a
	}
	z := new(Big)
	z.Context = a.Context
	return z
}

// cmpNorm compares x and y in the range [0.1, 0.999...] and returns true if x
// > y.
func cmpNorm(x int64, xs int32, y int64, ys int32) (ok bool) {
	if debug && (x == 0 || y == 0) {
		panic("x and/or y cannot be zero")
	}
	if diff := xs - ys; diff != 0 {
		if diff < 0 {
			x, ok = checked.MulPow10(x, -diff)
		} else {
			y, ok = checked.MulPow10(y, diff)
		}
	}
	if x != c.Inflated {
		if y != c.Inflated {
			return arith.AbsCmp(x, y) > 0
		}
		return false
	}
	return true
}

// cmpNormBig compares x and y in the range [0.1, 0.999...] and returns true if
// x > y.
func cmpNormBig(x *big.Int, xs int32, y *big.Int, ys int32) (ok bool) {
	diff := xs - ys
	if diff < 0 {
		x1 := new(big.Int).Set(x)
		return checked.MulBigPow10(x1, -diff).Cmp(y) > 0
	}
	y1 := new(big.Int).Set(y)
	return x.Cmp(checked.MulBigPow10(y1, diff)) > 0
}

// findScale determines the precision of a float64.
func findScale(f float64) (precision int32) {
	switch {
	case f == 0.0, math.Floor(f) == f:
		return 0
	case math.IsNaN(f), math.IsInf(f, 0):
		return c.BadScale
	}

	e := float64(1)
	p := int32(0)
	for {
		e *= 10
		p++
		cmp := round(f*e) / e
		if math.IsNaN(cmp) || cmp == f {
			break
		}
	}
	return p
}

// The default rounding should be unbiased rounding.
// It takes marginally longer than
//
// 		if f < 0 {
// 			return math.Ceil(f - 0.5)
// 		}
// 		return math.Floor(f + 0.5)
//
// But returns more accurate results.
func round(f float64) float64 {
	d, frac := math.Modf(f)
	if f > 0.0 && (frac > +0.5 || (frac == 0.5 && uint64(d)%2 != 0)) {
		return d + 1.0
	}
	if f < 0.0 && (frac < -0.5 || (frac == -0.5 && uint64(d)%2 != 0)) {
		return d - 1.0
	}
	return d
}

// "stolen" from https://golang.org/pkg/math/big/#Rat.SetFloat64
// Removed non-finite case because we already check for
// Inf/NaN values
func bigIntFromFloat(f float64) *big.Int {
	const expMask = 1<<11 - 1
	bits := math.Float64bits(f)
	mantissa := bits & (1<<52 - 1)
	exp := int((bits >> 52) & expMask)
	if exp == 0 { // denormal
		exp -= 1022
	} else { // normal
		mantissa |= 1 << 52
		exp -= 1023
	}

	shift := 52 - exp

	// Optimization (?): partially pre-normalise.
	for mantissa&1 == 0 && shift > 0 {
		mantissa >>= 1
		shift--
	}

	if shift < 0 {
		shift = -shift
	}

	var a big.Int
	a.SetUint64(mantissa)
	return a.Lsh(&a, uint(shift))
}

// scalex adjusts x by scale. If scale < 0, x = x * 10^-scale, otherwise
// x = x / 10^scale.
func scalex(x int64, scale int32) (int64, bool) {
	if scale < 0 {
		x, ok := checked.MulPow10(x, -scale)
		if !ok {
			return 0, false
		}
		return x, true
	}
	p, ok := pow.Ten64(int64(scale))
	if !ok {
		return 0, false
	}
	return x / p, true
}
