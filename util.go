package decimal

import (
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	"github.com/ericlagergren/decimal/internal/arith/pow"
	"github.com/ericlagergren/decimal/internal/compat"
)

func precision(x *Big) (p int) {
	p = x.Context.Precision
	if p == 0 {
		return DefaultPrecision
	}
	if p < 0 && p != UnlimitedPrecision {
		p = -p
	}
	return p
}

func mode(x *Big) OperatingMode { return x.Context.OperatingMode }

// copybits can be useful when we want to allocate a big.Int without calling
// new or big.Int.Set. For example:
//
//   var x big.Int
//   if foo {
//       x.SetBits(copybits(y.Bits()))
//   }
//   ...
//
func copybits(x []big.Word) []big.Word {
	z := make([]big.Word, len(x))
	copy(z, x)
	return z
}

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
