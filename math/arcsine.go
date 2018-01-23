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
	"github.com/ericlagergren/decimal"
)

//Asin returns the arcsine value in radians.
// Input range : -1 <= value <= 1
// Output range: -pi/2 <= Asin() <= pi/2
// Notes:
//		Asin(-1)		-> -pi/2
//		Asin(1)			->  pi/2
//		Asin(NaN)		->   NaN
//		Asin(|value|>1)	->   NaN
//		Asin(+/-Inf)	->   NaN
func Asin(z *decimal.Big, value *decimal.Big) *decimal.Big {
	// here we'll use the half-angle formula
	// Asin(x) = 2atan(x/(1+sqrt(1-x*x)))
	calculatingPrecision := z.Context.Precision + defaultExtraPrecision

	if value.IsInf(0) || value.IsNaN(0) || one.CmpAbs(value) < 0 {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
		//TODO remove: return nil, fmt.Errorf("input value must be between [-1,1]")
	}

	if one.CmpAbs(value) == 0 {
		piOver2 := Pi(z)
		piOver2.Quo(piOver2, two)

		if value.Signbit() {
			piOver2.Neg(piOver2)
		}

		return z.Set(piOver2)
	}

	xsq := decimal.WithPrecision(calculatingPrecision).Mul(value, value)
	x := xsq.Quo(value, xsq.Add(Sqrt(xsq, xsq.Sub(one, xsq)), one))
	result := Atan(decimal.WithPrecision(calculatingPrecision), x)

	result = result.Mul(two, result)
	return z.Set(result)
}
