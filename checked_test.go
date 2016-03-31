package decimal

// Note: This does a horrid job of testing without the branch
// predictor interfering.

import (
	"math"
	"testing"

	"github.com/EricLagergren/decimal/internal/c"
)

// This could screw with escape analysis, but only returns
// a bool and requires one branch.
func _checkedAdd1(x, y int64, z *int64) bool {
	sum := x + y
	return (sum^x)&(sum^y) >= 0
}

// This requires two branches, but doesn't mess with escape analysis.
func _checkedAdd2(x, y int64) (sum int64) {
	sum = x + y
	if (sum^x)&(sum^y) < 0 {
		return c.Inflated
	}
	return sum
}

// This returns two values, but only requires one branch and does
// not mess with escape analysis.
func _checkedAdd3(x, y int64) (sum int64, ok bool) {
	sum = x + y
	return sum, (sum^x)&(sum^y) >= 0
}

var v int64
var x, y int64

func init() {
	x = 500
	y = math.MaxInt64
}

func Benchmark1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var n int64
		if !_checkedAdd1(x, y, &n) {
			v = n
		}
	}
}

func Benchmark2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := _checkedAdd2(x, y)
		if n == c.Inflated {
			v = n
		}
	}
}

func Benchmark3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n, ok := _checkedAdd3(x, y)
		if !ok {
			v = n
		}
	}
}
