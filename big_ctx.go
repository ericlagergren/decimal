package decimal

import (
	"math"
	"math/big"

	"github.com/ericlagergren/decimal/internal/arith"
	"github.com/ericlagergren/decimal/internal/arith/checked"
	cst "github.com/ericlagergren/decimal/internal/c"
	"github.com/ericlagergren/decimal/internal/compat"
)

// Add sets z to x + y and returns z.
func (c Context) Add(z *Big, x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}
	if z.validateContext(c) {
		return z
	}

	if x.IsFinite() && y.IsFinite() {
		neg := c.add(z, x, x.Signbit(), y, y.Signbit())
		z.form = finite
		if neg {
			z.form |= signbit
		}
		return c.round(z.norm())
	}

	// NaN + NaN
	// NaN + y
	// x + NaN
	if z.checkNaNs(x, y, addition) {
		return z
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 && x.form^y.form == signbit {
			// +Inf + -Inf
			// -Inf + +Inf
			return z.setNaN(InvalidOperation, qnan, addinfinf)
		}
		// ±Inf + y
		// +Inf + +Inf
		// -Inf + -Inf
		return z.Set(x)
	}
	// x + ±Inf
	return z.Set(y)
}

func (c Context) add(z *Big, x *Big, xn bool, y *Big, yn bool) (neg bool) {
	hi, lo := x, y
	hineg, loneg := xn, yn
	if hi.exp < lo.exp {
		hi, lo = lo, hi
		hineg, loneg = loneg, hineg
	}

	if neg, ok := c.tryTinyAdd(z, hi, hineg, lo, loneg); ok {
		return neg
	}

	if hi.isCompact() {
		if lo.isCompact() {
			neg = c.addCompact(z, hi.compact, hineg, lo.compact, loneg, uint64(hi.exp-lo.exp))
		} else {
			neg = c.addMixed(z, &lo.unscaled, loneg, lo.exp, hi.compact, hineg, hi.exp)
		}
	} else if lo.isCompact() {
		neg = c.addMixed(z, &hi.unscaled, hineg, hi.exp, lo.compact, loneg, lo.exp)
	} else {
		neg = c.addBig(z, &hi.unscaled, hineg, &lo.unscaled, loneg, uint64(hi.exp-lo.exp))
	}
	z.exp = lo.exp
	return neg
}

// tryTinyAdd returns true if hi + lo requires a huge shift that will produce
// the same results as a smaller shift. E.g., 3 + 0e+9999999999999999 with a
// precision of 5 doesn't need to be shifted by a large number.
func (c Context) tryTinyAdd(z *Big, hi *Big, hineg bool, lo *Big, loneg bool) (neg, ok bool) {
	if hi.compact == 0 {
		return false, false
	}

	exp := hi.exp - 1
	if hp, zp := hi.precision, precision(c); hp <= zp {
		exp += hp - zp - 1
	}

	if lo.adjusted() >= exp {
		return false, false
	}

	var tiny uint64
	if lo.compact != 0 {
		tiny = 1
	}
	tinyneg := loneg

	if hi.isCompact() {
		shift := uint64(hi.exp - exp)
		neg = c.addCompact(z, hi.compact, hineg, tiny, tinyneg, shift)
	} else {
		neg = c.addMixed(z, &hi.unscaled, hineg, hi.exp, tiny, tinyneg, exp)
	}
	z.exp = exp
	return neg, true
}

