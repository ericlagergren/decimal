package math

import (
	"fmt"

	"github.com/ericlagergren/decimal"
)

func newDecimal(s string) *decimal.Big {
	x, ok := new(decimal.Big).SetString(s)
	if !ok {
		panic(fmt.Sprintf("bad input: %q", s))
	}
	return x
}

const constPrec = 100

var (
	_E      = newDecimal("2.718281828459045235360287471352662497757247093699959574966967627724076630353547594571382178525166427")
	_Pi     = newDecimal("3.141592653589793238462643383279502884197169399375105820974944592307816406286208998628034825342117067")
	_Ln10   = newDecimal("2.302585092994045684017991454684364207601101488628772976033327900967572609677352480235997205089598298")
	_Log10e = newDecimal("0.4342944819032518276511289189166050822943970058036665661144537831658646492088707747292249493384317483")
	//_Gamma = newDecimal("0.577215664901532860606512090082402431042159335939923598805767234884867726777664670936947063291746749")
	//_Ln2   = newDecimal("0.693147180559945309417232121458176568075500134360255254120680009493393621969694715605863326996418687")
)

// E sets z to the mathematical constant e.
func E(z *decimal.Big) *decimal.Big {
	prec := precision(z)
	if prec <= constPrec {
		return z.Set(_E)
	}

	var (
		fac  = decimal.WithContext(z.Context).SetMantScale(1, 0)
		incr = decimal.WithContext(z.Context).SetMantScale(1, 0)
		sum  = decimal.WithContext(z.Context).SetMantScale(2, 0)
		term = decimal.WithContext(z.Context).SetMantScale(0, 0)
		prev = decimal.WithContext(z.Context).SetMantScale(0, 0)
	)

	for sum.Round(prec).Cmp(prev) != 0 {
		fac.Mul(fac, incr.Add(incr, one))
		prev.Copy(sum)
		sum.Add(sum, term.Quo(one, fac))
	}
	return sum
}

// Pi sets z to the mathematical constant π.
func Pi(z *decimal.Big) *decimal.Big {
	prec := precision(z)
	if prec <= constPrec {
		return z.Set(_Pi)
	}

	var (
		lasts = decimal.WithContext(z.Context).SetMantScale(0, 0)
		t     = decimal.WithContext(z.Context).SetMantScale(3, 0)
		s     = decimal.WithContext(z.Context).SetMantScale(3, 0)
		n     = decimal.WithContext(z.Context).SetMantScale(1, 0)
		na    = decimal.WithContext(z.Context).SetMantScale(0, 0)
		d     = decimal.WithContext(z.Context).SetMantScale(0, 0)
		da    = decimal.WithContext(z.Context).SetMantScale(24, 0)
	)

	for s.Round(prec).Cmp(lasts) != 0 {
		lasts.Set(s)
		n.Add(n, na)
		na.Add(na, eight)
		d.Add(d, da)
		da.Add(da, thirtyTwo)
		t.Mul(t, n)
		t.Quo(t, d)
		s.Add(s, t)
	}
	return s
}

// ln10 sets z to log(10) and returns z.
func ln10(z *decimal.Big, prec int) *decimal.Big {
	if prec <= constPrec {
		return z.Set(_Ln10)
	}

	// TODO(eric): we can speed this up by selecting a log10 constant that's some
	// truncation of our continued fraction and setting the starting term to
	// that position in our continued fraction.

	prec += 3
	g := lgen{
		prec: prec,
		pow:  eightyOne, // 9 * 9
		z2:   eleven,    // 9 + 2
		k:    -1,
		t:    Term{A: decimal.WithPrecision(prec), B: decimal.WithPrecision(prec)},
	}
	return z.Quo(eighteen /* 9 * 2 */, Lentz(z, &g))
}

// log10e sets z to log10(e).
func log10e(z *decimal.Big) *decimal.Big {
	if prec := precision(z); prec < constPrec {
		return z.Set(_Log10e)
	}
	return Log10(z, E(z))
}

/*
// Gamma sets z to the mathematical constant γ,
func Gamma(z *decimal.Big) *decimal.Big {
	prec := z.Context.Precision()
	if prec <= 100 {
		return z.Set(_Gamma)
	}

	// Antonino Machado Souza Filho and Georges Schwachheim. 1967.
	// Algorithm 309: Gamma function with arbitrary precision.
	// Commun. ACM 10, 8 (August 1967), 511-512.
	// DOI=http://dx.doi.org/10.1145/363534.363561

}

func loggamma(z, t *decimal.Big) *decimal.Big {
	var tmin *decimal.Big

	zcp := z.Context.Precision()
	if zcp >= 18 {
		tmin = decimal.New(int64(zcp), 0)
	} else {
		tmin = decimal.New(7, 0)
	}

	if t.Cmp(tmin) {
		return lgm(z, t)
	}

	f := new(decimal.Big).Copy(t)
	t0 := new(decimal.Big).Copy(t)

	for {
		t0.Add(t0, one)
		if t0.Comp(tmin) >= 0 {
			break
		}
		f.Mul(f, t0)
	}

	lgm(z, t0)

	tmp := z.Context.New(0, 0)
	Ln(tmp, f)

	return z.Sub(z, ln(tmp, f))
}

func lgm(z, w *decimal.Big) *decimal.Big {
	var c [20]*decimal.Big

	w0 := new(decimal.Big).Copy(w)
	den := new(decimal.Big).Copy(w) // den := w
	w2 := new(decimal.Big).Copy(w)  // w2 := w

	tmp := z.Context.New(0, 0)

	presum := new(decimal.Big)
	// presum := (w - .5) * ln(w) - w + const
	presum.Sub(w, ptFive)
	presum.Mul(presum, Ln(&tmp, w))
	presum.Sub(presum, tmp.Add(w, cnst))
}
*/
