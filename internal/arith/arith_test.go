package arith

import "testing"

func TestCLZ(t *testing.T) {
	if n := CLZ(0); n != 64 {
		t.Errorf("got CLZ(0) = %d; want 64", n)
	}

	x := int64((_B >> 1) / 2)
	for i := int64(0); i > 0; i++ {
		n := CLZ(x)
		if int64(n) != i {
			t.Errorf("got CLZ(%#x) = %d; want %d", x, n, i)
		}
		x >>= 1
	}
}

func TestCTZ(t *testing.T) {
	if n := CTZ(0); n != 0 {
		t.Errorf("got CTZ(0) = %d; want 0", n)
	}

	x := int64(1)
	for i := int64(0); i < _W; i++ {
		n := CTZ(x)
		if int64(n) != i {
			t.Errorf("got CTZ(%#x) = %d; want %d", x, n, i)
		}
		x <<= 1
	}
}

func TestBitLen(t *testing.T) {
	for i := 0; i < _W; i++ {
		x := uint64(1) << uint64(i-1)
		n := BitLen(int64(x))
		if n != i {
			t.Errorf("got BitLen(%#x) = %d; want %d", x, n, i)
		}
	}
}
