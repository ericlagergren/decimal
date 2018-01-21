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

func TestCos(t *testing.T) {
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
		//the input is string values (x)' truncated to specific digit length the result is from those values evaluated @ wolframaplha.com to the same number of digits
		{"0", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strZero, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"1", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver4, 100))}, newDecimal("0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207864"), false},
		{"2", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver4, 100))}, newDecimal("0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207864"), false},
		{"3", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver3, 100))}, newDecimal("0.5000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003"), false},
		{"4", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver3, 100))}, newDecimal("0.5000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003"), false},
		{"5", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver2, 100))}, newDecimal("9.910740432566411533235469223047752911158626797040642405587251420513509692605527798223114744774651910E-100"), false},
		{"6", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver2, 100))}, newDecimal("9.910740432566411533235469223047752911158626797040642405587251420513509692605527798223114744774651910E-100"), false},
		{"7", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str3PiOver4, 100))}, newDecimal("-0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207857"), false},
		{"8", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str3PiOver4, 100))}, newDecimal("-0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207857"), false},
		{"9", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPi, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"10", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPi, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"11", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str5PiOver4, 100))}, newDecimal("-0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207871"), false},
		{"12", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str5PiOver4, 100))}, newDecimal("-0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207871"), false},
		{"13", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str3PiOver2, 100))}, newDecimal("-9.732221297699234599706407669143258733475880391121927216761754261540529077816583394669344234323955729E-100"), false},
		{"14", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str3PiOver2, 100))}, newDecimal("-9.732221297699234599706407669143258733475880391121927216761754261540529077816583394669344234323955729E-100"), false},
		{"15", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str2Pi, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"16", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str2Pi, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Cos(tt.args.z, tt.args.theta)
			diff := decimal.WithPrecision(tt.args.z.Context.Precision).Sub(tt.want, got)
			errorBounds := decimal.New(1, tt.args.z.Context.Precision)

			if (err != nil) != tt.wantErr {
				t.Errorf("Cos(%v) error = %v\nwantErr %v\n", tt.args.theta, err, tt.wantErr)
				t.Errorf("Cos(%v) = %v\nwant %v\ndiff: %v\n", tt.args.theta, got, tt.want, diff)
				return
			}
			if errorBounds.CmpAbs(diff) <= 0 {
				t.Errorf("Cos(%v) = %v\nwant %v\ndiff: %v\n", tt.args.theta, got, tt.want, diff)
			}
		})
	}
}

func BenchmarkCosZero(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strZero, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}

func BenchmarkCosPiOver6(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPiOver6, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}

func BenchmarkCosPiOver5(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPiOver5, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}

func BenchmarkCosPiOver4(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPiOver4, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}

func BenchmarkCosPiOver3(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPiOver3, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}

func BenchmarkCosPiOver2(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPiOver2, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}

func BenchmarkCos2PiOver3(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(str2PiOver3, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}

func BenchmarkCosPi(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPi, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()
	var x *decimal.Big
	for i := 0; i < b.N; i++ {
		x, _ = Cos(result, theta)
	}
	result = x
}

func BenchmarkCos2Pi(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(str2Pi, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Cos(result, theta)
	}
}
