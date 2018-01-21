// +build amd64

#include "textflag.h"

// NOTE(eric): we get better performance by duplicating the function becase, as
// of Go 1.9, the compiler won't inline non-leaf functions. This may change in
// the future.

// func Mul128(x, y uint64) (z1, z0 uint64)
TEXT ·Mul128(SB),NOSPLIT,$0
	MOVQ x+0(FP), AX
	MULQ y+8(FP)
	MOVQ DX, z1+16(FP)
	MOVQ AX, z0+24(FP)
	RET

// func mulWW(x, y big.Word) (z1, z0 big.Word)
TEXT ·mulWW(SB),NOSPLIT,$0
	MOVQ x+0(FP), AX
	MULQ y+8(FP)
	MOVQ DX, z1+16(FP)
	MOVQ AX, z0+24(FP)
	RET
