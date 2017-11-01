// +build go1.9

package arith

import "math/bits"

func LeadingZeros64(x uint64) int { return bits.LeadingZeros64(x) }
func Len64(x uint64) int          { return bits.Len64(x) }
func OnesCount32(x uint32) int    { return bits.OnesCount32(x) }
