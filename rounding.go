package decimal

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
)

func (z *Big) shouldInc(c int, pos, odd bool) bool {
	switch r := z.Context.RoundingMode; r {
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
		z.signal(InvalidContext, fmt.Errorf("invalid rounding mode: %d", r))
		return false
	}
}

func (z *Big) needsInc(x, r int64, pos, odd bool) bool {
	m := 1
	if r > math.MinInt64/2 || r <= math.MaxInt64/2 {
		m = arith.AbsCmp(r*2, x)
	}
	return z.shouldInc(m, pos, odd)
}

func (z *Big) needsIncBig(x, r *big.Int, pos, odd bool) bool {
	x0 := new(big.Int).Mul(r, twoInt)
	m := arith.BigAbsCmp(x0, x)
	return z.shouldInc(m, pos, odd)
}
