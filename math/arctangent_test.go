package math

import (
	"testing"

	"github.com/ericlagergren/decimal"
)

func TestAtan(t *testing.T) {
	type args struct {
		z     *decimal.Big
		theta *decimal.Big
	}
	tests := []struct {
		name string
		args args
		want *decimal.Big
	}{
		{"0", args{decimal.WithPrecision(100), newDecimal("0")}, newDecimal("0")},
		{"1", args{decimal.WithPrecision(100), newDecimal(".500")}, newDecimal("0.4636476090008061162142562314612144020285370542861202638109330887201978641657417053006002839848878926")},
		{"2", args{decimal.WithPrecision(100), newDecimal("-.500")}, newDecimal("-0.4636476090008061162142562314612144020285370542861202638109330887201978641657417053006002839848878926")},
		{"3", args{decimal.WithPrecision(100), newDecimal(".9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")}, newDecimal("0.7853981633974483096156608458198757210492923498437764552437361480769541015715522496570087063355292669")},
		{"3", args{decimal.WithPrecision(100), newDecimal("-.9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")}, newDecimal("-0.7853981633974483096156608458198757210492923498437764552437361480769541015715522496570087063355292669")},
		{"4", args{decimal.WithPrecision(100), newDecimal("1.00")}, newDecimal("0.7853981633974483096156608458198757210492923498437764552437361480769541015715522496570087063355292670")},
		{"5", args{decimal.WithPrecision(100), newDecimal("-1.00")}, newDecimal("-0.7853981633974483096156608458198757210492923498437764552437361480769541015715522496570087063355292670")},
		{"6", args{decimal.WithPrecision(100), newDecimal(".999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")}, newDecimal("0.785398163397448309615660845819875721049292349843776455243736148076954101571552249657008706335529266995537021628320576661773461152387645557931339852032120279362571025675484630276389")},

		{"7", args{decimal.WithPrecision(100), newDecimal("100.0")}, newDecimal("1.560796660108231381024981575430471893537215347143176270859532877957451649939045719334570767484384444")},
		{"8", args{decimal.WithPrecision(100), newDecimal("0.9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")}, newDecimal("0.7853981633974483096156608458198757210492923498437764552437361480769541015715522496570087063355292669955370216283205766617734611523876455579313398520321202793625710256754846302763899111557372387325954911072027439164833615321189120584466957913178004772864121417308650871526135816620533484018150622853184311467516515788970437203802302407073135229288410919731475900028326326372051166303460367379853779023582643175914398979882730465293454831529482762796370186155949906873918379714381812228069845457529872824584183406101641607715053487365988061842976755449652359256926348042940732941880961687046169173512830001420317863158902069464428356894474022934092946803671102253062383575366373963427626980699223147308855049890280322554902160086045399534074436928274901296768028374999995932445124877649329332040240796487561148638367270756606305770633361712588154827970427525007844596882216468833020953551542944172868258995633726071888671827898907159705884468984379894454644451330428067016532504819691527989773041050497")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Atan(tt.args.z, tt.args.theta)
			diff := decimal.WithPrecision(tt.args.z.Context.Precision).Sub(tt.want, got)
			errorBounds := decimal.New(1, tt.args.z.Context.Precision)

			if errorBounds.CmpAbs(diff) <= 0 {
				t.Errorf("Atan(%v) = %v\nwant %v\ndiff: %v\n", tt.args.theta, got, tt.want, diff)

			}
		})
	}
}

func BenchmarkAtan(b *testing.B) {
	precision := 30
	four := decimal.New(4, 0)
	piOver4 := Pi(decimal.WithPrecision(precision))
	piOver4.Quo(piOver4, four)
	result := decimal.WithPrecision(precision)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Atan(result, piOver4)
	}
}

