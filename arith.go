package decimal

func abs(x int64) int64 {
	mask := -int64(uint64(x) >> 63)
	return (x + mask) ^ mask
}
