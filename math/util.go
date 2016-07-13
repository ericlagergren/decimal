package math

import (
	"github.com/EricLagergren/decimal"
	"github.com/EricLagergren/decimal/internal/arith/checked"
)

// An ErrNaN panic is raised by a Decimal operation that would lead to a NaN
// under IEEE-754 rules. An ErrNaN implements the error interface.
type ErrNaN struct {
	msg string
}

func (e ErrNaN) Error() string {
	return e.msg
}

func shiftRadixRight(x *decimal.Big, n int) {
	ns, ok := checked.Sub32(x.Scale(), int32(n))
	if !ok {
		panic(ok)
	}
	x.SetScale(ns)
}
