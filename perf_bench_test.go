package decimal

import (
	"math/big"
	"math/rand"
	"testing"
)

var bigInt = new(big.Int)

var even = big.NewInt(2)
var randInt = big.NewInt(rand.Int63())

func BenchmarkAnd(b *testing.B) {
	r := new(big.Int).Set(randInt)
	for i := 0; i < b.N; i++ {
		bigInt.And(r, oneInt)
	}
}

func BenchmarkRem(b *testing.B) {
	r := new(big.Int).Set(randInt)
	for i := 0; i < b.N; i++ {
		bigInt.Rem(r, even)
	}
}
