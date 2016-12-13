# decimal [![Build Status](https://travis-ci.org/ericlagergren/decimal.png?branch=master)](https://travis-ci.org/ericlagergren/decimal) [![GoDoc](https://godoc.org/github.com/ericlagergren/decimal?status.svg)](https://godoc.org/github.com/ericlagergren/decimal)

Decimal is a high-performance, arbitrary precision, fixed-point decimal library.

Note: The `math/` subpackage is under development and should be used with care.

## Features

 * Zero-value is 0 and is safe to use without initialization.
 * Addition, subtraction, and multiplication with no loss of precision.
 * Division with specified precision.
 * JSON and XML serialization and deserialization.
 * High performance

 TODO:
 * Useful math functions and methods like Exp, Log, Sqrt, Hypot, and Jacobi

## Install

`go get github.com/ericlagergren/decimal`

## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/ericlagergren/decimal"
)

// It's all very similar to math/big's API.
func main() {
    price := decimal.New(13602, 2)

    quantity := new(decimal.Big).SetFloat(3)

    fee, ok := new(decimal.Big).SetString(".035")
    if !ok {
        // handle invalid decimal
    }

    taxRate, ok := new(decimal.Big).SetString(".08875")
    if !ok {
        // handle invalid decimal
    }

    subtotal := new(decimal.Big).Mul(price, quantity)

    preTax := new(decimal.Big).Mul(subtotal, fee.Add(fee, decimal.New(1, 0)))

    total := new(decimal.Big).Mul(preTax, taxRate.Add(taxRate, decimal.New(1, 0)))

    fmt.Println("Subtotal:", subtotal)                                               // Subtotal: 408.06
    fmt.Println("Pre-tax:", preTax)                                                  // Pre-tax: 422.3421
    fmt.Println("Taxes:", total.Sub(total, preTax))                                  // Taxes: 37.482861375
    fmt.Println("Total:", total)                                                     // Total: 459.824961375
    fmt.Println("Tax rate:", new(decimal.Big).Sub(total, preTax).Quo(total, preTax)) // Tax rate: 0.08875
}
```

## Documentation

http://godoc.org/github.com/ericlagergren/decimal

## License

The [BSD 3 License](https://github.com/ericlagergren/decimal/blob/master/LICENSE)