func (c Context) addCompact(z *Big, hi uint64, hineg bool, lo uint64, loneg bool, shift uint64) bool {
	neg := hineg
	if hi, ok := checked.MulPow10(hi, shift); ok {
		// Try regular addition and fall back to 128-bit addition.
		if loneg == hineg {
			if z.compact, ok = checked.Add(hi, lo); !ok {
				arith.Add128(&z.unscaled, hi, lo)
				z.compact = cst.Inflated
			}
		} else {
			if z.compact, ok = checked.Sub(hi, lo); !ok {
				neg = !neg
				arith.Sub128(&z.unscaled, lo, hi)
				z.compact = cst.Inflated
			}
		}
		// "Otherwise, the sign of a zero result is 0 unless either both
		// operands were negative or the signs of the operands were different
		// and the rounding is round-floor."
		return (z.compact == 0 && c.RoundingMode == ToNegativeInf && neg) || neg
	}

	{
		hi := z.unscaled.SetUint64(hi)
		hi = checked.MulBigPow10(hi, hi, shift)
		if hineg == loneg {
			arith.Add(&z.unscaled, hi, lo)
		} else {
			// lo had to be promoted to a big.Int, so by definition it'll be
			// larger than hi. Therefore, we do not need to negate neg, nor do
			// we need to check to see if the result == 0.
			arith.Sub(&z.unscaled, hi, lo)
		}
	}
	z.compact = cst.Inflated
	return neg
}

func (c Context) addMixed(z *Big, x *big.Int, xneg bool, xs int, y uint64, yn bool, ys int) bool {
	switch {
	case xs < ys:
		shift := uint64(ys - xs)
		if y0, ok := checked.MulPow10(y, shift); ok {
			y = y0
			break
		}

		// See comment in addCompact.
		yb := alias(&z.unscaled, x).SetUint64(y)
		yb = checked.MulBigPow10(yb, yb, shift)

		neg := xneg
		if xneg == yn {
			z.unscaled.Add(x, yb)
		} else {
			if x.Cmp(yb) >= 0 {
				z.unscaled.Sub(x, yb)
			} else {
				neg = !neg
				z.unscaled.Sub(yb, x)
			}
		}
		if z.unscaled.Sign() == 0 {
			z.compact = 0
		} else {
			z.compact = cst.Inflated
		}
		return (z.compact == 0 && c.RoundingMode == ToNegativeInf && neg) || neg
	case xs > ys:
		x = checked.MulBigPow10(&z.unscaled, x, uint64(xs-ys))
	}

	if xneg == yn {
		arith.Add(&z.unscaled, x, y)
	} else {
		// x > y
		arith.Sub(&z.unscaled, x, y)
	}

	z.compact = cst.Inflated
	return xneg
}

func (c Context) addBig(z *Big, hi *big.Int, hineg bool, lo *big.Int, loneg bool, shift uint64) bool {
	if shift != 0 {
		hi = checked.MulBigPow10(alias(&z.unscaled, lo), hi, shift)
	}
	neg := hineg
	if hineg == loneg {
		z.unscaled.Add(hi, lo)
	} else {
		if hi.Cmp(lo) >= 0 {
			z.unscaled.Sub(hi, lo)
		} else {
			neg = !neg
			z.unscaled.Sub(lo, hi)
		}
	}
	if z.unscaled.Sign() == 0 {
		z.compact = 0
	} else {
		z.compact = cst.Inflated
	}
	return z.compact != 0 && neg
}

// FMA sets z to (x * y) + u without any intermediate rounding.
func (c Context) FMA(z, x, y, u *Big) *Big {
	if z.validateContext(c) {
		return z
	}
	// Create a temporary reciever in the case z == u so we handle the case
	// z.FMA(x, y, z) without clobbering z partway through.
	z0 := z
	if z == u {
		z0 = WithContext(c)
	}
	c.mul(z0, x, y, true)
	if z0.Context.Conditions&InvalidOperation != 0 {
		return z.setShared(z0)
	}
	return z.setShared(c.Add(z0, z0, u))
}

// Mul sets z to x * y and returns z.
func (c Context) Mul(z *Big, x, y *Big) *Big {
	if z.validateContext(c) {
		return z
	}
	return c.fix(c.mul(z, x, y, false))
}

