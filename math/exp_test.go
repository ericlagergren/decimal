package math

import (
	"fmt"
	"testing"

	"github.com/ericlagergren/decimal"
)

var exp_X, _ = new(decimal.Big).SetString("123.456")

func BenchmarkExp(b *testing.B) {
	for _, prec := range benchPrecs {
		b.Run(fmt.Sprintf("%d", prec), func(b *testing.B) {
			z := decimal.WithPrecision(prec)
			for j := 0; j < b.N; j++ {
				Exp(z, exp_X)
			}
			gB = z
		})
	}
}
