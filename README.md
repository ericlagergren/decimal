# decimal [![Build Status](https://travis-ci.org/EricLagergren/decimal.png?branch=master)](https://travis-ci.org/EricLagergren/decimal)

Package decimal implements an efficient, arbitrary precision, fixed-point decimal type.

## Features

 * Zero-value is 0 and is safe to use without initialization
 * Addition, subtraction, and multiplication with no loss of precision
 * Division with specified precision
 * Many useful functions and methods like Sqrt, Hypot, and Jacobi
 * database/sql serialization/deserialization
 * JSON, XML, and Gob serialization/deserialization
 * Includes a FizzBuzz function. (Biggest selling point right here.)

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

## FAQ

#### Why don't you just use float64/big.Float?

Because float64s (or any binary floating point type, actually) can't represent
numbers such as 0.1 exactly.

Consider this code: http://play.golang.org/p/TQBd4yJe6B You might expect that
it prints out `10`, but it actually prints `9.999999999999831`. Over time,
these small errors can really add up!

#### Why don't you just use big.Rat?

big.Rat is fine for representing rational numbers but Decimal is better for
representing money. Why? Here's a (contrived) example:

Let's say you use big.Rat and you have two numbers, x and y, both
representing 1/3, and you have `z = 1 - x - y = 1/3`. If you print each one
out, the string output has to stop somewhere (let's say it stops at 3 decimal
digits, for simplicity), so you'll get 0.333, 0.333, and 0.333. But where did
the other 0.001 go?

Here's the above example as code: http://play.golang.org/p/lCZZs0w9KE

With Decimal, the strings being printed out represent the number exactly. So,
if you have `x = y = 1/3` (with precision 3), they will actually be equal to
0.333, and when you do `z = 1 - x - y`, `z` will be equal to .334. No money is
unaccounted for!

You still have to be careful. If you want to split a number `N` 3 ways, you
can't just send `N/3` to three different people. You have to pick one to send
`N - (2/3*N)` to. That person will receive the fraction of a penny remainder.

It is much easier to be careful with Decimal than with big.Rat.

## License

The [MIT License (MIT)](https://github.com/EricLagergren/decimal/blob/master/LICENSE)