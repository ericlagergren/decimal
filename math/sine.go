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

var (
	sine4NMaxN        = uint64((stdMath.Sqrt(4.0*float64(stdMath.MaxUint64)+1.0) + 1.0) / 4.0)
	sinePrecisionList = make(map[int]int)
)

func prepareSineInput(precision int, xValue *decimal.Big) (*decimal.Big, *decimal.Big, error) {
	c2Pi := Pi(decimal.WithPrecision(precision + defaultExtraPrecision))
	c2Pi = c2Pi.Mul(c2Pi, two)
	var x *decimal.Big
	//we need to make sure the value we're working with a value is closer to zero (for better results)
	if c2Pi.CmpAbs(xValue) < 0 {
		//for cos to work correctly the input
		// must be |xValue|< 2Pi so we'll fix it

		v := decimal.WithPrecision(precision+defaultExtraPrecision).QuoInt(xValue, c2Pi)
		vInt, ok := v.Int64()
		if !ok {
			return nil, nil, fmt.Errorf("theta input value was to large")
		}
		//now we'll resize Pi to be a more accurate precision
		piPrecision := precision + defaultExtraPrecision + int(stdMath.Ceil(stdMath.Abs(float64(vInt))/float64(10)))
		piMultiplier := decimal.New(2*vInt, 0)
		toRemove := Pi(decimal.WithPrecision(piPrecision))
		toRemove = toRemove.Mul(toRemove, piMultiplier)
		//so toRemove = 2*Pi*vInt so
		// xValue-toRemove < 2*Pi
		//add 1 to the precision for the up comming squaring
		x = decimal.WithPrecision(precision+defaultExtraPrecision).Sub(xValue, toRemove)
	} else {
		//add 1 to the precision for the up comming squaring
		x = decimal.WithPrecision(precision + defaultExtraPrecision).Copy(xValue)
	}
	xsq := decimal.WithPrecision(precision+defaultExtraPrecision).Mul(x, x)

	//we need to make sure the value we're working with is closer to zero (better for better results)
	return x.Round(precision), (xsq.Neg(xsq)).Round(precision), nil
}

func getSineA() func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		return one
	}
}

func getSineP(x, negXSq *decimal.Big) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		if n == 0 {
			return x
		}
		return negXSq
	}
}

func getSineB() func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		return one
	}
}

func getSineQ(precision int) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		//(0) = 1, q(n) = 2n(2n+1) for n > 0
		if n == 0 {
			return one
		}

		//most of the time n will be a small number so
		// use the fastest method to calculate 2n(2n+1)
		if n < sine4NMaxN {
			return new(decimal.Big).SetUint64(((n * n) << 2) + (n << 1))
		}
		c2N := new(decimal.Big).SetUint64(n)
		c2N.Mul(c2N, two)
		c2NPlus1 := decimal.WithPrecision(precision).Add(c2N, one)
		return decimal.WithPrecision(precision).Mul(c2N, c2NPlus1)
	}
}

//Sin returns the sine of theta(radians).
// Input range : all real numbers
// Output range: -1 <= Sin() <= 1
// Notes:
//  	Sin(-Inf) ->   NaN
//		Sin(Inf)  ->   NaN
//      Sin(NaN)  ->   NaN
//		Sin(nil)  -> error
func Sin(z *decimal.Big, theta *decimal.Big) (*decimal.Big, error) {
	calculatingPrecision := z.Context.Precision + defaultExtraPrecision

	if theta == nil {
		return nil, fmt.Errorf("there was an error, input value was nil")
	}

	if theta.IsInf(0) || theta.IsNaN(0) {
		return decimal.WithPrecision(z.Context.Precision).SetNaN(theta.Signbit()), nil
	}

	piOver2 := Pi(decimal.WithPrecision(calculatingPrecision))
	piOver2.Quo(piOver2, two)

	result, err := Cos(decimal.WithPrecision(calculatingPrecision), piOver2.Sub(piOver2, theta))
	if err != nil {
		return nil, fmt.Errorf("could not calculate Sin(%v), there was an error %v", theta, err)
	}

	return z.Set(result), nil
}
