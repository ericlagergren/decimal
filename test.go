// +build ignore

package main

import (
	"fmt"

	"github.com/EricLagergren/decimal"
)

func main() {
	v1 := decimal.New(1, 0)
	v2 := decimal.New(10, 0)
	v3 := decimal.New(100, 0)

	v2.Quo(v1, v2) // 1 / 10 == 0.1
	v3.Quo(v3, v1) // 100 / 1 == 100
	fmt.Println(v2, v3)
}