// mul is the implementation of Mul, but with a boolean to toggle rounding. This
// is useful for FMA.
func (c Context) mul(z *Big, x, y *Big, fma bool) *Big {
	if debug {
		x.validate()
		y.validate()
	}

	sign := x.form&signbit ^ y.form&signbit

	if x.IsFinite() && y.IsFinite() {
		// Multiplication is simple, so inline it.
		if x.isCompact() {
			if y.isCompact() {
				if prod, ok := checked.Mul(x.compact, y.compact); ok {
					z.compact = prod
				} else {
					// Overflow: use 128 bit multiplication.
					arith.Mul128(&z.unscaled, x.compact, y.compact)
					z.compact = cst.Inflated
				}
			} else { // y.isInflated
				arith.MulUint64(&z.unscaled, &y.unscaled, x.compact)
				z.compact = cst.Inflated
			}
		} else if y.isCompact() { // x.isInflated
			arith.MulUint64(&z.unscaled, &x.unscaled, y.compact)
			z.compact = cst.Inflated
		} else {
			z.unscaled.Mul(&x.unscaled, &y.unscaled)
			z.compact = cst.Inflated
		}

		z.form = finite | sign
		z.exp = x.exp + y.exp
		z.norm()
		if !fma {
			c.round(z)
		}
		return z
	}

	// NaN * NaN
	// NaN * y
	// x * NaN
	if z.checkNaNs(x, y, multiplication) {
		return z
	}

	if (x.IsInf(0) && y.compact != 0) ||
		(y.IsInf(0) && x.compact != 0) ||
		(y.IsInf(0) && x.IsInf(0)) {
		// ±Inf * y
		// x * ±Inf
		// ±Inf * ±Inf
		return z.SetInf(sign != 0)
	}

	// 0 * ±Inf
	// ±Inf * 0
	return z.setNaN(InvalidOperation, qnan, mul0inf)
}

// Quantize sets z to the number equal in value and sign to z with the scale, n.
func (c Context) Quantize(z *Big, n int) *Big {
	if debug {
		z.validate()
	}
	if z.validateContext(c) {
		return z
	}

	n = -n
	if z.isSpecial() {
		if z.form&inf != 0 {
			return z.setNaN(InvalidOperation, qnan, quantinf)
		}
		z.checkNaNs(z, z, quantization)
		return z
	}

	if n > MaxScale || n < c.etiny() {
		return z.setNaN(InvalidOperation, qnan, quantminmax)
	}

	shift := z.exp - n
	if z.precision+shift > precision(c) {
		return z.setNaN(InvalidOperation, qnan, quantprec)
	}

	z.exp = n
	if shift == 0 || z.compact == 0 {
		return z
	}

	if shift < 0 {
		z.Context.Conditions |= Rounded
	}

	m := c.RoundingMode
	neg := z.form & signbit
	if z.isCompact() {
		if shift > 0 {
			if zc, ok := checked.MulPow10(z.compact, uint64(shift)); ok {
				return z.setTriple(zc, neg, n)
			}
			// shift < 0
		} else if yc, ok := arith.Pow10(uint64(-shift)); ok {
			return z.quo(m, z.compact, neg, yc, 0)
		}
		z.unscaled.SetUint64(z.compact)
		z.compact = cst.Inflated
	}

	if shift > 0 {
		_ = checked.MulBigPow10(&z.unscaled, &z.unscaled, uint64(shift))
		z.precision = arith.BigLength(&z.unscaled)
		return z
	}
	return z.quoBig(m, &z.unscaled, neg, arith.BigPow10(uint64(-shift)), 0)
}

