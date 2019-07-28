package benchmarks

import (
	"math"

	"github.com/apmckinlay/gsuneido/util/dnum"
	"github.com/cockroachdb/apd"
	"github.com/ericlagergren/decimal"
	ssdec "github.com/shopspring/decimal"
	"gopkg.in/inf.v0"
)

func adjustPrecision(prec int) int {
	return int(math.Ceil(float64(prec) * 1.1))
}

var (
	infEight     = inf.NewDec(8, 0)
	infThirtyTwo = inf.NewDec(32, 0)
)

// PiInf calculates π to the desired precision using gopkg.in/inf.v0.
func PiInf(prec int) *inf.Dec {
	var (
		lasts = inf.NewDec(0, 0)
		t     = inf.NewDec(3, 0)
		s     = inf.NewDec(3, 0)
		n     = inf.NewDec(1, 0)
		na    = inf.NewDec(0, 0)
		d     = inf.NewDec(0, 0)
		da    = inf.NewDec(24, 0)

		work = adjustPrecision(prec)
	)

	for s.Cmp(lasts) != 0 {
		lasts.Set(s)
		n.Add(n, na)
		na.Add(na, infEight)
		d.Add(d, da)
		da.Add(da, infThirtyTwo)
		t.Mul(t, n)
		t.QuoRound(t, d, inf.Scale(work), inf.RoundHalfUp)
		s.Add(s, t)
	}
	// -1 because inf's precision == digits after radix
	return s.Round(s, inf.Scale(prec-1), inf.RoundHalfUp)
}

var (
	ssdecEight     = ssdec.New(8, 0)
	ssdecThirtyTwo = ssdec.New(32, 0)
)

// PiShopSpring calculates π to the desired precision using
// github.com/shopspring/decimal.
func PiShopSpring(prec int32) ssdec.Decimal {
	var (
		lasts = ssdec.New(0, 0)
		t     = ssdec.New(3, 0)
		s     = ssdec.New(3, 0)
		n     = ssdec.New(1, 0)
		na    = ssdec.New(0, 0)
		d     = ssdec.New(0, 0)
		da    = ssdec.New(24, 0)

		work = int32(adjustPrecision(int(prec)))
	)

	for s.Cmp(lasts) != 0 {
		lasts = s
		n = n.Add(na)
		na = na.Add(ssdecEight)
		d = d.Add(da)
		da = da.Add(ssdecThirtyTwo)
		t = t.Mul(n)
		t = t.DivRound(d, work)
		s = s.Add(t)
	}
	// -1 because shopSpring's prec == digits after radix
	return s.Round(prec - 1)
}

var (
	dnumEight     = dnum.New(+1, 8, 0)
	dnumThirtyTwo = dnum.New(+1, 32, 0)
)

// PiDnum calculates π to its maximum precision of 16 digits using
// github.com/apmckinlay/gsuneido/util/dnum.
func PiDnum() dnum.Dnum {
	var (
		lasts = dnum.New(+1, 0, 0)
		t     = dnum.New(+1, 3, 0)
		s     = dnum.New(+1, 3, 0)
		n     = dnum.New(+1, 1, 0)
		na    = dnum.New(+1, 0, 0)
		d     = dnum.New(+1, 0, 0)
		da    = dnum.New(+1, 24, 0)
	)

	for dnum.Compare(s, lasts) != 0 {
		lasts = s
		n = dnum.Add(n, na)
		na = dnum.Add(na, dnumEight)
		d = dnum.Add(d, da)
		da = dnum.Add(da, dnumThirtyTwo)
		t = dnum.Mul(t, n)
		t = dnum.Div(t, d)
		s = dnum.Add(s, t)
	}
	return s
}

// PiFloat calculates π to its maximum precision of 19 digits using Go's native
// float64.
func PiFloat64() float64 {
	var (
		lasts = 0.0
		t     = 3.0
		s     = 3.0
		n     = 1.0
		na    = 0.0
		d     = 0.0
		da    = 24.0
	)

	for s != lasts {
		lasts = s
		n += na
		na += 8
		d += da
		da += 32
		t = (t * n) / d
		s = t
	}
	return s
}

var (
	eight     = decimal.New(8, 0)
	thirtyTwo = decimal.New(32, 0)
)

// PiDecimal_Go calculates π to the desired precision using
// github.com/ericlagergren/decimal with the operating mode set to Go.
func PiDecimal_Go(prec int) *decimal.Big {
	var (
		ctx = decimal.Context{
			Precision:     adjustPrecision(prec),
			OperatingMode: decimal.Go,
		}

		lasts = new(decimal.Big)
		t     = decimal.New(3, 0)
		s     = decimal.New(3, 0)
		n     = decimal.New(1, 0)
		na    = new(decimal.Big)
		d     = new(decimal.Big)
		da    = decimal.New(24, 0)
		eps   = decimal.New(1, prec)
	)

	for {
		lasts.Copy(s)
		ctx.Add(n, n, na)
		ctx.Add(na, na, eight)
		ctx.Add(d, d, da)
		ctx.Add(da, da, thirtyTwo)
		ctx.Mul(t, t, n)
		ctx.Quo(t, t, d)
		ctx.Add(s, s, t)
		if ctx.Sub(lasts, s, lasts).CmpAbs(eps) < 0 {
			return s.Round(prec)
		}
	}
}

// PiDecimal_GDA calculates π to the desired precision using
// github.com/ericlagergren/decimal with the operating mode set to GDA.
func PiDecimal_GDA(prec int) *decimal.Big {
	var (
		ctx = decimal.Context{
			Precision:     adjustPrecision(prec),
			OperatingMode: decimal.GDA,
		}

		lasts = new(decimal.Big)
		t     = decimal.New(3, 0)
		s     = decimal.New(3, 0)
		n     = decimal.New(1, 0)
		na    = new(decimal.Big)
		d     = new(decimal.Big)
		da    = decimal.New(24, 0)
	)

	for s.Cmp(lasts) != 0 {
		lasts.Copy(s)
		ctx.Add(n, n, na)
		ctx.Add(na, na, eight)
		ctx.Add(d, d, da)
		ctx.Add(da, da, thirtyTwo)
		ctx.Mul(t, t, n)
		ctx.Quo(t, t, d)
		ctx.Add(s, s, t)
	}
	return s.Round(prec)
}

var (
	apdEight     = apd.New(8, 0)
	apdThirtyTwo = apd.New(32, 0)
)

// PiAPD calculates π to the desired precision using github.com/cockroachdb/apd.
func PiAPD(prec uint32) *apd.Decimal {
	var (
		ctx   = apd.BaseContext.WithPrecision(uint32(adjustPrecision(int(prec))))
		lasts = apd.New(0, 0)
		t     = apd.New(3, 0)
		s     = apd.New(3, 0)
		n     = apd.New(1, 0)
		na    = apd.New(0, 0)
		d     = apd.New(0, 0)
		da    = apd.New(24, 0)
	)

	for s.Cmp(lasts) != 0 {
		lasts.Set(s)
		ctx.Add(n, n, na)
		ctx.Add(na, na, apdEight)
		ctx.Add(d, d, da)
		ctx.Add(da, da, apdThirtyTwo)
		ctx.Mul(t, t, n)
		ctx.Quo(t, t, d)
		ctx.Add(s, s, t)
	}
	ctx.Precision = prec
	ctx.Round(s, s)
	return s
}
