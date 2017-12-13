package arith

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/ericlagergren/decimal/internal/c"
)

func TestBigTen(t *testing.T) {
	errc := make(chan error)
	workc := make(chan struct{}, 1000)
	for i := 0; i < 1000; i++ {
		workc <- struct{}{}
	}

	var wg sync.WaitGroup
	for i := uint64(0); i < BigPowTabLen+10; i++ {
		<-workc
		wg.Add(1)
		go func(i uint64) {
			comp := BigPow10(i)
			n := new(big.Int).SetUint64(i)
			act := n.Exp(c.TenInt, n, nil)
			if act.Cmp(comp) != 0 {
				cs, as := comp.String(), act.String()
				errc <- fmt.Errorf(`%d:
got   : (%d) %s
wanted: (%d) %s
`, i, len(cs), cs, len(as), as)
			}
			wg.Done()
			workc <- struct{}{}
		}(i)
	}

	donec := make(chan struct{})
	go func() {
		wg.Wait()
		donec <- struct{}{}
	}()

	for {
		select {
		case err := <-errc:
			t.Fatal(err.Error())
		case <-donec:
			return
		}
	}
}
