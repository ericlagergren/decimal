#include "textflag.h"

// func clz(x int64) (n int64)
TEXT ·clz(SB),NOSPLIT,$0
	BSRQ 	x+0(FP), AX
	JZ 		Z1
	XORQ	$63, AX
	MOVQ 	AX, n+8(FP)
	RET

Z1:	MOVQ 	$0, n+8(FP)
	RET

// func ctz(x int64) (n int64)
TEXT ·ctz(SB),NOSPLIT,$0
	BSRQ 	x+0(FP), AX
	JZ 		Z1
	ADDQ	$1, AX
	MOVQ 	AX, n+8(FP)
	RET

Z1:	MOVQ 	$0, n+8(FP)
	RET
