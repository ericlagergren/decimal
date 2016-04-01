#include "textflag.h"

// func CLZ(x int64) (n int)
TEXT ·CLZ(SB),NOSPLIT,$0
	B ·clz(SB)

// func CTZ(x int64) (n int)
TEXT ·CTZ(SB),NOSPLIT,$0
	B ·ctz(SB)

// func BitLen(x int64) (n int)
TEXT ·BitLen(SB),NOSPLIT,$0
	MOVW	x+0(FP), R0
	CLZ 	R0, R0
	RSB		$32, R0
	MOVW	R0, n+4(FP)
	RET
