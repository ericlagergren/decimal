// Package decimal implements arbitrary precision, decimal floating-point numbers.
//
// Overview
//
// Decimal implements decimals according to the General Decimal Arithmetic[0]
// (GDA) specification, version 1.70.
//
// Decimal arithmetic is useful for financial programming or calculations (like
// in CAD) where larger, more accurate representations of numbers are required.
//
// Usage
//
// The following type is supported:
//
//     Big (arbitrary precision) decimal numbers
//
// The zero value for a Big corresponds with 0, meaning all the following are
// valid:
//
//     var x Big
//     y := new(Big)
//     z := &Big{}
//
// Method naming is the same as the ``math/big'' package', meaning:
//
//     func (z *T) SetV(v V) *T           // z = v
//     func (z *T) Unary(x *T) *T         // z = unary x
//     func (z *T) Binary(x, y *T) *T     // z = x binary y
//     func (z *T) Ternary(x, y, u *T) *T // z = x ternary y ternary u
//     func (x *T) Pred() P               // p = pred(x)
//
// Arguments are allowed to alias, meaning the following is valid:
//
//     x := New(1, 0)
//     x.Add(x, x) // x == 2
//
//     y := New(1, 0)
//     y.FMA(y, x, y) // y == 3
//
// Unless otherwise specified, the only argument that will be modified is the
// result, typically a receiver named ``z''. This means the following is valid
// and race-free:
//
//    x := New(1, 0)
//    var z1, z2 Big
//
//    go func() { z1.Add(x, x) }()
//    go func() { z2.Add(x, x) }()
//
// However, this is not:
//
//    x := New(1, 0)
//    var z Big
//
//    go func() { z.Add(x, x) }() // BAD! RACE CONDITION!
//    go func() { z.Add(x, x) }() // BAD! RACE CONDITION!
//
// Go's conventions differ from the GDA specification in a few key areas. For
// example, ``math/big.Float'' panics on NaN values, while the GDA specification
// does not require the exception created by the NaN value to be trapped.
//
// Because of this, there are two operating modes that allow users to better
// select their desired behavior:
//
//     GDA: strictly adhere to the GDA specification (default)
//     Go: utilize Go idioms, more flexibility
//
// Users can specify explicit contexts for arithmetic operations, and though
// while recommended, this is not required. Decimal also provides access to NaN
// payloads and is more lenient when parsing a decimal from a string than the
// GDA specification requires.
//
// In addition to basic arithmetic operations (addition, subtraction,
// multiplication, and division) this package offers a variety of mathematical
// functions, including the logarithms and continued fractions. (See:
// ``decimal/math''.)
//
// References:
//
//    - [0]: http://speleotrove.com/decimal/decarith.html
//
package decimal
