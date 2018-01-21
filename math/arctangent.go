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

var atanMax = uint64((stdMath.MaxUint64 - 1) / 2)

func prepareAtanInput(precision int, xValue *decimal.Big) (*decimal.Big, *decimal.Big, *decimal.Big, int, int) {
	calculatingPrecision := precision + defaultExtraPrecision
	x := decimal.WithPrecision(calculatingPrecision).Copy(xValue)

	//since smaller values converge faster
	// we'll use argument reduction
	//if |xValue| > 1 then set x = 1/xValue
	segment := 0
	if one.CmpAbs(xValue) < 0 {
		segment = 2
		x = x.Quo(one, x)
	} else if decimal.New(577350, 6).CmpAbs(xValue) < 0 {
		// if |xValue| > sqrt(3)/3  (we approx to 0.577350)
		segment = 1
		//then set x = (sqrt(3)*x-1)/(sqrt(3)+x)
		sqrt3 := Sqrt(decimal.WithPrecision(calculatingPrecision), three)
		sqrt3XMinus1 := decimal.WithPrecision(calculatingPrecision).Mul(sqrt3, x)
		sqrt3XMinus1.Sub(sqrt3XMinus1, one)
		sqrt3PlusX := x.Add(sqrt3, x)
		x = sqrt3XMinus1.Quo(sqrt3XMinus1, sqrt3PlusX)
	}

	//next we'll use argument halving technic
	// atan(y) = 2 atan(y/(1+sqrt(1+y^2)))
	// we'll repeat this up to a point
	// we have compeating operations at some
	// point the sqrt causes a problem
	// note (http://fredrikj.net/blog/2014/12/faster-high-precision-Atangents/)
	// suggests that r= 8 times is a good value for
	// precision with 1000s of digits to millions
	// however it was easy to determine there is a
	// better sliding window to use instead
	// which is what we use as it turns out
	// when the precision is large, a larger r value
	// compared to 8 is every effective
	xFloat64, _ := x.Float64()
	xFloat64 = stdMath.Abs(xFloat64)
	//the formula simple but works a bit better then a fix value (8)
	r := stdMath.Max(0, stdMath.Ceil(0.31554321636851*stdMath.Pow(float64(precision), 0.654095561044508)))
	p := 1 / stdMath.Pow(2, r)
	halfed := int(stdMath.Ceil(stdMath.Log(xFloat64/p) / stdMath.Log(2)))

	//if the value is already less than 1/(2^r) then halfed
	// will be negative and we don't need to apply
	// the double angle formula because it would hurt performance
	// so we'll set halfed to zero
	if halfed < 0 {
		halfed = 0
	}

	if halfed > 0 {
		x = decimal.WithPrecision(calculatingPrecision).Copy(x)
	}

	for i := 0; i < halfed; i++ {
		xsq := decimal.WithPrecision(calculatingPrecision).Mul(x, x)
		x = x.Quo(x, xsq.Add(Sqrt(xsq, xsq.Add(xsq, one)), one))
	}

	xSquared := decimal.WithPrecision(calculatingPrecision).Mul(x, x)
	xSqPlus1 := decimal.WithPrecision(calculatingPrecision).Add(xSquared, one)
	return x, xSquared, xSqPlus1, segment, halfed
}

func getAtanA(x *decimal.Big) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		return x
	}
}

func getAtanP(precision int, xSq *decimal.Big) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		if n == 0 {
			return one
		}

		c2N := decimal.WithPrecision(precision).SetUint64(n)
		c2N.Mul(c2N, two)
		c2NXSq := c2N.Mul(c2N, xSq)
		return c2NXSq
	}
}

func getAtanB() func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		return one
	}
}

func getAtanQ(precision int, xSqPlus1 *decimal.Big) func(n uint64) *decimal.Big {
	return func(n uint64) *decimal.Big {
		//b(n) = (2n+1) for n >= 0
		//most of the time n will be a small number so
		// use the fastest method to calculate (2n+1)
		var c2NPlus1 *decimal.Big
		if n < atanMax {
			c2NPlus1 = new(decimal.Big).SetUint64((n << 1) + 1)
		}
		c2N := decimal.WithPrecision(precision).SetUint64(n)
		c2N.Mul(c2N, two)
		c2NPlus1 = decimal.WithPrecision(precision).Add(c2N, one)
		return c2NPlus1.Mul(c2NPlus1, xSqPlus1)
	}
}

//Atan returns the arctangent value in radians.
// Input range : all real numbers
// Output range: -pi/2 <= Atan() <= pi/2
// Notes:
//		Atan(-Inf) -> -pi/2
//		Atan(Inf)  ->  pi/2
//		Atan(NaN)  ->   NaN
//		Atan(nil)  -> error
func Atan(z *decimal.Big, value *decimal.Big) (*decimal.Big, error) {
	calculatingPrecision := z.Context.Precision + defaultExtraPrecision

	if value == nil {
		return nil, fmt.Errorf("there was an error, input value was nil")
	}

	if value.IsInf(0) {
		z := Pi(z)
		z.Quo(z, two)

		if value.IsInf(-1) {
			z.Neg(z)
		}

		return z, nil
	}

	if value.IsNaN(0) {
		return decimal.WithPrecision(z.Context.Precision).SetNaN(value.Signbit()), nil
	}

	y, ySq, ySqPlus1, segment, halfed := prepareAtanInput(calculatingPrecision, value)
	result, err := BinarySplitDynamicCalculate(calculatingPrecision,
		getAtanA(y), getAtanP(calculatingPrecision, ySq), getAtanB(), getAtanQ(calculatingPrecision, ySqPlus1))

	if err != nil {
		return nil, fmt.Errorf("could not calculate Atan(%v), there was an error %v", value, err)
	}

	//undo the double angle part
	twoMultiplier := Pow(decimal.WithPrecision(calculatingPrecision), two, decimal.New(int64(halfed), 0))
	result = result.Mul(result, twoMultiplier)

	// to handle the argument reduction step
	//check which segment the value was from
	// seg 0 : 0 < value <= sqrt(3)/3  // then result = result
	// seg 1 : sqrt(3)/3 < value <= 1  // then result = pi/6 + result
	// set 2 : 1 < value               // then result = pi/2 - result
	switch segment {
	case 1:
		piOver6 := Pi(decimal.WithPrecision(calculatingPrecision))
		piOver6.Quo(piOver6, six)
		result = decimal.WithPrecision(calculatingPrecision).Add(piOver6, result)
	case 2:
		piOver2 := Pi(decimal.WithPrecision(calculatingPrecision))
		piOver2.Quo(piOver2, two)
		result = decimal.WithPrecision(calculatingPrecision).Sub(piOver2, result)
	}
	return z.Set(result), nil
}
