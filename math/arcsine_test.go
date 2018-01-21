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

func TestArcsin(t *testing.T) {
	type args struct {
		z     *decimal.Big
		value *decimal.Big
	}
	tests := []struct {
		name    string
		args    args
		want    *decimal.Big
		wantErr bool
	}{
		{"0", args{decimal.WithPrecision(100), newDecimal("0")}, newDecimal("0"), false},
		{"1", args{decimal.WithPrecision(100), newDecimal("1.00")}, newDecimal("1.570796326794896619231321691639751442098584699687552910487472296153908203143104499314017412671058534"), false},
		{"2", args{decimal.WithPrecision(100), newDecimal("-1.00")}, newDecimal("-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499314017412671058534"), false},
		{"3", args{decimal.WithPrecision(100), newDecimal("0.5")}, newDecimal("0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331046724708903528447"), false},
		{"4", args{decimal.WithPrecision(100), newDecimal("-0.50")}, newDecimal("-0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331046724708903528447"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Arcsin(tt.args.z, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Arcsin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			diff := decimal.WithPrecision(tt.args.z.Context.Precision).Sub(tt.want, got)
			errorBounds := decimal.New(1, tt.args.z.Context.Precision)
			if errorBounds.CmpAbs(diff) <= 0 {
				t.Errorf("Arcsin(%v) = %v\nwant %v\ndiff: %v\n", tt.args.value, got, tt.want, diff)

			}
		})
	}
}
