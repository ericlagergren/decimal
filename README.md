# decimal [![Build Status](https://travis-ci.org/EricLagergren/decimal.png?branch=master)](https://travis-ci.org/EricLagergren/decimal)

Decima is a high-performance, arbitrary precision, fixed-point decimal library.

## Features

 * Zero-value is 0 and is safe to use without initialization
 * Addition, subtraction, and multiplication with no loss of precision
 * Division with specified precision
 * database/sql serialization/deserialization
 * JSON, XML, and Gob serialization/deserialization

 TODO:
 * Many useful functions and methods like Sqrt, Hypot, and Jacobi

## Install

`go get github.com/EricLagergren/decimal`

## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/EricLagergren/decimal"
)

// It's all very similar to math/big's API.
func main() {
	price, err := decimal.NewFromString("136.02")
    if err != nil {
        log.Fatalln(err)
    }

	quantity := decimal.NewFromFloat(3)

    fee, _ := decimal.NewFromString(".035")
    taxRate, _ := decimal.NewFromString(".08875")

    subtotal := new(decimal.Decimal).Mul(price, quantity)

    preTax := new(decimal.Decimal).Mul(subtotal, fee.Add(fee, decimal.New(1, 0)))

    total := new(decimal.Decimal).Mul(preTax, taxRate.Add(taxRate, decimal.New(1, 0)))

    fmt.Println("Subtotal:", subtotal)                                                   // Subtotal: 408.06
    fmt.Println("Pre-tax:", preTax)                                                      // Pre-tax: 422.3421
    fmt.Println("Taxes:", total.Sub(total, preTax))                                      // Taxes: 37.482861375
    fmt.Println("Total:", total)                                                         // Total: 459.824961375
    fmt.Println("Tax rate:", new(decimal.Decimal).Sub(total, preTax).Quo(total, preTax)) // Tax rate: 0.08875
}
```

## Documentation

http://godoc.org/github.com/EricLagergren/decimal

## License

The [BSD 3 License](https://github.com/EricLagergren/decimal/blob/master/LICENSE)
