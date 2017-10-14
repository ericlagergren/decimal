package decimal

import "fmt"

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
	// 4.2e+3 has a precision of 2
	// 12.345 has a precision of 5
	// 0.00003 has a precision of 1
}

func ExampleBig_Round() {
	a := New(1234, 0).Round(2)
	b := New(544, 1).Round(2)
	c := New(60, 0).Round(5)

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	// Output:
	// 1.2e+3
	// 54
	// 60
}
