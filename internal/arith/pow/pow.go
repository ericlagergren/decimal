// Package pow implements basic power functions.
package pow

import (
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ericlagergren/decimal/internal/c"
)

const (
	// TabLen is the largest cached power for integers.
	TabLen = 20

	// BigTabLen is the largest cached power for *big.Ints.
	BigTabLen = 1e5
)

var (
	pow10tab    [TabLen]uint64
	bigMu       sync.Mutex // protects writes to bigPow10Tab
	bigPow10Tab atomic.Value
)

func loadBigTable() []*big.Int    { return *(bigPow10Tab.Load().(*[]*big.Int)) }
func storeBigTable(x *[]*big.Int) { bigPow10Tab.Store(x) }

// BigTen computes 10 ** n. The returned *big.Int must not be modified.
func BigTen(n uint64) *big.Int {
	tab := loadBigTable()

	tabLen := uint64(len(tab))
	if n < tabLen {
		return tab[n]
	}

	// Too large for our table.
	if n > BigTabLen {
		// Optimization: we don't need to start from scratch each time. Start
		// from the largest term we've found so far.
		partial := tab[tabLen-1]
		p := new(big.Int).SetUint64(n - (tabLen - 1))
		return p.Mul(partial, p.Exp(c.TenInt, p, nil))
	}
	return growBigTen(n, tab)
}

func growBigTen(n uint64, tab []*big.Int) *big.Int {
	// We need to expand our table to contain the value for 10 ** n.
	bigMu.Lock()

	// n >= BigTabLen

	tableLen := uint64(len(tab))
	newLen := tableLen * 2
	for newLen <= n {
		newLen *= 2
	}
	if newLen > BigTabLen {
		newLen = BigTabLen
	}
	for i := tableLen; i < newLen; i++ {
		tab = append(tab, new(big.Int).Mul(tab[i-1], c.TenInt))
	}

	storeBigTable(&tab)
	bigMu.Unlock()
	return tab[n]
}

// Ten returns 10 ** e and a boolean indicating whether the result fits into
// an uint64.
func Ten(e uint64) (uint64, bool) {
	if e < TabLen {
		return pow10tab[e], true
	}
	return 0, false
}

// Ten returns 10 ** e and a boolean indicating whether the result fits into
// an int64.
func TenInt(e uint64) (int64, bool) {
	if e < TabLen-1 {
		return int64(pow10tab[e]), true
	}
	return 0, false
}

func init() {
	pow10tab[0] = 1
	pow10tab[1] = 10

	tab := make([]*big.Int, TabLen)
	tab[0] = c.OneInt
	tab[1] = c.TenInt
	for i := 2; i < TabLen; i++ {
		m := i / 2
		pow10tab[i] = pow10tab[m] * pow10tab[i-m]
		tab[i] = new(big.Int).SetUint64(pow10tab[i])
	}
	// Set first power of 10 so our calculations don't have to handle that case.
	storeBigTable(&tab)
}