func TestAtan2(t *testing.T) {
	naN := new(decimal.Big).SetNaN(false)
	naN.Context.Conditions |= decimal.InvalidOperation
	zeroNeg, _ := new(decimal.Big).SetString("-0")
	piPos := Pi(new(decimal.Big))
	piNeg := new(decimal.Big).Neg(piPos)
	oneNeg := new(decimal.Big).Neg(one)
	piOver2 := Pi(new(decimal.Big))
	piOver2.Quo(piOver2, two)
	piOver2Neg := new(decimal.Big).Neg(piOver2)
	infPlus := new(decimal.Big).SetInf(false)
	infNeg := new(decimal.Big).SetInf(true)
	piOver4 := Pi(new(decimal.Big))
	piOver4.Quo(piOver4, four)
	piOver4Neg := new(decimal.Big).Neg(piOver4)
	threePiOver4 := new(decimal.Big).Mul(piOver4, three)
	threePiOver4Neg := new(decimal.Big).Mul(piOver4Neg, three)
	oneOverOneAtan := Atan(new(decimal.Big), one)
	oneOverOneNegAtan := Atan(new(decimal.Big), new(decimal.Big).Quo(one, oneNeg))
	zeroOverOneNegAtan := Atan(new(decimal.Big), new(decimal.Big).Quo(zero, oneNeg))
	oneNegOverOneAtan := Atan(new(decimal.Big), new(decimal.Big).Quo(oneNeg, one))
	oneNegOverOneNegAtan := Atan(new(decimal.Big), new(decimal.Big).Quo(oneNeg, oneNeg))
	zeroOverOneNegAtanPlusPi := new(decimal.Big).Add(zeroOverOneNegAtan, Pi(new(decimal.Big)))
	oneOverOneNegAtanPlusPi := new(decimal.Big).Add(oneOverOneNegAtan, Pi(new(decimal.Big)))
	oneNegOverOneNegAtanMinusPi := new(decimal.Big).Sub(oneNegOverOneNegAtan, Pi(new(decimal.Big)))
	type args struct {
		z *decimal.Big
		y *decimal.Big
		x *decimal.Big
	}
	tests := []struct {
		name string
		args args
		want *decimal.Big
	}{
		// Atan2(NaN, NaN) -> NaN
		{"0", args{new(decimal.Big), naN, naN}, naN},

		// Atan2(y, NaN) -> NaN
		{"1", args{new(decimal.Big), zero, naN}, naN},
		{"2", args{new(decimal.Big), one, naN}, naN},

		// Atan2(NaN, x) -> NaN
		{"3", args{new(decimal.Big), naN, zero}, naN},
		{"4", args{new(decimal.Big), naN, one}, naN},

		// Atan2(+/-0, x>=0) -> +/-0
		{"5", args{new(decimal.Big), zero, zero}, zero},
		{"6", args{new(decimal.Big), zeroNeg, zero}, zeroNeg},
		{"7", args{new(decimal.Big), zero, one}, zero},
		{"8", args{new(decimal.Big), zeroNeg, one}, zeroNeg},

		// Atan2(+/-0, x<=-0) -> +/-pi
		{"9", args{new(decimal.Big), zero, zeroNeg}, piPos},
		{"10", args{new(decimal.Big), zeroNeg, zeroNeg}, piNeg},
		{"11", args{new(decimal.Big), zero, oneNeg}, piPos},
		{"12", args{new(decimal.Big), zeroNeg, oneNeg}, piNeg},

		// Atan2(y>0, 0) -> +pi/2
		{"13", args{new(decimal.Big), one, zero}, piOver2},

		// Atan2(y<0, 0) -> -pi/2
		{"14", args{new(decimal.Big), oneNeg, zero}, piOver2Neg},

		// Atan2(+/-Inf, +Inf) -> +/-pi/4
		{"15", args{new(decimal.Big), infPlus, infPlus}, piOver4},
		{"16", args{new(decimal.Big), infNeg, infPlus}, piOver4Neg},

		// Atan2(+/-Inf, -Inf) -> +/-3pi/4
		{"17", args{new(decimal.Big), infPlus, infNeg}, threePiOver4},
		{"18", args{new(decimal.Big), infNeg, infNeg}, threePiOver4Neg},

		// Atan2(y, +Inf) -> 0
		{"19", args{new(decimal.Big), oneNeg, infPlus}, zero},
		{"20", args{new(decimal.Big), zero, infPlus}, zero},
		{"21", args{new(decimal.Big), one, infPlus}, zero},

		// Atan2(y>0, -Inf) -> +pi
		{"22", args{new(decimal.Big), one, infNeg}, piPos},

		// Atan2(y<0, -Inf) -> -pi
		{"23", args{new(decimal.Big), oneNeg, infNeg}, piNeg},

		// Atan2(+/-Inf, x) -> +/-pi/2
		{"24", args{new(decimal.Big), infPlus, oneNeg}, piOver2},
		{"25", args{new(decimal.Big), infNeg, oneNeg}, piOver2Neg},
		{"26", args{new(decimal.Big), infPlus, zero}, piOver2},
		{"27", args{new(decimal.Big), infNeg, zero}, piOver2Neg},
		{"28", args{new(decimal.Big), infPlus, one}, piOver2},
		{"29", args{new(decimal.Big), infNeg, one}, piOver2Neg},

		// Atan2(y,x>0) -> Atan(y/x)
		{"30", args{new(decimal.Big), oneNeg, one}, oneNegOverOneAtan},
		{"31", args{new(decimal.Big), one, one}, oneOverOneAtan},

		// Atan2(y>=0, x<0) -> Atan(y/x) + pi
		{"32", args{new(decimal.Big), zero, oneNeg}, zeroOverOneNegAtanPlusPi},
		{"34", args{new(decimal.Big), one, oneNeg}, oneOverOneNegAtanPlusPi},

		// Atan2(y<0, x<0) -> Atan(y/x) - pi
		{"35", args{new(decimal.Big), oneNeg, oneNeg}, oneNegOverOneNegAtanMinusPi},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if got := Atan2(tt.args.z, tt.args.y, tt.args.x); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Atan2() = %v, want %v", got, tt.want)
			// }
			if got := Atan2(tt.args.z, tt.args.y, tt.args.x); got.Cmp(tt.want) != 0 {
				t.Errorf("Atan2() = %v, want %v", got, tt.want)
			}
		})
	}
}
