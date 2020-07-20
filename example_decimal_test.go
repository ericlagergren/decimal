package decimal_test

import (
	"fmt"

	"github.com/ericlagergren/decimal"
)

func ExampleBig_Format() {
	print := func(format, xs string) {
		x, _ := new(decimal.Big).SetString(xs)
		fmt.Printf(format+"\n", x)
	}

	print("%.3g", "12.34")
	print("%.1f", "12.34")
	print("`%6.4g`", "500.44")
	print("'%-10.f'", "-404.040")
	// Output:
	// 12.3
	// 12.3
	// ` 500.4`
	// '-404      '
}

func ExampleBig_Precision() {
	a := decimal.New(12, 0)
	b := decimal.New(42, -2)
	c := decimal.New(12345, 3)
	d := decimal.New(3, 5)

	fmt.Printf(`
%g has a precision of %d
%g has a precision of %d
%g has a precision of %d
%g has a precision of %d
`, a, a.Precision(), b, b.Precision(), c, c.Precision(), d, d.Precision())
	// Output:
	//
	// 12 has a precision of 2
	// 4.2e+03 has a precision of 2
	// 12.345 has a precision of 5
	// 0.00003 has a precision of 1
}

func ExampleBig_Round() {
	a, _ := new(decimal.Big).SetString("1234")
	b, _ := new(decimal.Big).SetString("54.4")
	c, _ := new(decimal.Big).SetString("60")
	d, _ := new(decimal.Big).SetString("0.0022")

	fmt.Println(a.Round(2))
	fmt.Println(b.Round(2))
	fmt.Println(c.Round(5))
	fmt.Println(d.Round(1))
	// Output:
	// 1.2e+03
	// 54
	// 60
	// 0.002
}

func ExampleBig_Quantize() {
	a, _ := decimal.WithContext(decimal.Context32).SetString("2.17")
	b, _ := decimal.WithContext(decimal.Context64).SetString("217")
	c, _ := decimal.WithContext(decimal.Context128).SetString("-0.1")
	d, _ := decimal.WithContext(decimal.Context{OperatingMode: decimal.GDA}).SetString("-0")

	fmt.Printf("A: %g\n", a.Quantize(3))
	fmt.Printf("B: %g\n", b.Quantize(-2))
	fmt.Printf("C: %g\n", c.Quantize(1))
	fmt.Printf("D: %g\n", d.Quantize(-5))
	// Output:
	// A: 2.17
	// B: 2e+02
	// C: -0.1
	// D: -0e+05
}
