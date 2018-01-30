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

const (
	constPrec             = 100
	defaultExtraPrecision = 3
)

var (
	_E     = newDecimal("2.718281828459045235360287471352662497757247093699959574966967627724076630353547594571382178525166427")
	_Ln10  = newDecimal("2.302585092994045684017991454684364207601101488628772976033327900967572609677352480235997205089598298")
	_Pi    = newDecimal("3.141592653589793238462643383279502884197169399375105820974944592307816406286208998628034825342117068")
	_Pi2   = newDecimal("1.570796326794896619231321691639751442098584699687552910487472296153908203143104499314017412671058534")
	_Sqrt3 = newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248576")

	//_Gamma = newDecimal("0.577215664901532860606512090082402431042159335939923598805767234884867726777664670936947063291746749")
	//_Ln2   = newDecimal("0.693147180559945309417232121458176568075500134360255254120680009493393621969694715605863326996418687")
)

// E sets z to the mathematical constant e and returns z.
func E(z *decimal.Big) *decimal.Big {
	ctx := decimal.Context{Precision: precision(z)}
	if ctx.Precision <= constPrec {
		return ctx.Set(z, _E)
	}

	ctx.Precision += 5
	var (
		sum  = z.SetUint64(2)
		fac  = new(decimal.Big).SetUint64(1)
		term = new(decimal.Big)
		prev = new(decimal.Big)
	)

	for i := uint64(2); sum.Cmp(prev) != 0; i++ {
		// Use term as our intermediate storage for our factorial. SetUint64
		// should be marginally faster than ctx.Add(incr, incr, one), but either
		// the costly call to Quo makes it difficult to notice.
		term.SetUint64(i)
		ctx.Mul(fac, fac, term)
		ctx.Quo(term, one, fac)
		prev.Copy(sum)
		ctx.Add(sum, sum, term)
	}

	ctx.Precision -= 5
	return ctx.Set(z, sum)
}

// pi2 sets z to the mathematical constant pi / 2 and returns z.
func pi2(z *decimal.Big, ctx decimal.Context) *decimal.Big {
	if ctx.Precision <= constPrec {
		return ctx.Set(z, _Pi2)
	}
	return ctx.Quo(z, pi(z, ctx), two)
}

// Pi sets z to the mathematical constant pi and returns z.
func Pi(z *decimal.Big) *decimal.Big {
	return pi(z, decimal.Context{Precision: precision(z)})
}

// pi sets z to the mathematical constant pi and returns z.
func pi(z *decimal.Big, ctx decimal.Context) *decimal.Big {
	if ctx.Precision <= constPrec {
		return ctx.Set(z, _Pi)
	}

	var (
		lasts = new(decimal.Big)
		t     = new(decimal.Big).SetUint64(3)
		s     = z.SetUint64(3)
		n     = new(decimal.Big).SetUint64(1)
		na    = new(decimal.Big)
		d     = new(decimal.Big)
		da    = new(decimal.Big).SetUint64(24)
	)

	for s.Cmp(lasts) != 0 {
		ctx.Set(lasts, s)
		ctx.Add(n, n, na)
		ctx.Add(na, na, eight)
		ctx.Add(d, d, da)
		ctx.Add(da, da, thirtyTwo)
		ctx.Mul(t, t, n)
		ctx.Quo(t, t, d)
		ctx.Add(s, s, t)
	}
	return ctx.Round(z) // z == s
}

// ln10 sets z to log(10) and returns z.
func ln10(z *decimal.Big, prec int) *decimal.Big {
	ctx := decimal.Context{Precision: prec}
	if ctx.Precision <= constPrec {
		return ctx.Set(z, _Ln10)
	}

	// TODO(eric): we can (possibly?) speed this up by selecting a log10 constant
	// that's some truncation of our continued fraction and setting the starting
	// term to that position in our continued fraction.

	ctx.Precision += 3
	g := lgen{
		ctx: ctx,
		pow: eightyOne, // 9 * 9
		z2:  eleven,    // 9 + 2
		k:   -1,
		t:   Term{A: new(decimal.Big), B: new(decimal.Big)},
	}
	ctx.Quo(z, eighteen /* 9 * 2 */, Lentz(z, &g))
	ctx.Precision -= 3
	return ctx.Round(z)
}

// sqrt3 sets z to sqrt(3) and returns z.
func sqrt3(z *decimal.Big, ctx decimal.Context) *decimal.Big {
	if ctx.Precision <= constPrec {
		return ctx.Set(z, _Sqrt3)
	}
	// TODO(eric): get rid of this allocation.
	return ctx.Set(z, Sqrt(decimal.WithContext(ctx), three))
}
