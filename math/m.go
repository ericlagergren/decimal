// +build ignore

package main

import (
	"github.com/ericlagergren/decimal"
	"github.com/ericlagergren/decimal/math"
)

func main() {
	const N = 4950

	z1 := decimal.WithPrecision(N)
	math.E(z1)

	z2 := decimal.WithPrecision(N)
	e(z2)

	if z1.Cmp(z2) != 0 {
		panic("!= 0")
	}
}

var one = new(decimal.Big).SetUint64(1)

func e(z *decimal.Big) *decimal.Big {
	ctx := decimal.Context{Precision: z.Context.Precision}
	ctx.Precision += 4
	var (
		sum  = z.SetUint64(2)
		fac  = new(decimal.Big).SetUint64(1)
		term = new(decimal.Big)
		prev = new(decimal.Big)
	)

	for i := uint64(2); sum.Cmp(prev) != 0; i++ {
		// Use term as our intermediate storage for our factorial. SetUint64
		// should be marginally faster than ctx.Add(incr, incr, one), but either
		// the costly call to Quo makes it difficult to notice.
		term.SetUint64(i)
		ctx.Mul(fac, fac, term)
		ctx.Quo(term, one, fac)
		prev.Copy(sum)
		ctx.Add(sum, sum, term)
	}

	/*
		fmt.Println("fac", fac)
		fmt.Println("term", term)
		fmt.Println("sum", sum)
	*/

	ctx.Precision -= 4
	return ctx.Set(z, sum)
}
