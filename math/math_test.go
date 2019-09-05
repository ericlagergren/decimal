package math_test

import (
	"testing"

	"github.com/ericlagergren/decimal/v4"
	"github.com/ericlagergren/decimal/v4/internal/test"
	"github.com/ericlagergren/decimal/v4/math"
)

func TestExp(t *testing.T)   { test.Exp.Test(t) }
func TestLog(t *testing.T)   { test.Log.Test(t) }
func TestLog10(t *testing.T) { test.Log10.Test(t) }
func TestPow(t *testing.T)   { test.Pow.Test(t) }
func TestSqrt(t *testing.T)  { test.Sqrt.Test(t) }

func TestIssue146(t *testing.T) {
	one := decimal.New(1, 0)
	for i := int64(0); i < 10; i++ {
		n := decimal.New(i, 1)
		firstPow := math.Pow(new(decimal.Big), n, one)
		if n.Cmp(firstPow) != 0 {
			t.Errorf("%v^1 != %v", n, firstPow)
		} else {
			t.Logf("%v^1 == %v", n, firstPow)
		}
	}
}
