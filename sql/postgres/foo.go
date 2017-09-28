// +build ignore

package main

import (
	"fmt"

	"github.com/ericlagergren/decimal"
)

func main() {
	x, ok := new(decimal.Big).SetString("+999999999999999999999999999999999999999999999999999999999999999999999999999999999999.999e+33000000")
	if !ok {
		panic(ok)
	}
	x.Round(4)
	fmt.Printf("%e\n", x)
	fmt.Println(x.String())
}
