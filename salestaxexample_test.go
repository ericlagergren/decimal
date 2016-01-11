package decimal_test

import (
	"fmt"
	"log"

	. "github.com/EricLagergren/decimal"
)

func Example() {
	price, err := NewFromString("136.02")
	if err != nil {
		log.Fatalln(err)
	}

	quantity := NewFromFloat(3)

	fee, _ := NewFromString(".035")
	taxRate, _ := NewFromString(".08875")

	subtotal := new(Decimal).Mul(price, quantity)

	preTax := new(Decimal).Mul(subtotal, fee.Add(fee, New(1, 0)))

	total := new(Decimal).Mul(preTax, taxRate.Add(taxRate, New(1, 0)))

	fmt.Println("Subtotal:", subtotal)                                           // Subtotal: 408.06
	fmt.Println("Pre-tax:", preTax)                                              // Pre-tax: 422.3421
	fmt.Println("Taxes:", total.Sub(total, preTax))                              // Taxes: 37.482861375
	fmt.Println("Total:", total)                                                 // Total: 459.824961375
	fmt.Println("Tax rate:", new(Decimal).Sub(total, preTax).Div(total, preTax)) // Tax rate: 0.08875
}
