package decimal

import (
	"math"
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
)

func (r RoundingMode) needsInc(c int, pos, odd bool) bool {
	switch r {
	case Unneeded:
		panic("decimal: rounding is necessary")
	case AwayFromZero:
		return true
	case ToZero:
		return false
	case ToPositiveInf:
		return pos
	case ToNegativeInf:
		return !pos
	case ToNearestEven, ToNearestAway:
		if c < 0 {
			return false
		}
		if c > 0 {
			return true
		}
		if r == ToNearestEven {
			return odd
		}
		return true
	default:
		panic("decimal: unknown RoundingMode")
	}
}

func (z *Big) needsInc(x, r int64, pos, odd bool) bool {
	m := 1
	if r > math.MinInt64/2 || r <= math.MaxInt64/2 {
		m = arith.AbsCmp(r<<1, x)
	}
	return z.Context.rmode.needsInc(m, pos, odd)
}

func (z *Big) needsIncBig(x, r *big.Int, pos, odd bool) bool {
	var x0 big.Int
	m := arith.BigAbsCmp(*x0.Mul(r, twoInt), *x)
	return z.Context.rmode.needsInc(m, pos, odd)
}