// Quo sets z to x / y and returns z.
func (c Context) Quo(z *Big, x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}
	if z.validateContext(c) {
		return z
	}

	sign := (x.form & signbit) ^ (y.form & signbit)
	if x.isSpecial() || y.isSpecial() {
		// NaN / NaN
		// NaN / y
		// x / NaN
		if z.checkNaNs(x, y, division) {
			return z
		}

		if x.form&inf != 0 {
			if y.form&inf != 0 {
				// ±Inf / ±Inf
				z.setNaN(InvalidOperation, qnan, quoinfinf)
				return z
			}
			// ±Inf / y
			return z.SetInf(sign != 0)
		}
		// x / ±Inf
		z.Context.Conditions |= Clamped
		return z.setZero(sign, c.etiny())
	}

	if y.compact == 0 {
		if x.compact == 0 {
			// 0 / 0
			return z.setNaN(InvalidOperation|DivisionUndefined, qnan, quo00)
		}
		// x / 0
		z.Context.Conditions |= DivisionByZero
		return z.SetInf(sign != 0)
	}
	if x.compact == 0 {
		// 0 / y
		return c.fix(z.setZero(sign, x.exp-y.exp))
	}

	m := c.RoundingMode
	yp := y.precision
	zp := precision(c)
	if zp == UnlimitedPrecision {
		m = unnecessary
		zp = x.precision + int(math.Ceil(10*float64(yp)/3))
	}

	if x.isCompact() && y.isCompact() {
		if cmpNorm(x.compact, x.precision, y.compact, yp) {
			yp--
		}

		shift := zp + yp - x.precision
		z.exp = (x.exp - y.exp) - shift
		if shift > 0 {
			if sx, ok := checked.MulPow10(x.compact, uint64(shift)); ok {
				return z.quo(m, sx, x.form, y.compact, y.form)
			}
			xb := z.unscaled.SetUint64(x.compact)
			xb = checked.MulBigPow10(xb, xb, uint64(shift))
			yb := new(big.Int).SetUint64(y.compact)
			return z.quoBig(m, xb, x.form, yb, y.form)
		}
		if shift < 0 {
			if sy, ok := checked.MulPow10(y.compact, uint64(-shift)); ok {
				return z.quo(m, x.compact, x.form, sy, y.form)
			}
			yb := z.unscaled.SetUint64(y.compact)
			yb = checked.MulBigPow10(yb, yb, uint64(-shift))
			xb := new(big.Int).SetUint64(x.compact)
			return z.quoBig(m, xb, x.form, yb, y.form)
		}
		return z.quo(m, x.compact, x.form, y.compact, y.form)
	}

	xb, yb := &x.unscaled, &y.unscaled
	if x.isCompact() {
		xb = new(big.Int).SetUint64(x.compact)
	} else if y.isCompact() {
		yb = new(big.Int).SetUint64(y.compact)
	}

	if cmpNormBig(&z.unscaled, xb, x.precision, yb, yp) {
		yp--
	}

	shift := zp + yp - x.precision
	z.exp = (x.exp - y.exp) - shift
	if shift > 0 {
		tmp := alias(&z.unscaled, yb)
		xb = checked.MulBigPow10(tmp, xb, uint64(shift))
	} else if shift < 0 {
		tmp := alias(&z.unscaled, xb)
		yb = checked.MulBigPow10(tmp, yb, uint64(-shift))
	}
	return z.quoBig(m, xb, x.form, yb, y.form)
}

func (z *Big) quo(m RoundingMode, x uint64, xneg form, y uint64, yneg form) *Big {
	z.form = xneg ^ yneg
	z.compact = x / y
	r := x % y
	if r == 0 {
		z.precision = arith.Length(z.compact)
		return z
	}

	z.Context.Conditions |= Inexact | Rounded
	if m == ToZero {
		z.precision = arith.Length(z.compact)
		return z
	}

	rc := 1
	if r2, ok := checked.Mul(r, 2); ok {
		rc = arith.Cmp(r2, y)
	}

	if m == unnecessary {
		return z.setNaN(
			InvalidOperation|InvalidContext|InsufficientStorage, qnan, quotermexp)
	}
	if m.needsInc(z.compact&1 != 0, rc, xneg == yneg) {
		z.Context.Conditions |= Rounded
		z.compact++
	}
	z.precision = arith.Length(z.compact)
	return z
}

