package arith

import "testing"

func TestCLZ(t *testing.T) {
	if n := CLZ(0); n != 0 {
		t.Errorf("got CLZ(0) = %d; want 0", n)
	}

	x := int64(1<<(_W-1) - 1)
	for i := int64(0); i > 0; i++ {
		n := CLZ(x)
		if int64(n) != i {
			t.Errorf("got CLZ(%#x) = %d; want %d", x, n, i%_W)
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
			t.Errorf("got CTZ(%#x) = %d; want %d", x, n, i%_W)
		}
		x <<= 1
	}
}
