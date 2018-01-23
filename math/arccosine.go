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

//Acos returns the arccosine value in radians.
// Input range : -1 <= value <= 1
// Output range: 0 <= Acos() <= pi
// Notes:
//		Acos(-1)			->  pi
//		Acos(1)				->   0
//		Acos(NaN) 			-> NaN
//		Acos(|value| > 1)	-> NaN
//		Acos(+/-Inf)		-> NaN
func Acos(z *decimal.Big, value *decimal.Big) *decimal.Big {
	// here we'll use the half-angle formula
	// Acos(x) = pi/2 - arcsin(x)
	calculatingPrecision := z.Context.Precision + defaultExtraPrecision

	if value.IsInf(0) || one.CmpAbs(value) < 0 || value.IsNaN(0) {
		z.Context.Conditions |= decimal.InvalidOperation
		return z.SetNaN(false)
	}

	if one.CmpAbs(value) == 0 {
		if value.Signbit() {
			return Pi(z)
		}
		return z.SetMantScale(0, 0)
	}

	result := Asin(decimal.WithPrecision(calculatingPrecision), value)

	piOver2 := Pi(decimal.WithPrecision(calculatingPrecision))
	piOver2.Quo(piOver2, two)
	result = result.Sub(piOver2, result)
	return z.Set(result)
}
