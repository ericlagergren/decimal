package arith

import "testing"

func testclz(t *testing.T, f func(int64) int) {
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

func TestCLZAsm(t *testing.T) {
	testclz(t, clz_asm)
}

func TestCLZGo(t *testing.T) {
	testclz(t, clz)
}

func testctz(t *testing.T, f func(int64) int) {
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

func TestCTZAsm(t *testing.T) {
	testctz(t, ctz_asm)
}

func TestCTZGo(t *testing.T) {
	testctz(t, ctz)
}

var globalCC int

// _W - BitLen instead
func BenchmarkCLZAsm2(b *testing.B) {
	var cc int
	for i := 0; i < b.N; i++ {
		cc = clz_asm2(int64(i))
	}
	globalCC = cc
}

// in arith_*.s
func BenchmarkCLZAsm(b *testing.B) {
	var cc int
	for i := 0; i < b.N; i++ {
		cc = clz_asm(int64(i))
	}
	globalCC = cc
}

// in arith_*.s
func BenchmarkCTZAsm(b *testing.B) {
	var cc int
	for i := 0; i < b.N; i++ {
		cc = ctz_asm(int64(i))
	}
	globalCC = cc
}

// in arith.go
func BenchmarkCLZGo(b *testing.B) {
	var cc int
	for i := 0; i < b.N; i++ {
		cc = clz(int64(i))
	}
	globalCC = cc
}

// in arith.go
func BenchmarkCTZGo(b *testing.B) {
	var cc int
	for i := 0; i < b.N; i++ {
		cc = ctz(int64(i))
	}
	globalCC = cc
}

// Definitions for testing.

func clz_asm(x int64) (n int)
func ctz_asm(x int64) (n int)

func clz_asm2(x int64) (n int) {
	return _W - bitlen(x)
}
