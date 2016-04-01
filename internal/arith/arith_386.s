#include "textflag.h"

// func CLZ(x int) (n int)
TEXT ·CLZ(SB),NOSPLIT,$0
	BSRL 	x+0(FP), AX
	JZ 		Z1
	XORL	$32, AX
	MOVL 	AX, n+4(FP)
	RET

Z1:	MOVL 	$0, n+4(FP)
	RET

// func CTZ(x int) (n int)
TEXT ·CTZ(SB),NOSPLIT,$0
	BSRL 	x+0(FP), AX
	JZ 		Z1
	MOVL 	AX, n+4(FP)
	RET

Z1:	MOVL 	$0, n+4(FP)
	RET

// func BitLen(x in64) (n int)
TEXT ·BitLen(SB),NOSPLIT,$0
	BSRL 	x+0(FP), AX
	JZ 		Z1
	INCL 	AX
	MOVL 	AX, n+4(FP)
	RET

Z1:	MOVL 	$0, n+4(FP)
	RET

// The following functions are for testing.

// func clz_asm(x int) (n int)
TEXT ·clz_asm(SB),NOSPLIT,$0
	JMP ·CLZ(SB)

// func ctz_asm(x int) (n int)
TEXT ·ctz_asm(SB),NOSPLIT,$0
	JMP ·CTZ(SB)
