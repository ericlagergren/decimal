// Package pow implements basic power functions.
package pow

import (
	"math/big"
	"sync"

	"github.com/ericlagergren/decimal/internal/c"
)

const Tab64Len = 19

var (
	pow10tab    [Tab64Len]int64
	bigPow10Tab struct {
		sync.RWMutex
		x []*big.Int
	}
)

// BigTen computes 10 ** n. The returned *big.Int must not be modified.
func BigTen(n int64) *big.Int {
	if n < 0 {
		return new(big.Int)
	}

	// If we don't have a constraint on the length of the big powers table we
	// could very well end up trying to eat up 32 bits of space (because our
	// scale is 32 bits). To keep from making silly mistkes like that, cap the
	// slice's size at something reasonable.
	if n > 1e5 {
		return new(big.Int).Exp(c.TenInt, big.NewInt(n), nil)
	}

	bigPow10Tab.RLock()
	if int(n) < len(bigPow10Tab.x) {
		p := bigPow10Tab.x[n]
		bigPow10Tab.RUnlock()
		return p
	}

	// We need to expand our table to contain the value for 10 ** n.

	// We need to drop our read lock so we can lock the table for writing.
	bigPow10Tab.RUnlock()

	bigPow10Tab.Lock()
	defer bigPow10Tab.Unlock()

	// However, we need to look again: another thread could have came in and
	// resized the table for us.
	if int(n) < len(bigPow10Tab.x) {
		return bigPow10Tab.x[n]
	}

	// In the clear. Double the table size.
	tableLen := int64(len(bigPow10Tab.x))
	newLen := tableLen * 2
	for newLen <= n {
		newLen *= 2
	}
	for i := tableLen; i < newLen; i++ {
		prev := bigPow10Tab.x[i-1]
		pow := new(big.Int).Mul(prev, c.TenInt)
		bigPow10Tab.x = append(bigPow10Tab.x, pow)
	}
	return bigPow10Tab.x[n]
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

func init() {
	pow10tab[1] = 10
	for i := 2; i < Tab64Len; i++ {
		m := i / 2
		pow10tab[i] = pow10tab[m] * pow10tab[i-m]
	}

	bigPow10Tab.x = make([]*big.Int, Tab64Len)
	for i := int64(0); i < Tab64Len; i++ {
		p, _ := Ten64(i)
		bigPow10Tab.x[i] = big.NewInt(p)
	}
}
