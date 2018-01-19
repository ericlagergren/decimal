// +build !amd64

package arith

// Mul returns the 128-bit multiplication of x and y. Overflow checking can test
// hi == 0.
func Mul(x, y uint64) (hi, lo uint64) {
	const mask = 0xFFFFFFFF

	x1 := x & mask
	y1 := y & mask
	t := x1 * y1
	w3 := t & mask
	k := t >> 32

	x >>= 32
	t = (x * y1) + k
	k = (t & mask)
	w1 := t >> 32

	y >>= 32
	t = (x1 * y) + k

	hi = (x * y) + w1 + (t >> 32)
	lo = (t << 32) + w3
	return hi, lo
}
