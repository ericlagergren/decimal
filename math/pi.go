package math

import (
	"github.com/ericlagergren/decimal"
	"math"
)

var (
	_10005             = decimal.New(10005, 0)
	_426880            = decimal.New(426880, 0)
	_13591409          = decimal.New(13591409, 0)
	_545140134         = decimal.New(545140134, 0)
	_10939058860032000 = decimal.New(10939058860032000, 0)
)

/*
piChudnovskyBrothers() return PI using binary splitting on the series definition of
Chudnovsky Brothers

                426880*sqrt(10005)
	Pi = --------------------------------
		  13591409*aSum + 545140134*bSum

	where
		           24(6n-5)(2n-1)(6n-1)
			a_n = ----------------------
					 (640320^3)*n^3

	and
			 aSum = sum_(n=0)^n a_n
			 bSum = sum_(n=0)^n a_n*n



a(n) = 1,
b(n) = 1,
p(0) = 1, p(n) = (13591409 + 545140134n)(5 − 6n)(2n − 1)(6n − 1),
q(0) = 1, q(n) = (n^3)(640320^3)/24 for n > 0

*/

func getPiA(ctx decimal.Context) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		//returns A + Bn
		var tmp decimal.Big
		tmp.SetUint64(n)
		ctx.Mul(&tmp, _545140134, &tmp)
		ctx.Add(&tmp, _13591409, &tmp)
		return &tmp
	}
}

func getPiP(ctx decimal.Context) func(n uint64) *decimal.Big {
	var tmp0, tmp1, tmp2, cn, c6n, c2n decimal.Big
	return func(n uint64) *decimal.Big {
		// returns (5-6n)(2n-1)(6n-1)

		if n == 0 {
			return one
		}

		//we'll choose to do normal integer math when small enough
		if n < 504103 {
			return decimal.New((5-6*int64(n))*(2*int64(n)-1)*(6*int64(n)-1), 0)
		}

		cn.SetUint64(n)
		ctx.Mul(&c6n, &cn, six)
		ctx.Mul(&c2n, &cn, two)

		ctx.Sub(&tmp0, five, &c6n)
		ctx.Sub(&tmp1, &c2n, one)
		ctx.Sub(&tmp2, &c6n, one)

		var tmp decimal.Big

		ctx.Mul(&tmp, &tmp0, &tmp1)
		ctx.Mul(&tmp, &tmp, &tmp2)

		return &tmp

	}
}

func getPiB() func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		return one
	}
}

func getPiQ(ctx decimal.Context) func(n uint64) *decimal.Big {
	//returns (0) = 1, q(n) = (n^3)(C^3)/24 for n > 0
	// (C^3)/24 = 10939058860032000
	return func(n uint64) *decimal.Big {
		if n == 0 {
			return decimal.WithContext(ctx).SetUint64(1)
		}

		var nn, tmp decimal.Big
		//when n is super small we can be super fast
		if n < 12 {
			tmp.SetUint64(n * n * n * 10939058860032000)
			return &tmp
		}

		// and until bit larger we can still speed up a portion
		if n < 2642246 {
			tmp.SetUint64(n * n * n)
		} else {
			nn.SetUint64(n)
			ctx.Mul(&tmp, &nn, &nn)
			ctx.Mul(&tmp, &tmp, &nn)
		}
		ctx.Mul(&tmp, &tmp, _10939058860032000)

		return &tmp
	}
}

func piChudnovskyBrothers(z *decimal.Big, ctx decimal.Context) *decimal.Big {
	//since this algorithm's rate of convergence is static calculating the number of
	// iteration required will always be faster than the dynamic method of binarysplit
	var value decimal.Big
	extraPrecision := 16
	calculatingPrecision := ctx.Precision + extraPrecision
	iterationNeeded := uint64(math.Ceil(float64(calculatingPrecision/14.0))) + 1
	ctx2 := decimal.Context{Precision: calculatingPrecision}

	var tmp decimal.Big

	BinarySplit(&value, ctx2, 0, iterationNeeded, getPiA(ctx2), getPiP(ctx2), getPiB(), getPiQ(ctx2))
	s := Sqrt(decimal.WithPrecision(calculatingPrecision), _10005)
	ctx2.Mul(&tmp, _426880, s)
	ctx2.Quo(z, &tmp, &value)

	return z.Round(ctx.Precision)
}
