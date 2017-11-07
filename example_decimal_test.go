package decimal

import (
	"fmt"
	"strings"
)

func ExampleBig_Format() {
	var (
		modeGo  = Context{OperatingMode: Go}
		modeGDA = Context{OperatingMode: GDA}
	)
	print := func(format, xs string) {
		var x *Big
		if strings.HasPrefix(xs, "Go: ") {
			x, _ = WithContext(modeGo).SetString(strings.TrimPrefix(xs, "Go: "))
		} else if strings.HasPrefix(xs, "GDA: ") {
			x, _ = WithContext(modeGDA).SetString(strings.TrimPrefix(xs, "GDA: "))
		} else {
			x, _ = WithContext(modeGDA).SetString(xs)
		}
		fmt.Printf(format+"\n", x)
	}

	print("%s", "Go: 12.34")
	print("%s", "GDA: 12.34")
	print("%.3g", "12.34")
	print("%.1f", "12.34")
	print("`%6.4g`", "500.44")
	print("'%-10.f'", "-404.040")
	// Output:
	// 12.34
	// 12.34
	// 12.3
	// 12.3
	// ` 500.4`
	// '-404.040  '
}

func ExampleBig_Precision() {
	a := New(12, 0)
	b := New(42, -2)
	c := New(12345, 3)
	d := New(3, 5)

	fmt.Printf(`
%s has a precision of %d
%s has a precision of %d
%s has a precision of %d
%s has a precision of %d
`, a, a.Precision(), b, b.Precision(), c, c.Precision(), d, d.Precision())
	// Output:
	//
	// 12 has a precision of 2
	// 4.2E+3 has a precision of 2
	// 12.345 has a precision of 5
	// 0.00003 has a precision of 1
}

func ExampleBig_Round() {
	a, _ := new(Big).SetString("1234")
	b, _ := new(Big).SetString("54.4")
	c, _ := new(Big).SetString("60")
	d, _ := new(Big).SetString("0.0022")

	fmt.Println(a.Round(2))
	fmt.Println(b.Round(2))
	fmt.Println(c.Round(5))
	fmt.Println(d.Round(1))
	// Output:
	// 1.2e+3
	// 54
	// 60
	// 0.002
}

func ExampleBig_Quantize() {
	a, _ := WithContext(Context32).SetString("2.17")
	b, _ := WithContext(Context64).SetString("217")
	c, _ := WithContext(Context128).SetString("-0.1")
	d, _ := WithContext(Context{OperatingMode: GDA}).SetString("-0")

	fmt.Printf("A: %s\n", a.Quantize(3)) // 3 digits after radix
	fmt.Printf("B: %s\n", b.Quantize(-2))
	fmt.Printf("C: %s\n", c.Quantize(1))
	fmt.Printf("D: %s\n", d.Quantize(-5))
	// Output:
	// A: 2.170
	// B: 2E+2
	// C: -0.1
	// D: -0E+5
}
