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

func TestSin(t *testing.T) {
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
		{"0", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strZero, 100))}, newDecimal("0"), false},
		{"1", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver4, 100))}, newDecimal("0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207863"), false},
		{"2", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver4, 100))}, newDecimal("-0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207863"), false},
		{"3", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver3, 100))}, newDecimal("0.8660254037844386467637231707529361834714026269051903140279034897259665084544000185405730933786242877"), false},
		{"4", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver3, 100))}, newDecimal("-0.8660254037844386467637231707529361834714026269051903140279034897259665084544000185405730933786242877"), false},
		{"5", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPiOver2, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"6", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPiOver2, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"7", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str3PiOver4, 100))}, newDecimal("0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207871"), false},
		{"8", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str3PiOver4, 100))}, newDecimal("-0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207871"), false},
		{"9", args{decimal.WithPrecision(100), newDecimal(valuePosStr(strPi, 100))}, newDecimal("9.821480865132823066470938446095505822317253594081284811174502841027019385211055596446229489549303820E-100"), false},
		{"10", args{decimal.WithPrecision(100), newDecimal(valueNegStr(strPi, 100))}, newDecimal("-9.821480865132823066470938446095505822317253594081284811174502841027019385211055596446229489549303820E-100"), false},
		{"11", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str5PiOver4, 100))}, newDecimal("-0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207857"), false},
		{"12", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str5PiOver4, 100))}, newDecimal("0.7071067811865475244008443621048490392848359376884740365883398689953662392310535194251937671638207857"), false},
		{"13", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str3PiOver2, 100))}, newDecimal("-1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"14", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str3PiOver2, 100))}, newDecimal("1.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), false},
		{"15", args{decimal.WithPrecision(100), newDecimal(valuePosStr(str2Pi, 100))}, newDecimal("-9.642961730265646132941876892191011644634507188162569622349005682054038770422111192892458979098607639E-100"), false},
		{"16", args{decimal.WithPrecision(100), newDecimal(valueNegStr(str2Pi, 100))}, newDecimal("9.642961730265646132941876892191011644634507188162569622349005682054038770422111192892458979098607639E-100"), false},
		{"17", args{decimal.WithPrecision(20), newDecimal("7.3303828583761842231")}, newDecimal("0.86602540378443864677"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sin(tt.args.z, tt.args.theta)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sin(%v) error = %v\nwantErr %v\n", tt.args.theta, err, tt.wantErr)
				return
			}
			diff := decimal.WithPrecision(tt.args.z.Context.Precision).Sub(tt.want, got)
			errorBounds := decimal.New(1, tt.args.z.Context.Precision)
			if errorBounds.CmpAbs(diff) <= 0 {
				t.Errorf("Sin(%v) = %v\nwant %v\ndiff: %v\n", tt.args.theta, got, tt.want, diff)

			}
		})
	}
}

func BenchmarkSine(b *testing.B) {
	precision := 30
	theta := newDecimal(valuePosStr(strPiOver4, uint(precision)))
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Sin(result, theta)
	}
}
