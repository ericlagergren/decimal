package decimal

import (
	"bytes"
	"io"
	"math/big"

	"github.com/EricLagergren/decimal/internal/arith"
	"github.com/EricLagergren/decimal/internal/arith/checked"
	"github.com/EricLagergren/decimal/internal/c"
)

const debug = true

// ez returns true if z == 0.
func (z *Big) ez() bool {
	return z.Sign() == 0
}

// ltz returns true if z < 0
func (z *Big) ltz() bool {
	return z.Sign() < 0
}

// ltez returns true if z <= 0
func (z *Big) ltez() bool {
	return z.Sign() <= 0
}

// gtz returns true if z > 0
func (z *Big) gtz() bool {
	return z.Sign() > 0
}

// gtez returns true if z >= 0
func (z *Big) gtez() bool {
	return z.Sign() >= 0
}

// cmpNorm compares x and y in the range [0.1, 0.999...] and
// returns true if x > y.
func cmpNorm(x int64, xs int32, y int64, ys int32) (ok bool) {
	if debug && (x == 0 || y == 0) {
		panic("x and y cannot be zero")
	}
	if diff := xs - ys; diff != 0 {
		if diff < 0 {
			x, ok = checked.MulPow10(x, -diff)
		} else {
			y, ok = checked.MulPow10(y, diff)
		}
	}
	if x != c.Inflated {
		if y != c.Inflated {
			return arith.AbsCmp(x, y) > 0
		}
		return false
	}
	return true
}

// cmpNormBig compares x and y in the range [0.1, 0.999...] and
// returns true if x > y.
func cmpNormBig(x *big.Int, xs int32, y *big.Int, ys int32) (ok bool) {
	diff := xs - ys
	if diff < 0 {
		return checked.MulBigPow10(x, -diff).Cmp(y) > 0
	}
	return x.Cmp(checked.MulBigPow10(y, diff)) > 0
}

type buffer struct{ bytes.Buffer }

func (b *buffer) String() string {
	// Trim zeros.
	buf := b.Bytes()
	i := len(buf) - 1
	for ; i >= 0 && buf[i] == '0'; i-- {
	}
	if buf[i] == '.' {
		i--
	}
	b.Truncate(i + 1)
	return b.Buffer.String()
}

type writer interface {
	io.Writer
	io.ByteWriter
	WriteString(string) (int, error)
	String() string // don't import "fmt"
}
