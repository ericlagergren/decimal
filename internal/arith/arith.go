// +build go1.9

package arith

import "math/bits"

func CLZ(x int64) int    { return bits.LeadingZeros64(uint64(x)) }
func BitLen(x int64) int { return bits.Len64(uint64(x)) }
