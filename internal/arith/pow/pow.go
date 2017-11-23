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
	pow10tab = [TabLen]uint64{
		0:  1,
		1:  10,
		2:  100,
		3:  1000,
		4:  10000,
		5:  100000,
		6:  1000000,
		7:  10000000,
		8:  100000000,
		9:  1000000000,
		10: 10000000000,
		11: 100000000000,
		12: 1000000000000,
		13: 10000000000000,
		14: 100000000000000,
		15: 1000000000000000,
		16: 10000000000000000,
		17: 100000000000000000,
		18: 1000000000000000000,
		19: 10000000000000000000,
	}
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
	if n >= BigTabLen {
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

	// n < BigTabLen

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

func Safe(e uint64) bool { return e < TabLen }

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
	// Can we move this into a var decl without copylock freaking out?
	storeBigTable(&[]*big.Int{
		0:  new(big.Int).SetUint64(1),
		1:  c.TenInt,
		2:  new(big.Int).SetUint64(100),
		3:  new(big.Int).SetUint64(1000),
		4:  new(big.Int).SetUint64(10000),
		5:  new(big.Int).SetUint64(100000),
		6:  new(big.Int).SetUint64(1000000),
		7:  new(big.Int).SetUint64(10000000),
		8:  new(big.Int).SetUint64(100000000),
		9:  new(big.Int).SetUint64(1000000000),
		10: new(big.Int).SetUint64(10000000000),
		11: new(big.Int).SetUint64(100000000000),
		12: new(big.Int).SetUint64(1000000000000),
		13: new(big.Int).SetUint64(10000000000000),
		14: new(big.Int).SetUint64(100000000000000),
		15: new(big.Int).SetUint64(1000000000000000),
		16: new(big.Int).SetUint64(10000000000000000),
		17: new(big.Int).SetUint64(100000000000000000),
		18: new(big.Int).SetUint64(1000000000000000000),
		19: new(big.Int).SetUint64(10000000000000000000),
	})
}
