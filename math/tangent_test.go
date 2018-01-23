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
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestTan(t *testing.T) {
	type args struct {
		z     *decimal.Big
		theta *decimal.Big
	}
	tests := []struct {
		name    string
		args    args
		want    *decimal.Big
	}{
		// note the expected values came from wolframalpha.com
		// the first group is    |theta|< Pi/2
		{"0", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strZero, 100))}, newDecimal("0")},
		{"1", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver4, 100))}, newDecimal("0.9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998")},
		{"2", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver4, 100))}, newDecimal("-0.9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998")},
		{"3", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574")},
		{"4", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574")},

		// "near" pi/2 .. because at p/2 with infinite precision the result would be Inf
		{"5", args{decimal.WithPrecision(26), newDecimal("1.4000000000000000000000000")}, newDecimal("5.7978837154828896437077202")},
		{"6", args{decimal.WithPrecision(26), newDecimal("-1.4000000000000000000000000")}, newDecimal("-5.7978837154828896437077202")},
		{"7", args{decimal.WithPrecision(17), newDecimal(valuePosStr(strPiOver2, 17))}, newDecimal("5.1998506188720271E16")},
		{"8", args{decimal.WithPrecision(17), newDecimal(valueNegStr(strPiOver2, 17))}, newDecimal("-5.1998506188720271E16")},
		{"9", args{decimal.WithPrecision(30), newDecimal(valuePosStr(strPiOver2, 30))}, newDecimal("1.02548934802693187243286598565E+29")},
		{"10", args{decimal.WithPrecision(30), newDecimal(valueNegStr(strPiOver2, 30))}, newDecimal("-1.02548934802693187243286598565E+29")},
		{"11", args{decimal.WithPrecision(50), newDecimal(valuePosStr(strPiOver2, 50))}, newDecimal("1.8899844771296019351604660349192831433443369043068E49")},
		{"12", args{decimal.WithPrecision(50), newDecimal(valueNegStr(strPiOver2, 50))}, newDecimal("-1.8899844771296019351604660349192831433443369043068E49")},
		{"13", args{decimal.WithPrecision(60), newDecimal(valuePosStr(strPiOver2, 60))}, newDecimal("4.35510876003321014579598280859551726015883372322953933150280E59")},
		{"14", args{decimal.WithPrecision(60), newDecimal(valueNegStr(strPiOver2, 60))}, newDecimal("-4.35510876003321014579598280859551726015883372322953933150280E59")},

		//the next group is   Pi/2 < |theta|
		{"15", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str2PiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578")},
		{"16", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str2PiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578")},
		{"17", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str3PiOver4, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002")},
		{"18", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str3PiOver4, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002")},
		{"19", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPi, 100))}, newDecimal("0")},
		{"20", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPi, 100))}, newDecimal("0")},
		{"21", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str5PiOver4, 100))}, newDecimal("0.999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998")},
		{"22", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str5PiOver4, 100))}, newDecimal("-0.999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998")},
		{"23", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str4PiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574")},
		{"24", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str4PiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574")},
		{"25", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str5PiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578")},
		{"26", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str5PiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578")},
		{"27", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str7PiOver4, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002")},
		{"28", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str7PiOver4, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002")},
		{"29", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str2Pi, 100))}, newDecimal("0")},
		{"30", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str2Pi, 100))}, newDecimal("0")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tan(tt.args.z, tt.args.theta)
			diff := decimal.WithPrecision(tt.args.z.Context.Precision).Sub(tt.want, got)
			errorBounds := decimal.New(1, tt.args.z.Context.Precision)

			if errorBounds.CmpAbs(diff) <= 0 {
				t.Errorf("Tan(%v) = %v\nwant: %v\ndiff: %v\n", tt.args.theta, got, tt.want, diff)
				piOver2 := Pi(decimal.WithPrecision(tt.args.z.Context.Precision))
				piOver2.Quo(piOver2, two)
				dd := tt.args.theta.Sub(tt.args.theta, piOver2)
				t.Errorf("Pi: %v\n", piOver2)
				t.Errorf("dd: %v\n", dd)
			}
		})
	}
}

func BenchmarkTangent(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPiOver4, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Tan(result, theta)
	}
}
