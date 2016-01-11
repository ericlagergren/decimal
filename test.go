// +build ignore

package main

import (
	"fmt"
	"math/big"

	. "github.com/EricLagergren/decimal"
)

func main() {
	a := New(1234, 3)
	a0 := uint64((a.BitLen() + 1) >> 1)
	fmt.Println(a0)
	fmt.Println(a.Rsh(a, a0))

	b := big.NewInt(1234)
	b0 := uint((b.BitLen() + 1) >> 1)
	fmt.Println(b0)
	fmt.Println(b.Rsh(b, b0))
}
