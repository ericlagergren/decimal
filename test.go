// +build ignore

package main

import (
	"fmt"

	. "github.com/EricLagergren/decimal"
)

func main() {
	p := New(1234, 3)
	q := New(987654123, 5)
	p.SetContext(Context{Prec: 2})
	p.Hypot(p, q)
	fmt.Println(p)
}
