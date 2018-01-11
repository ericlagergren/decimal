// +build ignore

package main

import (
	"fmt"
	gmath "math"

	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
)

func main() {
	z := new(decimal.Big)
	x, _ := new(decimal.Big).SetString("123.456")
	math.Exp(z, x)
	fmt.Println("g:", gmath.Exp(123.456))
	fmt.Println("d:", z)
}