func (z *Big) quoBig(m RoundingMode, x *big.Int, xneg form, y *big.Int, yneg form) *Big {
	z.compact = cst.Inflated
	z.form = xneg ^ yneg

	q, r := z.unscaled.QuoRem(x, y, new(big.Int))
	if r.Sign() == 0 {
		return z.norm()
	}

	z.Context.Conditions |= Inexact | Rounded
	if m == ToZero {
		return z.norm()
	}

	var rc int
	rv := r.Uint64()
	// Drop into integers if possible.
	if arith.IsUint64(r) && arith.IsUint64(y) && rv <= math.MaxUint64/2 {
		rc = arith.Cmp(rv*2, y.Uint64())
	} else {
		rc = compat.BigCmpAbs(new(big.Int).Mul(r, cst.TwoInt), y)
	}

	if m == unnecessary {
		return z.setNaN(
			InvalidOperation|InvalidContext|InsufficientStorage, qnan, quotermexp)
	}
	if m.needsInc(q.Bit(0) != 0, rc, xneg == yneg) {
		z.Context.Conditions |= Rounded
		arith.Add(q, q, 1)
	}
	return z.norm()
}

// QuoInt sets z to x / y with the remainder truncated. See QuoRem for more
// details.
func (c Context) QuoInt(z *Big, x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}
	if z.validateContext(c) {
		return z
	}

	sign := (x.form & signbit) ^ (y.form & signbit)
	if x.IsFinite() && y.IsFinite() {
		if y.compact == 0 {
			if x.compact == 0 {
				// 0 / 0
				return z.setNaN(InvalidOperation|DivisionUndefined, qnan, quo00)
			}
			// x / 0
			z.Context.Conditions |= DivisionByZero
			return z.SetInf(sign != 0)
		}
		if x.compact == 0 {
			// 0 / y
			return c.fix(z.setZero(sign, 0))
		}
		z, _ = c.quorem(z, nil, x, y)
		z.exp = 0
		return z
	}

	// NaN / NaN
	// NaN / y
	// x / NaN
	if z.checkNaNs(x, y, division) {
		return z
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 {
			// ±Inf / ±Inf
			return z.setNaN(InvalidOperation, qnan, quoinfinf)
		}
		// ±Inf / y
		return z.SetInf(sign != 0)
	}
	// x / ±Inf
	return z.setZero(sign, 0)
}

// QuoRem sets z to the quotient x / y and r to the remainder x % y, such that
// x = z * y + r, and returns the pair (z, r).
func (c Context) QuoRem(z *Big, x, y, r *Big) (*Big, *Big) {
	if debug {
		x.validate()
		y.validate()
	}
	if z.validateContext(c) {
		r.validateContext(c)
		return z, r
	}

	sign := (x.form & signbit) ^ (y.form & signbit)
	if x.IsFinite() && y.IsFinite() {
		if y.compact == 0 {
			if x.compact == 0 {
				// 0 / 0
				z.setNaN(InvalidOperation|DivisionUndefined, qnan, quo00)
				r.setNaN(InvalidOperation|DivisionUndefined, qnan, quo00)
			}
			// x / 0
			z.Context.Conditions |= DivisionByZero
			r.Context.Conditions |= DivisionByZero
			return z.SetInf(sign != 0), r.SetInf(x.Signbit())
		}
		if x.compact == 0 {
			// 0 / y
			z.setZero((x.form^y.form)&signbit, 0)
			r.setZero(x.form, y.exp-x.exp)
			return c.fix(z), c.fix(r)
		}
		return c.quorem(z, r, x, y)
	}

	// NaN / NaN
	// NaN / y
	// x / NaN
	if z.checkNaNs(x, y, division) {
		return z, r.Set(z)
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 {
			// ±Inf / ±Inf
			z.setNaN(InvalidOperation, qnan, quoinfinf)
			return z, r.Set(z)
		}
		// ±Inf / y
		return z.SetInf(sign != 0), r.SetInf(x.form&signbit != 0)
	}
	// x / ±Inf
	z.Context.Conditions |= Clamped
	z.setZero(sign, c.etiny())
	r.setZero(x.form&signbit /* ??? */, 0)
	return z, r
}

