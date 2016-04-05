package decimal

// ez returns true if z == 0.
func (z *Big) ez() bool {
	return z.Sign() == 0
}

// ltz returns true if z < 0
func (z *Big) ltz() bool {
	return z.Sign() < 0
}

// ltez returns true if z <= 0
func (z *Big) ltez() bool {
	return z.Sign() <= 0
}

// gtz returns true if z > 0
func (z *Big) gtz() bool {
	return z.Sign() > 0
}

// gtez returns true if z >= 0
func (z *Big) gtez() bool {
	return z.Sign() >= 0
}
