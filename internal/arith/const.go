package arith

// These are needed for both < Go 1.9 and 1.9 testing.

const (
	_m    = 18446744073709551615 // ^uint64(0)
	_logS = _m>>8&1 + _m>>16&1 + _m>>32&1
	_S    = 1 << _logS

	_W = _S << 3 // word size in bits
	_B = 1 << _W // digit base
	_M = _B - 1  // digit mask
)
