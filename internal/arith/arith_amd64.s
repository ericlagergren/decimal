#include "textflag.h"

// func CLZ(x int64) (n int)
TEXT ·CLZ(SB), NOSPLIT, $0
	MOVQ    x+0(FP), DI
	BSRQ    DI, AX
	MOVQ    $64, DX
	XORQ    $63, AX
	TESTQ   DI, DI
	CMOVQEQ DX, AX
	MOVQ    AX, n+8(FP)
	RET

// func BitLen(x int64) (n int)
TEXT ·BitLen(SB), NOSPLIT, $0
	BSRQ x+0(FP), AX
	JZ   Z1
	ADDQ $1, AX
	MOVQ AX, n+8(FP)
	RET

Z1:
	MOVQ $0, n+8(FP)
	RET
