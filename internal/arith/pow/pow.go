package pow

import (
	"math"
	"math/big"
	"sync"

	"github.com/EricLagergren/decimal/internal/c"
)

const (
	Tab64Len  = 19
	ThreshLen = 19
)

var (
	pow10tab     [Tab64Len]int64
	thresholdTab [ThreshLen]int64
	bigPow10Tab  struct {
		sync.RWMutex
		x []big.Int
	}
)

// BigTen computes 10 ** n.
func BigTen(n int64) (p big.Int) {
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
	tableLen := int64(len(bigPow10Tab.x))
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

// Ten64 returns 10 ** e and a boolean indicating whether
// it fits into an int64.
func Ten64(e int64) (int64, bool) {
	if e < 0 {
		p, ok := Ten64(-e)
		// Otherwise division by zero.
		if !ok {
			return 0, false
		}
		return 1 / p, ok
	}
	if e < Tab64Len {
		return pow10tab[e], true
	}
	return 0, false
}

// Thresh returns ...
func Thresh(t int32) int64 {
	if t < 0 {
		return 1 / Thresh(-t)
	}
	return thresholdTab[t]
}

func init() {
	pow10tab[1] = 10
	for i := 2; i < Tab64Len; i++ {
		m := i / 2
		pow10tab[i] = pow10tab[m] * pow10tab[i-m]
	}

	thresholdTab[0] = math.MaxInt64
	for i := int64(1); i < ThreshLen; i++ {
		p, _ := Ten64(i)
		thresholdTab[i] = math.MaxInt64 / p
	}

	bigPow10Tab.x = make([]big.Int, Tab64Len)
	for i := int64(1); i < Tab64Len; i++ {
		p, _ := Ten64(i)
		bigPow10Tab.x[i] = *big.NewInt(p)
	}
}