func (c Context) quorem(z0, z1, x, y *Big) (*Big, *Big) {
	m := c.RoundingMode
	zp := precision(c)

	if x.adjusted()-y.adjusted() > zp {
		if z0 != nil {
			z0.setNaN(DivisionImpossible, qnan, quorem_)
		}
		if z1 != nil {
			z1.setNaN(DivisionImpossible, qnan, quorem_)
		}
		return z0, z1
	}

	z := z0
	if z == nil {
		z = z1
	}

	if x.isCompact() && y.isCompact() {
		shift := x.exp - y.exp
		if shift > 0 {
			if sx, ok := checked.MulPow10(x.compact, uint64(shift)); ok {
				return m.quorem(z0, z1, sx, x.form, y.compact, y.form)
			}
			xb := z.unscaled.SetUint64(x.compact)
			xb = checked.MulBigPow10(xb, xb, uint64(shift))
			yb := new(big.Int).SetUint64(y.compact)
			return m.quoremBig(z0, z1, xb, x.form, yb, y.form)
		}
		if shift < 0 {
			if sy, ok := checked.MulPow10(y.compact, uint64(-shift)); ok {
				return m.quorem(z0, z1, x.compact, x.form, sy, y.form)
			}
			yb := z.unscaled.SetUint64(y.compact)
			yb = checked.MulBigPow10(yb, yb, uint64(-shift))
			xb := new(big.Int).SetUint64(x.compact)
			return m.quoremBig(z0, z1, xb, x.form, yb, y.form)
		}
		return m.quorem(z0, z1, x.compact, x.form, y.compact, y.form)
	}

	xb, yb := &x.unscaled, &y.unscaled
	if x.isCompact() {
		xb = new(big.Int).SetUint64(x.compact)
	} else if y.isCompact() {
		yb = new(big.Int).SetUint64(y.compact)
	}

	shift := x.exp - y.exp
	if shift > 0 {
		tmp := alias(&z.unscaled, yb)
		xb = checked.MulBigPow10(tmp, xb, uint64(shift))
	} else {
		tmp := alias(&z.unscaled, xb)
		yb = checked.MulBigPow10(tmp, yb, uint64(-shift))
	}
	return m.quoremBig(z0, z1, xb, x.form, yb, y.form)
}

// TODO(eric): quorem and quoremBig should not be methods on RoundingMode

func (m RoundingMode) quorem(
	z0, z1 *Big,
	x uint64, xneg form, y uint64, yneg form,
) (*Big, *Big) {
	if z0 != nil {
		z0.setTriple(x/y, xneg^yneg, 0)
	}
	if z1 != nil {
		z1.setTriple(x%y, xneg, 0)
	}
	return z0, z1
}

func (m RoundingMode) quoremBig(
	z0, z1 *Big,
	x *big.Int, xneg form,
	y *big.Int, yneg form,
) (*Big, *Big) {
	if z0 == nil {
		r := z1.unscaled.Rem(x, y)
		z1.compact = cst.Inflated
		z1.form = xneg
		z1.precision = arith.BigLength(r)
		return z0, z1.norm()
	}

	var q, r *big.Int
	if z1 != nil {
		q, r = z0.unscaled.QuoRem(x, y, &z1.unscaled)
		z1.compact = cst.Inflated
		z1.form = xneg
		z1.precision = arith.BigLength(r)
		z1.norm()
	} else {
		q, _ = z0.unscaled.QuoRem(x, y, new(big.Int))
	}
	if z0 != nil {
		z0.compact = cst.Inflated
		z0.form = xneg ^ yneg
		z0.precision = arith.BigLength(q)
	}
	return z0.norm(), z1
}

// Rem sets z to the remainder x % y. See QuoRem for more details.
func (c Context) Rem(z *Big, x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}
	if z.validateContext(c) {
		return z
	}

	if x.IsFinite() && y.IsFinite() {
		if y.compact == 0 {
			if x.compact == 0 {
				// 0 / 0
				return z.setNaN(InvalidOperation|DivisionUndefined, qnan, quo00)
			}
			// x / 0
			return z.setNaN(InvalidOperation|DivisionByZero, qnan, remx0)
		}
		if x.compact == 0 {
			// 0 / y
			return z.setZero(x.form&signbit, min(x.exp, y.exp))
		}
		_, z = c.quorem(nil, z, x, y)
		z.exp = min(x.exp, y.exp)
		return c.round(z)
	}

	// NaN / NaN
	// NaN / y
	// x / NaN
	if z.checkNaNs(x, y, division) {
		return z
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 {
			// ±Inf / ±Inf
			return z.setNaN(InvalidOperation, qnan, quoinfinf)
		}
		// ±Inf / y
		return z.setNaN(InvalidOperation, qnan, reminfy)
	}
	// x / ±Inf
	return z.Set(x)
}

