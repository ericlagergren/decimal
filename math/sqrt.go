package math

import (
	"math"

	"github.com/ericlagergren/decimal"
)

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	if z.CheckNaNs(p, q) {
		return z
	}
	prec := precision(z) + 1
	p0 := decimal.WithPrecision(prec).Mul(p, p)
	q0 := p0
	if p.Cmp(q) != 0 {
		q0 = decimal.WithPrecision(prec).Mul(q, q)
	}
	return Sqrt(z, decimal.WithPrecision(prec).Add(p0, q0))
}

// Sqrt sets z to the square root of x and returns z.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	if z.CheckNaNs(x, nil) {
		return z
	}

	ideal := -((-x.Scale() - (-x.Scale() & 1)) / 2)
	if xs := x.Sign(); xs <= 0 {
		if xs == 0 {
			return z.SetMantScale(0, ideal).CopySign(z, x)
		}
		// errors.New("math.Sqrt: cannot take square root of negative number"),
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	// Already checked for negative numbers.
	if x.IsInf(+1) {
		return z.SetInf(false)
	}

	prec := precision(z)

	// Fast path #1: use math.Sqrt if our decimal is small enough.
	if f, exact := x.Float64(); exact && prec <= 15 {
		return z.SetFloat64(math.Sqrt(f))
	}

	// Source for the following algorithm:
	//
	//  T. E. Hull and A. Abrham. 1985. Properly rounded variable precision
	//  square root. ACM Trans. Math. Softw. 11, 3 (September 1985), 229-237.
	//  DOI: https://doi.org/10.1145/214408.214413

	var (
		xprec = x.Precision()

		// The algorithm requires a normalized ``f ∈ [0.1, 1)'' Of the two ways
		// to normalize f, adjusting its scale is the quickest. However, it then
		// requires us to increment approx's scale by e/2 instead of simply
		// setting it to e/2.
		f = decimal.WithPrecision(prec).Copy(x).SetScale(xprec)
		e = -x.Scale() + xprec

		tmp    = decimal.WithPrecision(prec)
		approx = decimal.WithPrecision(prec)
	)

	if e&1 == 0 {
		approx.FMA(approx2, f, approx1) // approx := .259 + .819f
	} else {
		f.SetScale(f.Scale() + 1)       // f := f/10
		e++                             // e := e + 1
		approx.FMA(approx4, f, approx3) // approx := .0819 + 2.59f
	}

	var (
		maxp     = prec + 5 // extra prec to skip weird +/- 0.5 adjustments
		p    int = 3
	)

	for {
		// p := min(2*p - 2, maxp)
		if p = 2*p - 2; p > maxp {
			p = maxp
		}
		// precision p
		tmp.Context.Precision = p
		approx.Context.Precision = p
		// approx := .5*(approx + f/approx)
		approx.Mul(ptFive, tmp.Add(approx, tmp.Quo(f, approx)))
		if p == maxp {
			break
		}
	}

	// The paper also specifies an additional code block for adjusting approx.
	// This code never went into the branches that modified approx, and rounding
	// to half even does the same thing. The GDA spec requires us to use
	// rounding mode half even (speleotrove.com/decimal/daops.html#refsqrt)
	// anyway.

	approx.SetScale(approx.Scale() - e/2)
	decimal.ToNearestEven.Round(approx, prec)
	z.Context.Conditions |= approx.Context.Conditions
	if ideal != approx.Scale() {
		z.Context.Conditions |= decimal.Rounded | decimal.Inexact
	}
	return z.Set(approx)
}

var (
	approx1 = decimal.New(259, 3)
	approx2 = decimal.New(819, 3)
	approx3 = decimal.New(819, 4)
	approx4 = decimal.New(259, 2)
	ptFive  = decimal.New(5, 1)
)
