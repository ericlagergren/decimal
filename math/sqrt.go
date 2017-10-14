package math

import (
	"errors"
	"math"

	"github.com/ericlagergren/decimal"
)

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	switch {
	case p.IsInf(0), q.IsInf(0):
		return z.SetInf(true)
	case p.IsNaN(true), p.IsNaN(false), q.IsNaN(true), q.IsNaN(false):
		return z.SetNaN(true)
	}
	p0 := new(decimal.Big).Mul(p, p)
	q0 := new(decimal.Big).Mul(q, q)
	return Sqrt(z, z.Add(p0, q0))
}

// Sqrt sets z to the square root of x and returns z.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	if xs := x.Sign(); xs <= 0 {
		if xs == 0 {
			return z.SetMantScale(0, 0)
		}
		z.SetNaN(false)
		return signal(z,
			decimal.InvalidOperation,
			errors.New("math.Sqrt: cannot take square root of negative number"),
		)
	}
	if snan := x.IsNaN(true); snan || x.IsNaN(false) {
		x.SetNaN(snan)
		return signal(z,
			decimal.InvalidOperation, decimal.ErrNaN{"square root of NaN"})
	}
	if x.IsInf(1) {
		return z.SetInf(false)
	}

	prec := z.Context.Precision()

	// Fast path #1: use math.Sqrt if our decimal is small enough. 0 and 22
	// are implementation details of the Float64 method. If Float64 is altered,
	// change them.
	if prec <= 16 && (x.Scale() >= 0 && x.Scale() < 22) {
		return z.SetFloat64(math.Sqrt(x.Float64())).Round(prec)
	}

	// Fast path #2: x is a small perfect square.
	if x.IsInt() && !x.IsBig() {
		// https://stackoverflow.com/a/295678/2967113
		switch xc := x.Int64(); xc & 0xF {
		case 0, 1, 4, 9:
			// "Show that floating point sqrt(x*x) >= x for all long x."
			// https://math.stackexchange.com/a/238885/153292
			sqrt := int64(math.Sqrt(float64(xc)))
			if sqrt*sqrt == xc {
				return z.SetMantScale(sqrt, 0)
			}
		}
	}

	// Source for the following algorithm:
	//
	//  T. E. Hull and A. Abrham. 1985. Properly rounded variable precision
	//  square root. ACM Trans. Math. Softw. 11, 3 (September 1985), 229-237.
	//  DOI: https://doi.org/10.1145/214408.214413

	var (
		// TODO(eric): there's a narrowing here that might not always be correct
		// if x.Precision() > 1<<32 - 1. In practice, this is unlikely. But
		// theoretically there could be a decimal number with a _lot_ of digits.
		xprec = int32(x.Precision())

		// The algorithm requires a normalized ``f ∈ [0.1, 1)'' Of the two ways
		// to normalize f, adjusting its scale is the quickest. However, it then
		// requires us to increment approx's scale by e/2 instead of simply
		// setting it to e/2.
		f = new(decimal.Big).Copy(x).SetScale(xprec)

		// It also means we have to adjust e to equal out the sale adjustment.
		e = xprec - x.Scale()

		tmp    decimal.Big
		approx = alias(z, x)
	)

	if e&1 == 0 {
		approx.Add(approx1, tmp.Mul(approx2, f)) // approx := .259 + .819f
	} else {
		f.Quo(f, ten)                            // f := f/10
		e++                                      // e := e + 1
		approx.Add(approx3, tmp.Mul(approx4, f)) // approx := 0.819 + 2.59*f
	}

	var (
		maxp       = prec + 2
		p    int32 = 3
	)

	for p < maxp {
		// p := min(2*p - 2, maxp)
		if p = 2*p - 2; p > maxp {
			p = maxp
		}
		// precision p
		tmp.Context.SetPrecision(p)
		// approx := .5*(approx + f/approx)
		approx.Mul(ptFive, tmp.Add(approx, tmp.Quo(f, approx)))
	}

	// The paper also specifies an additional code block for adjusting approx.
	// This code never went into the branches that modified approx, and rounding
	// to half even does the same thing. The GDA spec requires us to use
	// rounding mode half even (speleotrove.com/decimal/daops.html#refsqrt)
	// anyway.

	approx.Context.RoundingMode = decimal.ToNearestEven
	return z.Set(approx.SetScale(approx.Scale() - e/2).Round(prec))
}

var (
	approx1 = decimal.New(259, 3)
	approx2 = decimal.New(819, 3)
	approx3 = decimal.New(819, 4)
	approx4 = decimal.New(259, 2)
	ptFive  = decimal.New(5, 1)
)
