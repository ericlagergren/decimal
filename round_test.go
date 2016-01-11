package decimal

import "testing"

func test(t *testing.T, fn func(float64) float64, set map[float64]float64) {
	for f, i := range set {
		if g := fn(f); g != i {
			t.Errorf("%.1f: want %f, got %f\n", f, i, g)
		}
	}
}

// func TestRoundUp(t *testing.T) {
// 	test(t, roundUp, map[float64]int64{
// 		5.5:  6,
// 		2.5:  3,
// 		1.6:  2,
// 		1.1:  1,
// 		1.0:  1,
// 		-1.0: -1,
// 		-1.1: -1,
// 		-1.6: -2,
// 		-2.5: -3,
// 		-5.5: -6,
// 	})
// }

// func TestRoundDown(t *testing.T) {
// 	test(t, roundDown, map[float64]int64{
// 		5.5:  5,
// 		2.5:  2,
// 		1.6:  1,
// 		1.1:  1,
// 		1.0:  1,
// 		-1.0: -1,
// 		-1.1: -1,
// 		-1.6: -1,
// 		-2.5: -2,
// 		-5.5: -5,
// 	})
// }

// func TestRoundCeiling(t *testing.T) {
// 	test(t, roundCeiling, map[float64]int64{
// 		5.5:  6,
// 		2.5:  3,
// 		1.6:  2,
// 		1.1:  2,
// 		1.0:  1,
// 		-1.0: -1,
// 		-1.1: -1,
// 		-1.6: -1,
// 		-2.5: -2,
// 		-5.5: -5,
// 	})
// }

// func TestRoundFloor(t *testing.T) {
// 	test(t, roundFloor, map[float64]int64{
// 		5.5:  5,
// 		2.5:  2,
// 		1.6:  1,
// 		1.1:  1,
// 		1.0:  1,
// 		-1.0: -1,
// 		-1.1: -2,
// 		-1.6: -2,
// 		-2.5: -3,
// 		-5.5: -6,
// 	})
// }

// func TestRoundHalfUp(t *testing.T) {
// 	test(t, roundHalfUp, map[float64]int64{
// 		5.5:  6,
// 		2.5:  3,
// 		1.6:  2,
// 		1.1:  1,
// 		1.0:  1,
// 		-1.0: -1,
// 		-1.1: -1,
// 		-1.6: -2,
// 		-2.5: -3,
// 		-5.5: -6,
// 	})
// }

// func TestRoundHalfDown(t *testing.T) {
// 	test(t, roundHalfDown, map[float64]int64{
// 		5.5:  5,
// 		2.5:  2,
// 		1.6:  2,
// 		1.1:  1,
// 		1.0:  1,
// 		-1.0: -1,
// 		-1.1: -1,
// 		-1.6: -2,
// 		-2.5: -2,
// 		-5.5: -5,
// 	})
// }

// func TestRoundHalfEven(t *testing.T) {
// 	test(t, round, map[float64]float64{
// 		5.5:   6,
// 		2.5:   2,
// 		1.6:   2,
// 		1.454: 1.5,
// 		1.1:   1,
// 		1.0:   1,
// 		-1.0:  -1,
// 		-1.1:  -1,
// 		-1.6:  -2,
// 		-2.5:  -2,
// 		-5.5:  -6,
// 	})
// }
