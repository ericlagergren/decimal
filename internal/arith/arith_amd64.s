#include "textflag.h"

// func CLZ(x int64) (n int)
TEXT ·CLZ(SB),NOSPLIT,$0
	BSRQ 	x+0(FP), AX
	JZ 		Z1
	XORQ	$63, AX
	MOVQ 	AX, n+8(FP)
	RET

Z1:	MOVQ 	$0, n+8(FP)
	RET

// func BitLen(x int64) (n int)
TEXT ·BitLen(SB),NOSPLIT,$0
	BSRQ	x+0(FP), AX
	JZ 		Z1
	ADDQ 	$1, AX
	MOVQ	AX, n+8(FP)
	RET

Z1: MOVQ	$0, n+8(FP)
	RET

// func CTZ(x int64) (n int)
TEXT ·CTZ(SB),NOSPLIT,$0
	// Faster to JMP to Go code than it is to run `ctz_asm`
	JMP ·ctz(SB)

// The following functions are for testing.

// func clz_asm(x int64) (n int)
TEXT ·clz_asm(SB),NOSPLIT,$0
	JMP ·CLZ(SB)

// func ctz_asm(x int64) (n int)
TEXT ·ctz_asm(SB),NOSPLIT,$0
	BSRQ 	x+0(FP), AX
	JZ 		Z1
	MOVQ 	AX, n+8(FP)
	RET

Z1:	MOVQ 	$0, n+8(FP)
	RET
