package checked

import (
	"math"
	"math/big"
	"sync"

	"github.com/EricLagergren/decimal/internal/c"
)

const (
	pow10tab64Len = 19
	thresholdLen  = 19
	pow10tab32Len = 8
)

var (
	pow10tab     [pow10tab64Len]int64
	thresholdTab [thresholdLen]int64
	bigPow10Tab  struct {
		sync.RWMutex
		x []big.Int
	}
)

// bigPow10 computes 10 ** n.
func bigPow10(n int32) (p big.Int) {
	if n < 0 {
		return p
	}

	bigPow10Tab.RLock()
	if int(n) < len(bigPow10Tab.x) {
		p = bigPow10Tab.x[n]
		bigPow10Tab.RUnlock()
		return p
	}

	// If we don't have a constraint on the length of the big powers table
	// we could very well end up trying to eat up 32 bits of
	// space (because our scale is that big).
	// To keep from making silly mistkes like that, keep the slice's
	// size at something reasonable.
	if n > 1e5 {
		p.Exp(c.TenInt, big.NewInt(int64(n)), nil)
		bigPow10Tab.RUnlock()
		return p
	}

	// Drop the RLock in order to properly Lock it.
	bigPow10Tab.RUnlock()

	// Expand our table to contain the value for 10 ** n.
	bigPow10Tab.Lock()
	tableLen := int32(len(bigPow10Tab.x))
	newLen := tableLen << 1
	for newLen <= n {
		newLen *= 2
	}
	for i := tableLen; i < newLen; i++ {
		prev := bigPow10Tab.x[i-1]
		pow := new(big.Int).Mul(&prev, c.TenInt)
		bigPow10Tab.x = append(bigPow10Tab.x, *pow)
	}
	p = bigPow10Tab.x[n]
	bigPow10Tab.Unlock()
	return p
}

// pow10 is a wrapper around math.Pow10.
func pow10(e int32) float64 {
	// Should be lossless.
	return math.Pow10(int(e))
}

func pow10int32(e int32) (int32, bool) {
	if e < 0 {
		p, ok := pow10int32(-e)
		return 1 / p, ok
	}
	if e < pow10tab32Len {
		p, ok := pow10int64(int64(e))
		return int32(p), ok
	}
	return c.BadScale, false
}

// pow10int64 returns 10 ** e and a boolean indicating whether
// it fits into an int64.
func pow10int64(e int64) (int64, bool) {
	if e < 0 {
		p, ok := pow10int64(-e)
		return 1 / p, ok
	}
	if e < pow10tab64Len {
		return pow10tab[e], true
	}
	return c.Inflated, false
}

// thresh returns ...
func thresh(t int32) int64 {
	if t < 0 {
		return 1 / thresh(-t)
	}
	return thresholdTab[t]
}

func init() {
	pow10tab[1] = 10
	for i := 2; i < pow10tab64Len; i++ {
		m := i / 2
		pow10tab[i] = pow10tab[m] * pow10tab[i-m]
	}

	thresholdTab[0] = math.MaxInt64
	for i := int64(1); i < thresholdLen; i++ {
		p, _ := pow10int64(i)
		thresholdTab[i] = math.MaxInt64 / p
	}

	bigPow10Tab.x = make([]big.Int, pow10tab64Len)
	for i := int64(1); i < pow10tab64Len; i++ {
		p, _ := pow10int64(i)
		bigPow10Tab.x[i] = *big.NewInt(p)
	}
}
