package math

/*
Copyright 2018 W. Nathan Hack

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
	list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
	this list of conditions and the following disclaimer in the documentation and/or
	other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be
	used to endorse or promote products derived from this software without specific
	prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT
SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
import (
	"fmt"
	stdMath "math"

	"github.com/ericlagergren/decimal"
)

var cosine4NMaxN = uint64((stdMath.Sqrt(4.0*float64(stdMath.MaxUint64)+1.0) + 1.0) / 4.0)

func prepareCosineInput(precision int, xValue *decimal.Big) (*decimal.Big, int, error) {
	c2Pi := Pi(decimal.WithPrecision(precision))
	c2Pi = c2Pi.Mul(c2Pi, two)
	var x *decimal.Big
	calculatingPrecision := precision
	//we need to make sure the value we're working with a value is closer to zero (for better results)
	if c2Pi.CmpAbs(xValue) < 0 {
		//for cos to work correctly the input
		// must be |xValue|< 2Pi so we'll fix it

		v := decimal.WithPrecision(calculatingPrecision).QuoInt(xValue, c2Pi)
		vInt, ok := v.Int64()
		if !ok {
			return nil, 0, fmt.Errorf("theta input value was to large")
		}
		//now we'll resize Pi to be a more accurate precision
		piPrecision := precision + int(stdMath.Ceil(stdMath.Abs(float64(vInt))/float64(10)))
		piMultiplier := decimal.New(2*vInt, 0)
		toRemove := Pi(decimal.WithPrecision(piPrecision))
		toRemove = toRemove.Mul(toRemove, piMultiplier)
		//so toRemove = 2*Pi*vInt so
		// xValue-toRemove < 2*Pi
		//add 1 to the precision for the up comming squaring
		x = decimal.WithPrecision(calculatingPrecision).Sub(xValue, toRemove)
	} else {
		//add 1 to the precision for the up comming squaring
		x = decimal.WithPrecision(calculatingPrecision).Copy(xValue)
	}
	//now preform the half angle
	// we'll repeat this up to log(precision)/log(2) and keep track
	// since we'll be dividing x we need to give a bit more precision
	//we'll be repeated applying the double angle formula
	// we could use a higher angle formula but wouldn't buy us any thing.
	// Each time we half we'll have to increase the values precision by 1
	// and since we'll dividing at most 11 time that means at most 11 digits
	// but we'll figure out the minimum time we'll apply the double angle
	// formula

	// we'll we reduce until it's x <= p where p= 1/2^8 (approx 0.0039)
	// we figure out the number of time to divide by solving for r
	// in x/p = 2^r  so r = log(x/p)/log(2)

	xFloat64, _ := x.Float64()
	xFloat64 = stdMath.Abs(xFloat64)
	halfed := 0
	if xFloat64 > 0 {
		p := 1 / stdMath.Pow(2, 11)
		halfed = int(stdMath.Ceil(stdMath.Log(xFloat64/p) / stdMath.Log(2)))
		//if the value is already less than 1/2^r then it's neg
		//so we'll reset to zero
		if halfed < 0 {
			halfed = 0
		}
		//we increase the precision based on the number of divides
		calculatingPrecision += halfed

		if halfed > 0 {
			x = decimal.WithPrecision(calculatingPrecision).Copy(x)
			x = x.Quo(x, decimal.New(int64(stdMath.Pow(2, float64(halfed))), 0))
		}
	}
	//lastly do the
	x.Mul(x, x)

	return x.Neg(x), halfed, nil
}

func getCosineA() func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		return one
	}
}

func getCosineP(negXSq *decimal.Big) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		if n == 0 {
			return one
		}
		return negXSq
	}
}

func getCosineB() func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		return one
	}
}

func getCosineQ(precision int) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		//(0) = 1, q(n) = 2n(2n-1) for n > 0
		if n == 0 {
			return one
		}

		//most of the time n will be a small number so
		// use the fastest method to calculate 2n(2n-1)
		if n < cosine4NMaxN {
			return new(decimal.Big).SetUint64(((n * n) << 2) - (n << 1))
		}
		c2N := new(decimal.Big).SetUint64(n)
		c2N.Mul(c2N, two)
		c2NMinus1 := decimal.WithPrecision(precision).Sub(c2N, one)
		return decimal.WithPrecision(precision).Mul(c2N, c2NMinus1)
	}
}

//Cos returns the cosine of theta(radians).
// Input range : all real numbers
// Output range: -1 <= Cos() <= 1
// Notes:
//		Cos(-Inf) ->   NaN
//		Cos(Inf)  ->   NaN
//		Cos(NaN)  ->   NaN
//		Cos(nil)  -> error
func Cos(z *decimal.Big, theta *decimal.Big) (*decimal.Big, error) {
	calculatingPrecision := z.Context.Precision + defaultExtraPrecision

	if theta == nil {
		return nil, fmt.Errorf("there was an error, input value was nil")
	}

	if theta.IsInf(0) || theta.IsNaN(0) {
		return decimal.WithPrecision(z.Context.Precision).SetNaN(theta.Signbit()), nil
	}

	negXSq, halfed, err := prepareCosineInput(calculatingPrecision, theta)
	if err != nil {
		return nil, fmt.Errorf("could not prepare value %v, there was an error %v", theta, err)
	}

	calculatingPrecision = calculatingPrecision + halfed
	result, err := BinarySplitDynamicCalculate(calculatingPrecision,
		getCosineA(), getCosineP(negXSq), getCosineB(), getCosineQ(calculatingPrecision))

	if err != nil {
		return nil, fmt.Errorf("could not calculate Cos(%v), there was an error %v", theta, err)
	}

	//now undo the half angle bit
	for i := 0; i < halfed; i++ {
		result = result.Mul(result, result)
		result = result.Mul(result, two)
		result = result.Sub(result, one)
	}

	return result.Round(z.Context.Precision), nil
}
