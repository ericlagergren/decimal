// +build ignore

package main

import (
	"fmt"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
)

func main() {
	y := new(decimal.Big)
	x := decimal.New(4, 0)
	math.Exp(y, x)
	fmt.Println(y.String())
}
