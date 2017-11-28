package math

import "github.com/ericlagergren/decimal"

// Log sets z to the natural logarithm of x.
func Log(z, x *decimal.Big) *decimal.Big {
	x0 := new(decimal.Big).Copy(x)
	x0.Sub(x0, one)
	g := lng{
		recv: alias(z, x),
		z:    x0,
		t:    Term{A: new(decimal.Big).Set(x0), B: new(decimal.Big)},
	}
	return z.Quo(x0, Lentz(z, &g))
}

type lng struct {
	recv *decimal.Big // reciever in Ln, can be nil
	z    *decimal.Big // input value
	k    int64        // term number
	a    int64
	t    Term
}

func (l *lng) Next() bool { return true }

var LP = 16

func (l *lng) Lentz() (f, Δ, C, D, eps *decimal.Big) {
	prec := precision(l.recv)

	f = new(decimal.Big)
	Δ = new(decimal.Big)
	C = new(decimal.Big)
	D = new(decimal.Big)
	eps = decimal.New(1, prec-1)

	// TODO(eric): (mathematically) figure out why we need this extra precision.
	//n := 1 + int32(math.Ceil(float64(prec)/25))
	//fmt.Println(n)
	C.Context.Precision = LP //prec + n)
	D.Context.Precision = LP //prec + n)
	return
}

func (l *lng) Term() Term {
	// References:
	//
	// [Cuyt] - Cuyt, A.; Petersen, V.; Brigette, V.; Waadeland, H.; Jones, W.B.
	// (2008). Handbook of Continued Fractions for Special Functions. Springer
	// Netherlands. https://doi.org/10.1007/978-1-4020-6949-9

	l.k++
	if l.k&1 == 0 && l.k != 1 {
		l.a++
		l.t.A.SetMantScale(l.a*l.a, 0)
		l.t.A.Mul(l.t.A, l.z)
	}
	l.t.B.SetMantScale(l.k, 0)
	return l.t
}
