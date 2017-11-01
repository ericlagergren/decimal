package decimal

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/compat"
)

// cmpNorm compares x and y in the range [0.1, 0.999...] and returns true if x
// > y.
func cmpNorm(x int64, xs int, y int64, ys int) (ok bool) {
	goodx, goody := true, true

	// xs, ys > 0, so no overflow
	if diff := xs - ys; diff != 0 {
		if diff < 0 {
			x, goodx = checked.MulPow10(x, -uint64(diff))
		} else {
			y, goody = checked.MulPow10(y, uint64(diff))
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
func cmpNormBig(x *big.Int, xs int, y *big.Int, ys int) (ok bool) {
	if diff := xs - ys; diff != 0 {
		if diff < 0 {
			x = checked.MulBigPow10(new(big.Int).Set(x), uint64(-diff))
		} else {
			y = checked.MulBigPow10(new(big.Int).Set(y), uint64(diff))
		}
	}
	return compat.BigCmpAbs(x, y) > 0
}

// scalex adjusts x by scale. If scale < 0, x = x * 10^-scale, otherwise
// x = x / 10^scale.
func scalex(x int64, scale int32) (sx int64, ok bool) {
	if scale < 0 {
		sx, ok = checked.MulPow10(x, -uint64(scale))
		if !ok {
			return 0, false
		}
		return sx, true
	}
	p, ok := pow.TenInt(uint64(scale))
	if !ok {
		return 0, false
	}
	return x / p, true
}
