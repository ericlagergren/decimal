// +build go1.9

package compat

import "math/big"

func IsInt64(x *big.Int) bool { return x.IsInt64() }
