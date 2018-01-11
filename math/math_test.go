package math_test

import (
	"testing"

	"github.com/ericlagergren/decimal/internal/test"
)

func TestExp(t *testing.T)   { test.Exp.Test(t) }
func TestLog(t *testing.T)   { test.Log.Test(t) }
func TestLog10(t *testing.T) { test.Log10.Test(t) }
func TestPow(t *testing.T)   { test.Pow.Test(t) }
func TestSqrt(t *testing.T)  { test.Sqrt.Test(t) }
