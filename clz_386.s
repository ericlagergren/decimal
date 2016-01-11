#include "textflag.h"

// func clz(x int64) (n int64)
TEXT ·clz(SB),NOSPLIT,$0
	BSRL 	x+0(FP), AX
	JZ 		Z1
	XORL	$32, AX
	MOVL 	AX, n+4(FP)
	RET

Z1:	MOVL 	$1, n+4(FP)
	RET

// func ctz(x int64) (n int64)
TEXT ·ctz(SB),NOSPLIT,$0
	BSRL 	x+0(FP), AX
	JZ 		Z1
	ADDL	$1, AX
	MOVL 	AX, n+4(FP)
	RET

Z1:	MOVL 	$1, n+4(FP)
	RET
