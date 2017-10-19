package decimal

import (
	"fmt"
	"math"
	"math/big"
	"sync"

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

var intPool = sync.Pool{New: func() interface{} { return new(big.Int) }}

func get() *big.Int { return intPool.Get().(*big.Int) }

func getInt(x *big.Int) *big.Int { return get().Set(x) }
func getInt64(x int64) *big.Int  { return get().SetInt64(x) }

func putInt(b *big.Int) { intPool.Put(b) }

// cmpNorm compares x and y in the range [0.1, 0.999...] and returns true if x
// > y.
func cmpNorm(x int64, xs int32, y int64, ys int32) (ok bool) {
	if debug && (x == 0 || y == 0) {
		panic("x and/or y cannot be zero")
	}
	goodx, goody := true, true
	if diff := xs - ys; diff != 0 {
		if diff < 0 {
			x, goodx = checked.MulPow10(x, -diff)
		} else {
			y, goody = checked.MulPow10(y, diff)
		}
	}
	if goodx {
		if goody {
			return arith.AbsCmp(x, y) > 0
		}
		return false
	}
	return true
}

// cmpNormBig compares x and y in the range [0.1, 0.999...] and returns true if
// x > y.
func cmpNormBig(x *big.Int, xs int32, y *big.Int, ys int32) (ok bool) {
	if diff := xs - ys; diff < 0 {
		x = checked.MulBigPow10(getInt(x), -diff)
		defer putInt(x)
	} else {
		y = checked.MulBigPow10(getInt(y), diff)
		defer putInt(y)
	}
	return arith.BigAbsCmp(x, y) > 0
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

// TODO(eric): use math.Round when 1.10 lands.

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
func scalex(x int64, scale int32) (sx int64, ok bool) {
	if scale < 0 {
		sx, ok = checked.MulPow10(x, -scale)
		if !ok {
			return 0, false
		}
		return sx, true
	}
	p, ok := pow.Ten64(int64(scale))
	if !ok {
		return 0, false
	}
	return x / p, true
}
