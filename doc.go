// Package decimal implements an efficient, arbitrary precision, fixed-point decimal type.
//
// It is loosely based on General Decimal Arithmetic Specification
// Version 1.70 (2009)
//
// 		http://speleotrove.com/decimal/decarith.html
//
// Before using this library, it's recommended to read this document.
//
// While the library attempts to mimic math/big as much as possible
// and therefore be somewhat familiar, Decimals have some quirks
// that should be addressed before hand.
//
// Much like math/big, methods are typically of the form:
//
// 		func (d *Decimal) Unary(x *Decimal) *Decimal        // d = op x
// 		func (d *Decimal) Binary(x, y *Decimal) *Decimal    // d = x op y
// 		func (d *Decimal) M() T                       	    // d = x.M()
//
// For unary and binary operations, the result is the receiver; if it
// is one of the operands x or y it may be overwritten (and its memory
// reused). To enable chaining of operations, the result is also returned.
// Methods returning a result other than *Decimal take an
// operand as the receiver (usually named d in that case).
//
// As you would expect, the zero-value of a Decimal is simply 0.
//
// Decimals are represented as an integer value multiplied by 10
// raised to some power, referred to as the scale.
//
// 		d = v * 10 ^ scale
//
// When creating a Decimal a positive scale is the number of digits
// following the radix (decimal point). Therefore:
//
// 		New(1234, 2) == 12.34
//
// If the scale is negative it denotes the number of trailing
// zeros at the end of the Decimal. Therefore:
//
// 		New(1234, -2) == 123400
//
// Decimals can be created in a variety of ways:
//
// 		* From an integer value and integer scale
// 		* From a string
// 		* From a float
// 		* From float value and integer scale
// 		* From a big.Int
// 		* From a slice of bytes interpreted as an unsigned, big-endian integer
// 		* From another Decimal
// 		* From one integer
//
// The recommended way to create a Decimal is either with
// integers or strings as they are they only 100% accurate method
// of creating Decimals. While this library strives to make conversions
// to and from floats as accurate as possible, there are limitations
// because of how floats are stored in memory.
//
// Because floats are lossy by nature, specific values cannot be
// absolutely represented.
//
// For example, 0.1 appears to be "just" 0.1, but in reality it's
//		 0.1000000000000000055511151231257827021181583404541015625
// (See: fmt.Printf("%.55f", 0.1))
//
// In order to fix this, our method of creating Decimals from floats
// uses multiple rounds of unbiased rounding. Our error rate
// converting a float into a Decimal is roughly ± 1 ULP for 2.3% of floats.
//
// The decimal library tries to have various methods and functions
// for most basic arithmetic calculations.
// All basic calculations (addition, subtraction, multiplication, and division)
// are included.
//
// To make more advanced calculations possible, included are:
// 		* Modf function which splits a Decimal into its integral (characteristic) and fractional (mantissa) parts
// 		* Exp method which raises one Decimal to the power of another mod |m|
// 		* Jacobi function which returns the Jacobi symbol (x/y), either +1, -1, or 0
// 		* Binomial coefficient method
// 		* Sqrt method which returns the square root of a Decimal.
// 		* Hypot function which returns Sqrt(p*p + q*q)
// 		* Fib method which calculates the provided fibonacci number
// 		* Min and Max functions
// 		* Much, much more
//
// Package decimal tries to be as efficient and memory-conscious as
// possible. It reuses memory when it can, and if the decimal
// value is small enough it stores it inside an integer (instead of
// a big.Int) so that hardware arithmetic operations can be used
// instead of software.
//
// Decimal will panic if arithmetic operations exceed the bounds of
// the Decimal type.
//
// These bounds are as follows:
//
// 		Assuming d = integer * 10 ^ scale
//
//		integer: (-Inf, +Inf)
// 		scale:   (-9223372036854775808, 9223372036854775807]
//
// The bounds are chosen based on the sizes of their types.
// The integer portion is a big.Int, which is arbitrary-precision
// (solely based on the memory size of your machine).
// The scale portion is an int64 which has 64 bits of precision.
//
// As of this writing (Dec 2015) 64 bits should be more than enough
// precision for most machines. The choice of an int64 was made
// because it not only gives 1.8446744e+19 more digits of precision
// than an int32, but is also the largest native integer type Go supports.
// This means we get a wider range of decimal digits than other
// decimal libraries (e.g., Haskell only allows up to 255 digits and C# up to
// 29) as well as hardware performance.
// Using a big.Int for precision would make a true arbitrary-precision decimal,
// but would hamper performance and take up extra space.
//
// If you do need more than (2^63)-1 digits of precision, this library is
// MIT licensed so feel free to fork it and make the necessary changes.
//
// Please tell me if you do. I'd be interested to know what on earth you need more than nine quintillion digits of precision for.
//
//
// Currently there are no known bugs. If any are found, please do not
// hesitate to create an issue either via email or GitHub.
package decimal
