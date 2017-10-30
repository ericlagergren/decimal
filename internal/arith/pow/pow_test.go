package pow

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ericlagergren/decimal/internal/c"
)

func randPow(n int64) uint64 {
	return uint64(rand.Int63n(n))
}

func TestBigTen(t *testing.T) {
	for i := 0; i < 5000; i++ {
		p := randPow(BigTabLen * 2)
		comp := BigTen(p)
		n := new(big.Int).SetUint64(p)
		act := n.Exp(c.TenInt, n, nil)
		if act.Cmp(comp) != 0 {
			fmt.Println(BigTabLen)
			t.Fatalf("%d: got len of %d, want len of %d", p,
				len(comp.String()), len(act.String()))
		}
	}
}
