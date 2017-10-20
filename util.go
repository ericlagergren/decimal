package decimal

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/internal/arith/pow"
)

const debug = true

// cmpNorm compares x and y in the range [0.1, 0.999...] and returns true if x
// > y.
func cmpNorm(x int64, xs int32, y int64, ys int32) (ok bool) {
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
		x = checked.MulBigPow10(new(big.Int).Set(x), -diff)
	} else {
		y = checked.MulBigPow10(new(big.Int).Set(y), diff)
	}
	return arith.BigAbsCmp(x, y) > 0
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