// Round rounds z down to the Context's precision and returns z. The result is
// undefined if z is not finite. The result of Round will always be within the
// interval [⌊10**x⌋, z] where x = the precision of z.
func (c Context) Round(z *Big) *Big {
	if debug {
		z.validate()
	}
	if z.validateContext(c) {
		return z
	}

	n := precision(c)
	if n == UnlimitedPrecision || z.isSpecial() {
		return z
	}

	if z.precision <= n {
		return c.fix(z)
	}

	shift := z.precision - n
	if shift > MaxScale {
		return z.xflow(false, true)
	}
	z.exp += shift

	z.Context.Conditions |= Rounded
	c.shiftr(z, uint64(shift))

	// Test to see if we accidentally increased precision because of rounding.
	// For example, given n = 17 and RoundingMode = ToNearestEven, rounding
	//
	//   0.9999999999999999994284
	//
	// results in
	//
	//   0.99999999999999999 (precision = 17)
	//
	// which is rounded up to
	//
	//   1.00000000000000000 (precision = 18)
	//
	// Ideally, we'd have a more efficient method of correcting the extra
	// precision than this double shift. Inside the quo and quoBig methods we
	// might try something like
	//
	//   if m.needsInc( ... ) {
	//      z.precision = arith.Length(z.compact)
	//      z.compact++
	//      if arith.Length(z.compact) != z.precision {
	//          z.compact /= 10
	//      }
	//   }
	//
	// But that might have too much overhead for the general case division.
	if z.precision != n {
		c.shiftr(z, 1)
		z.exp++
	}
	return c.fix(z)
}

func (c Context) shiftr(z *Big, n uint64) bool {
	// TODO(eric): return value from this function.

	if n >= uint64(z.precision) {
		z.compact = 0
		z.precision = 1
		return true
	}

	if z.compact == 0 {
		return true
	}

	m := c.RoundingMode
	if z.isCompact() {
		if y, ok := arith.Pow10(n); ok {
			z.quo(m, z.compact, z.form, y, 0)
			return true
		}
		z.unscaled.SetUint64(z.compact)
		z.compact = cst.Inflated
	}
	z.quoBig(m, &z.unscaled, z.form, arith.BigPow10(n), 0)
	return true
}

func (c Context) round(z *Big) *Big {
	if c.OperatingMode == GDA {
		return c.Round(z)
	}
	return z
}

// Set sets z to x and returns z. The result might be rounded, even if z == x.
func (c Context) Set(z, x *Big) *Big {
	return c.Round(z.Copy(x))
}

// Sub sets z to x - y and returns z.
func (c Context) Sub(z *Big, x, y *Big) *Big {
	if debug {
		x.validate()
		y.validate()
	}
	if z.validateContext(c) {
		return z
	}

	if x.IsFinite() && y.IsFinite() {
		neg := c.add(z, x, x.Signbit(), y, !y.Signbit())
		z.form = finite
		if neg {
			z.form |= signbit
		}
		return c.round(z.norm())
	}

	// NaN - NaN
	// NaN - y
	// x - NaN
	if z.checkNaNs(x, y, subtraction) {
		return z
	}

	if x.form&inf != 0 {
		if y.form&inf != 0 && (x.form&signbit == y.form&signbit) {
			// -Inf - -Inf
			// -Inf - -Inf
			return z.setNaN(InvalidOperation, qnan, subinfinf)
		}
		// ±Inf - y
		// -Inf - +Inf
		// +Inf - -Inf
		return z.Set(x)
	}
	// x - ±Inf
	return z.Neg(y)
}
