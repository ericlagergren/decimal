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
		wantErr bool
	}{
		// note the expected values came from wolframalpha.com
		// the first group is    |theta|< Pi/2
		{"0", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strZero, 100))}, newDecimal("0"), false},
		{"1", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver4, 100))}, newDecimal("0.9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998"), false},
		{"2", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver4, 100))}, newDecimal("-0.9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998"), false},
		{"3", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574"), false},
		{"4", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574"), false},

		// "near" pi/2 .. because at p/2 with infinite precision the result would be Inf
		{"5", args{decimal.WithPrecision(26), newDecimal("1.4000000000000000000000000")}, newDecimal("5.7978837154828896437077202"), false},
		{"6", args{decimal.WithPrecision(26), newDecimal("-1.4000000000000000000000000")}, newDecimal("-5.7978837154828896437077202"), false},
		{"7", args{decimal.WithPrecision(17), newDecimal(valuePosStr(strPiOver2, 17))}, newDecimal("5.1998506188720271E16"), false},
		{"8", args{decimal.WithPrecision(17), newDecimal(valueNegStr(strPiOver2, 17))}, newDecimal("-5.1998506188720271E16"), false},
		{"9", args{decimal.WithPrecision(30), newDecimal(valuePosStr(strPiOver2, 30))}, newDecimal("1.02548934802693187243286598565E+29"), false},
		{"10", args{decimal.WithPrecision(30), newDecimal(valueNegStr(strPiOver2, 30))}, newDecimal("-1.02548934802693187243286598565E+29"), false},
		{"11", args{decimal.WithPrecision(50), newDecimal(valuePosStr(strPiOver2, 50))}, newDecimal("1.8899844771296019351604660349192831433443369043068E49"), false},
		{"12", args{decimal.WithPrecision(50), newDecimal(valueNegStr(strPiOver2, 50))}, newDecimal("-1.8899844771296019351604660349192831433443369043068E49"), false},
		{"13", args{decimal.WithPrecision(60), newDecimal(valuePosStr(strPiOver2, 60))}, newDecimal("4.35510876003321014579598280859551726015883372322953933150280E59"), false},
		{"14", args{decimal.WithPrecision(60), newDecimal(valueNegStr(strPiOver2, 60))}, newDecimal("-4.35510876003321014579598280859551726015883372322953933150280E59"), false},

		//the next group is   Pi/2 < |theta|
		{"15", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str2PiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578"), false},
		{"16", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str2PiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578"), false},
		{"17", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str3PiOver4, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002"), false},
		{"18", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str3PiOver4, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002"), false},
		{"19", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPi, 100))}, newDecimal("0"), false},
		{"20", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPi, 100))}, newDecimal("0"), false},
		{"21", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str5PiOver4, 100))}, newDecimal("0.999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998"), false},
		{"22", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str5PiOver4, 100))}, newDecimal("-0.999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999998"), false},
		{"23", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str4PiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574"), false},
		{"24", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str4PiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248574"), false},
		{"25", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str5PiOver3, 100))}, newDecimal("-1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578"), false},
		{"26", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str5PiOver3, 100))}, newDecimal("1.732050807568877293527446341505872366942805253810380628055806979451933016908800037081146186757248578"), false},
		{"27", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str7PiOver4, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002"), false},
		{"28", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str7PiOver4, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002"), false},
		{"29", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str2Pi, 100))}, newDecimal("0"), false},
		{"30", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str2Pi, 100))}, newDecimal("0"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Tan(tt.args.z, tt.args.theta)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

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
