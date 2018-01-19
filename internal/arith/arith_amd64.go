// +build amd64

package arith

// Mul returns the 128-bit multiplication of x and y. Overflow checking can test
// hi == 0.
func Mul(x, y uint64) (z1, z0 uint64)
