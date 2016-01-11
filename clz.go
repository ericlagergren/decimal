// +build !amd64,!368

package decimal

import "fmt"

func clz(x int64) (n int64) {
	fmt.Println("hello")
	for ; x >= 0x8000; x >>= 16 {
		n += 16
	}
	if x >= 0x80 {
		x >>= 8
		n += 8
	}
	if x >= 0x8 {
		x >>= 4
		n += 4
	}
	if x >= 0x2 {
		x >>= 2
		n += 2
	}
	if x >= 0x1 {
		n++
	}
	return n
}

func ctz(x int64) (n int64) {
	return uint(deBruijn64Lookup[((x&-x)*(deBruijn64&_M))>>58])
}
