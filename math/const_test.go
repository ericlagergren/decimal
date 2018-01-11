package math_test

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

var gB *decimal.Big

var _E, _ = new(decimal.Big).SetString("2.718281828459045235360287471352662497757247093699959574966967627724076630353547594571382178525166427")

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
