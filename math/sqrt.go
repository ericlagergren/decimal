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
	// TODO(eric): figure out what precision these intermediate variables need
	// and if we even need 'r' at all.
	var p0, q0, r decimal.Big
	p0.Context.Precision = decimal.UnlimitedPrecision
	q0.Context.Precision = decimal.UnlimitedPrecision
	r.Context.Precision = decimal.UnlimitedPrecision
	p0.Mul(p, p)
	q0.Mul(q, q)
	return Sqrt(z, r.Add(p0.Mul(p, p), q0.Mul(q, q)))
}

// Sqrt sets z to the square root of x and returns z.
func Sqrt(z, x *decimal.Big) *decimal.Big {
	if xs := x.Sign(); xs <= 0 {
		if xs == 0 {
			return z.SetMantScale(0, x.Scale()>>1)
		}
		// errors.New("math.Sqrt: cannot take square root of negative number"),
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}
	if z.CheckNaNs(x, nil) {
		return z
	}

	// Already checked for negative numbers.
	if x.IsInf(+1) {
		return z.SetInf(false)
	}

	prec := precision(z)

	// Fast path #1: use math.Sqrt if our decimal is small enough. 0 and 22
	// are implementation details of the Float64 method. If Float64 is altered,
	// change them.
	if prec <= 16 && (x.Scale() >= 0 && x.Scale() < 22) {
		return z.SetFloat64(math.Sqrt(x.Float64()))
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
		xprec = x.Precision()

		// The algorithm requires a normalized ``f ∈ [0.1, 1)'' Of the two ways
		// to normalize f, adjusting its scale is the quickest. However, it then
		// requires us to increment approx's scale by e/2 instead of simply
		// setting it to e/2.
		f = decimal.WithContext(x.Context).Copy(x).SetScale(xprec)
		e = -x.Scale() + xprec

		tmp    = decimal.WithContext(decimal.ContextUnlimited)
		approx = decimal.WithContext(z.Context)
	)

	if e&1 == 0 {
		approx.Add(approx1, tmp.Mul(approx2, f)) // approx := .259 + .819f
	} else {
		f.SetScale(f.Scale() + 1)                // f := f/10
		e++                                      // e := e + 1
		approx.Add(approx3, tmp.Mul(approx4, f)) // approx := 0.819 + 2.59*f
	}

	var (
		maxp     = prec + 2
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
	decimal.ToNearestEven.Round(approx, precision(z))
	return z.Set(approx)
}

var (
	approx1 = decimal.New(259, 3)
	approx2 = decimal.New(819, 3)
	approx3 = decimal.New(819, 4)
	approx4 = decimal.New(259, 2)
	ptFive  = decimal.New(5, 1)
)
