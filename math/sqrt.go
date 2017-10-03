package math

import (
	"errors"
	"fmt"
	"math"

	"github.com/ericlagergren/checked"
	"github.com/ericlagergren/decimal"
)

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	p0 := new(decimal.Big).Set(p)
	q0 := new(decimal.Big).Set(q)
	if p0.Sign() <= 0 {
		p0.Neg(p0)
	}
	if q0.Sign() <= 0 {
		q0.Neg(q0)
	}
	if p0.Sign() == 0 {
		return z.SetMantScale(0, 0)
	}
	p0.Mul(p0, p0)
	q0.Mul(q0, q0)
	return Sqrt(z, p0.Add(p0, q0))
}

// Sqrt sets z to the square root of x and returns z. Sqrt will panic on
// negative values since decimal.Big cannot represent imaginary numbers.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	switch {
	case x.Signbit():
		signal(
			z,
			decimal.InvalidOperation,
			errors.New("math.Sqrt: cannot take square root of negative number"),
		)
	case x.IsNaN(true), x.IsNaN(false):
		signal(z, decimal.InvalidOperation, decimal.ErrNaN{"square root of NaN"})
	case x.IsInf(0):
		return z.SetInf(false)
	case x.Sign() == 0:
		return z.SetMantScale(0, 0)
	}

	// First fast path—check if x is a perfect square. If so, we can avoid
	// having to inflate x and can possibly use can use the hardware SQRT.
	// Note that we can only catch perfect squares that aren't big.Ints.
	if sq, ok := perfectSquare(x); ok {
		z.SetMantScale(sq, 0)
		return z
	}

	f := new(decimal.Big).Set(x)
	e := x.Scale()

	approx := alias(z, x)
	if e&1 == 0 {
		// approx := .259 + .819*f
		approx.Add(approx1, new(decimal.Big).Mul(approx2, f))
	} else {
		// f := f/10
		f.Quo(f, ten)
		// e := e + 1
		e++
		// approx := .0819 + 2.59*f
		approx.Add(approx3, new(decimal.Big).Mul(approx4, f))
	}

	p := int32(3)
	maxp := x.Context.Precision() + 2

	var tmp decimal.Big
	for {
		p = min(2*p-2, maxp)
		//approx.Context.SetPrecision(int32(p))
		// approx := .5*(approx + f/approx)
		tmp.Quo(f, approx)
		tmp.Add(approx, &tmp)
		approx.Mul(ptFive, &tmp)
		fmt.Println(approx)
		if p == maxp {
			break
		}
	}

	p = x.Context.Precision()
	approx.Context.SetPrecision(p + 2)
	return z.Set(approx).SetScale(e / 2)
}

func min(x, y int32) int32 {
	if x > y {
		return y
	}
	return x
}

var (
	approx1 = decimal.New(259, 3)
	approx2 = decimal.New(819, 3)
	approx3 = decimal.New(819, 4)
	approx4 = decimal.New(259, 2)
	ptFive  = decimal.New(5, 1)
)

// perfectSquare algorithm slightly partially borrowed from
// https://stackoverflow.com/a/295678/2967113
func perfectSquare(x *decimal.Big) (square int64, ok bool) {
	if x.IsBig() || !x.IsInt() {
		return 0, false
	}
	xc := x.Int64()
	h := xc & 0xF
	if h > 9 {
		return 0, false
	}
	switch h {
	case 0, 1, 4, 9:
		// "Show that floating point sqrt(x*x) >= x for all long x."
		// https://math.stackexchange.com/a/238885/153292
		tst := int64(math.Sqrt(float64(xc)))
		return tst, tst*tst == xc
	default:
		return 0, false
	}
}

func shiftRadixRight(x *decimal.Big, n int) {
	ns, ok := checked.Sub32(x.Scale(), int32(n))
	if !ok {
		panic(ok)
	}
	x.SetScale(ns)
}
