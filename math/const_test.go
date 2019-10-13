package math

import (
	"fmt"
	"testing"

	"github.com/ericlagergren/decimal"
)

// Test whether Set or Scan is faster for setting constants.

func BenchmarkSet(b *testing.B) {
	var lB decimal.Big
	var ctx decimal.Context
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Precision = i % 100
		ctx.Set(&lB, _E)
	}
	gB = &lB
}

func BenchmarkScan(b *testing.B) {
	const e = "2.718281828459045235360287471352662497757247093699959574966967627724076630353547594571382178525166427"
	var lB decimal.Big
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lB.SetString(e[:i%100])
	}
	gB = &lB
}

func BenchmarkPi(b *testing.B) {
	for _, prec := range benchPrecs {
		b.Run(fmt.Sprintf("%d", prec), func(b *testing.B) {
			z := decimal.WithPrecision(prec)
			for j := 0; j < b.N; j++ {
				Pi(z)
			}
			gB = z
		})
	}
}
