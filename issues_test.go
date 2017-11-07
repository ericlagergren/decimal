package decimal

import (
	"math/big"
	"testing"

	"github.com/ericlagergren/decimal/internal/c"
)

func Test_Issue15(t *testing.T) {
	b1 := &Big{
		compact:  c.Inflated,
		scale:    5,
		Context:  Context{Precision: 0, RoundingMode: 0},
		form:     finite,
		unscaled: *new(big.Int).SetInt64(181050000),
	}
	b2 := &Big{
		compact: 18105,
		scale:   1,
		Context: Context{Precision: 0, RoundingMode: 0},
		form:    finite,
	}
	if b1.Cmp(b2) != 0 {
		t.Errorf("failed comparing %v with %v: %v", b1, b2, b1.Cmp(b2))
	}
}

func Test_Issue20(t *testing.T) {
	x := New(10240000000000, 0)
	x.Mul(x, New(976563, 9))
	if x.Int64() != 10000005120 {
		t.Error("error int64: ", x.Int64(), x.Int(nil).Int64())
	}
}
