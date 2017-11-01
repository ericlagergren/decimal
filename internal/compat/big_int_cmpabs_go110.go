// +build go1.10

package compat

import "math/big"

func BigCmpAbs(x, y *big.Int) int { return x.CmpAbs(y) }
