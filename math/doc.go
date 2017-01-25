// Package math provides various useful mathematical functions and constants.
//
// NOTE: Routines in this package are under construction! The completed
// routines return correct results but have little testing and no algorithm
// has been implemented to determine how many digits of precision are necessary
// for some of the functions (notably Exp) to return results not skewed by
// rounding.
//
// Routines in this package are generally 'binary' functions, as described in
// the decimal package's documentation. However, since math cannot add methods
// to decimal's types, the functions usually have the signature
//
//     func F(z, x *T) *T
//
// compared to
//
//     func (z *T) F(x *T) *T
//
// Most functions utilize continued fractions in lieu of power series in order
// to converge quicker.
package math
