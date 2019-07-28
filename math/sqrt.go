package math

import "github.com/ericlagergren/decimal"

// Hypot sets z to Sqrt(p*p + q*q) and returns z.
func Hypot(z, p, q *decimal.Big) *decimal.Big {
	if z.CheckNaNs(p, q) {
		return z
	}

	ctx := decimal.Context{Precision: precision(z) + 1}

	var p0 decimal.Big
	ctx.Mul(&p0, p, p)

	if p == q {
		return Sqrt(z, ctx.Add(z, &p0, &p0))
	}

	var q0 decimal.Big
	ctx.Mul(&q0, q, q)
	return Sqrt(z, ctx.Add(z, &p0, &q0))
}

var (
	approx1 = decimal.New(259, 3)
	approx2 = decimal.New(819, 3)
	approx3 = decimal.New(819, 4)
	approx4 = decimal.New(259, 2)
	ptFive  = decimal.New(5, 1)
)

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
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	// Already checked for negative numbers.
	if x.IsInf(+1) {
		return z.SetInf(false)
	}

	var (
		prec = precision(z)
		ctx  = decimal.Context{Precision: prec}
		rnd  = z.Context.Conditions&decimal.Rounded != 0
		ixt  = z.Context.Conditions&decimal.Inexact != 0
	)

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
		f = new(decimal.Big).Copy(x).SetScale(xprec)
		e = -x.Scale() + xprec

		tmp decimal.Big // scratch space
	)

	if e&1 == 0 {
		ctx.FMA(z, approx2, f, approx1) // approx := .259 + .819f
	} else {
		f.SetScale(f.Scale() + 1)       // f := f/10
		e++                             // e := e + 1
		ctx.FMA(z, approx4, f, approx3) // approx := .0819 + 2.59f
	}

	maxp := prec + 5 // extra prec to skip weird +/- 0.5 adjustments
	ctx.Precision = 3
	for {
		// p := min(2*p - 2, maxp)
		ctx.Precision = min(2*ctx.Precision-2, maxp)

		// approx := .5*(approx + f/approx)
		ctx.Mul(z, ptFive, ctx.Add(&tmp, z, ctx.Quo(&tmp, f, z)))
		if ctx.Precision == maxp {
			break
		}
	}

	// The paper also specifies an additional code block for adjusting approx.
	// This code never went into the branches that modified approx, and rounding
	// to half even does the same thing. The GDA spec requires us to use
	// rounding mode half even (speleotrove.com/decimal/daops.html#refsqrt)
	// anyway.

	ctx.Reduce(z.SetScale(z.Scale() - e/2))
	if z.Precision() <= prec {
		if !rnd {
			z.Context.Conditions &= ^decimal.Rounded
		}
		if !ixt {
			z.Context.Conditions &= ^decimal.Inexact
		}
	}
	ctx.Precision = prec
	return ctx.Round(z)
}
