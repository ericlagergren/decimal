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

func prepareTangentInput(precision int, theta *decimal.Big) (*decimal.Big, error) {
	cPi := Pi(decimal.WithPrecision(precision))
	cPiOver2 := decimal.WithPrecision(precision).Quo(cPi, two)
	var x *decimal.Big
	if cPiOver2.CmpAbs(theta) < 0 {
		//for cos to work correctly the input
		// must be |theta|< 2Pi so we'll fix it
		var m *decimal.Big
		if theta.Signbit() {
			m = decimal.WithPrecision(precision).Add(theta, cPiOver2)
		} else {
			m = decimal.WithPrecision(precision).Sub(theta, cPiOver2)
		}
		m = decimal.WithPrecision(precision).QuoInt(m, cPi)
		if theta.Signbit() {
			m.Sub(m, one)
		} else {
			m.Add(m, one)
		}
		mInt, ok := m.Int64()
		if !ok {
			return nil, fmt.Errorf("theta input value was to large")
		}
		//now we'll resize Pi to be a more accurate precision
		piPrecision := precision + int(stdMath.Ceil(stdMath.Abs(float64(mInt))/float64(10)))
		toRemove := Pi(decimal.WithPrecision(piPrecision))
		toRemove = toRemove.Mul(toRemove, m)

		//so toRemove = mPi so
		// |theta-toRemove| < Pi/2
		//add 1 to the precision for the up comming squaring
		x = decimal.WithPrecision(precision+1).Sub(theta, toRemove)
	} else {
		//add 1 to the precision for the up comming squaring
		x = decimal.WithPrecision(precision + 1).Copy(theta)
	}
	return x, nil
}

//Tan returns the tangent of theta(radians).
// Input range : -pi/2 <= theta <= pi/2
// Output range: all real numbers
// Notes:
//		Tan(-Inf) ->   NaN
//		Tan(Inf)  ->   NaN
//		Tan(NaN)  ->   NaN
//		Tan(nil)  -> error
func Tan(z *decimal.Big, theta *decimal.Big) (*decimal.Big, error) {
	//tan(x) = sign(x)*sqrt(1/cos(x)^2-1)
	if theta == nil {
		return nil, fmt.Errorf("there was an error, input value was nil")
	}

	if theta.IsInf(0) || theta.IsNaN(0) {
		return decimal.WithPrecision(z.Context.Precision).SetNaN(theta.Signbit()), nil
	}

	calculatingPrecision := z.Context.Precision + defaultExtraPrecision
	x, err := prepareTangentInput(calculatingPrecision, theta)
	if err != nil {
		return nil, fmt.Errorf("could not prepare value %v, there was an error %v", theta, err)
	}

	// tangent has an asymptote at pi/2 and we'll need more precision as we get closer
	// the reason we need it is as we approach pi/2 is due to the squaring portion,
	// it will cause the small value of Cosine to be come extremely small.
	// we COULD fix it by simply doubling the precision, however,
	// when the precision gets larger it will be a significant impact
	// to performance, instead we'll only add the extra precision when we need it
	// by using the difference to see how much extra precision we need
	// we'll speed things up by only using a quick compare to see if we need
	// to do a deeper inspection
	tmpCalculation := calculatingPrecision
	if decimal.New(14, 1).CmpAbs(theta) < 0 {
		piOver2 := Pi(decimal.WithPrecision(calculatingPrecision))
		piOver2.Quo(piOver2, two)
		var dd *decimal.Big
		if theta.Signbit() {
			dd = decimal.WithPrecision(calculatingPrecision).Add(theta, piOver2)
		} else {
			dd = decimal.WithPrecision(calculatingPrecision).Sub(theta, piOver2)
		}
		tmpCalculation = calculatingPrecision + dd.Scale() - dd.Precision()
	}

	result, err := Cos(decimal.WithPrecision(tmpCalculation), x)
	if err != nil {
		return nil, fmt.Errorf("could not calculate Tan(%v), there was an error %v", x, err)
	}

	result = decimal.WithPrecision(calculatingPrecision).Copy(result)
	result = result.Mul(result, result)
	result = result.Quo(one, result)
	result = result.Sub(result, one)
	result = Sqrt(decimal.WithPrecision(calculatingPrecision), result)

	if x.Signbit() {
		result.Neg(result)
	}

	return z.Set(result), nil
}
