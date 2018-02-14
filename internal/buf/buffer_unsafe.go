// +build !purego

package buf

import "unsafe"

func (b *B) String() string { return *(*string)(unsafe.Pointer(&b.buf)) }
