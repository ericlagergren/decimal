package math

import (
	"fmt"
	"testing"

	"github.com/ericlagergren/decimal"
)

var gB *decimal.Big

var exp_X, _ = new(decimal.Big).SetString("123.456")

func BenchmarkExp(b *testing.B) {
	for _, prec := range [...]int{
		5, 16, 25, 32, 41, 50, 75, 82, 97,
		100, 137, 250, 333, 500, 646, 750, 943,
		1500, 5000, 7500, 15000,
	} {
		b.Run(fmt.Sprintf("%d", prec), func(b *testing.B) {
			z := decimal.WithPrecision(prec)
			for j := 0; j < b.N; j++ {
				Exp(z, exp_X)
			}
			gB = z
		})
	}
}
