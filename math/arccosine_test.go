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

func TestAcos(t *testing.T) {
	type args struct {
		z     *decimal.Big
		value *decimal.Big
	}
	tests := []struct {
		name    string
		args    args
		want    *decimal.Big
	}{
		{"0", args{decimal.WithPrecision(100), newDecimal("-1.00")}, newDecimal("3.141592653589793238462643383279502884197169399375105820974944592307816406286208998628034825342117068")},
		{"1", args{decimal.WithPrecision(100), newDecimal("-.9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")}, newDecimal("3.141592653589793238462643383279502884197169399375091678839320861357328389398966901647249128623363298")},
		{"2", args{decimal.WithPrecision(100), newDecimal("-0.50")}, newDecimal("2.094395102393195492308428922186335256131446266250070547316629728205210937524139332418689883561411379")},
		{"3", args{decimal.WithPrecision(100), newDecimal("0")}, newDecimal("1.570796326794896619231321691639751442098584699687552910487472296153908203143104499314017412671058534")},
		{"4", args{decimal.WithPrecision(100), newDecimal("0.5")}, newDecimal("1.047197551196597746154214461093167628065723133125035273658314864102605468762069666209344941780705689")},
		{"5", args{decimal.WithPrecision(100), newDecimal(".9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")}, newDecimal("1.414213562373095048801688724209698078569671875376948073176679737990732478462107038850387534327641573E-50")},
		{"6", args{decimal.WithPrecision(100), newDecimal("1.00")}, newDecimal("0")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Acos(tt.args.z, tt.args.value)
			diff := decimal.WithPrecision(tt.args.z.Context.Precision).Sub(tt.want, got)
			errorBounds := decimal.New(1, tt.args.z.Context.Precision)

			if errorBounds.CmpAbs(diff) <= 0 {
				t.Errorf("Acos(%v) = %v\nwant %v\ndiff: %v\n", tt.args.value, got, tt.want, diff)

			}
		})
	}
}
