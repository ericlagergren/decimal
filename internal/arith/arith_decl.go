// +build amd64

package arith

// in arith_amd64.s

// CLZ returns the number of leading zeros in x.
func CLZ(x int64) (n int)

// BitLen returns the number of bits required to hold x.
func BitLen(x int64) (n int)
