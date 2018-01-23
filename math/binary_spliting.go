package math

/*
Copyright 2018 W. Nathan Hack

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
	list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
	this list of conditions and the following disclaimer in the documentation and/or
	other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be
	used to endorse or promote products derived from this software without specific
	prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT
SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
import (
	"fmt"
	stdMath "math"

	"github.com/ericlagergren/decimal"
)

/*
  The binary splitting algorithm is made of four functions
  a(n),b(n),q(n),p(n)
                                   a(n)p(0)...p(n)
	 S =  sum_(n=0)^infinity  -----------------
								   b(n)q(0)...q(n)

	We split it up into [n1,n2) slices and calculate
	using the following
	B = b(n1)...b(n2-1)
	P = p(n1)...p(n2-1)
	Q = q(n1)...q(n2-1)
	then assign
	T = BQS
	to solve for S
	S = T/BQ

	----
	The "trick" is that we plan to "binary-ly" split up the
	the range [0,n) such that for a given range [n1,n2)
	we will split it into two smaller ranges [n1,m) and [m,n2)
	where m = floor((n1+n2)/2).  When n2-n1 is either 1,2,3,4
	and we plan to calculate manually but for anything else larger
	we then define:
	For a given range [n1,n2) we split it into
	[n1,m) and [m,n2)  noting a "Left" and "Right" side
	then for each side with have a B,P,Q,and T
	we'll note the subscript via a L or R then we have
	the formulation is the following:
	B = B_l*B_r
	P = P_l*P_r
	Q = Q_l*Q_r
	T = B_l*P_l*T_r + B_r*Q_r*T_l
	(take care in noticing the Q_l and P_r aren't used in calculating T)
	then solve for S the same as above S = T/BQ
*/

//apbqBinarySplitState is used to hold intermediate values for each step in the calculation
type apbqBinarySplitState struct {
	B *decimal.Big
	P *decimal.Big
	Q *decimal.Big
	T *decimal.Big
}

//BinarySplitCalculate calculates using the APBQ version of Binary Split algorithm
// the function should be used when the number of terms is known a priori
// inclusiveNStart < exclusiveNStop
func BinarySplitCalculate(precision int, inclusiveNStart, exclusiveNStop uint64,
	A func(n uint64) *decimal.Big, P func(n uint64) *decimal.Big,
	B func(n uint64) *decimal.Big, Q func(n uint64) *decimal.Big) (*decimal.Big, error) {
	diff := exclusiveNStop - inclusiveNStart
	switch {
	case diff == 0:
		return nil, fmt.Errorf("the Start and End must be not be the same value")
	case diff < 0:
		return nil, fmt.Errorf("the End must be larger than Start")
	}

	finalState := calculate(inclusiveNStart, exclusiveNStop, precision, A, P, B, Q)
	//return T/(BQ)
	return quo(precision, finalState.T,
		mul(precision, finalState.B, finalState.Q)), nil
}

//BinarySplitDynamicCalculate calculates using the APBQ version of Binary Split algorithm
// the function should be used when the number of terms is unknown a priori it is
// somewhat slower than BinarySplitCalculate due to the checking for steady state
func BinarySplitDynamicCalculate(precision int,
	A func(n uint64) *decimal.Big, P func(n uint64) *decimal.Big,
	B func(n uint64) *decimal.Big, Q func(n uint64) *decimal.Big) *decimal.Big {

	//for this algorithm we start with a standard 16 terms to mark the first return values status,
	// then we calculate the next 4 terms and mark the difference (if ZERO return as is else repeat +4 terms until at least
	// 1 digit of precision is gained) then use that to linearly determine "last term" then repeat
	// when calculating each of the parts the following will be used:
	/*
		B = B_l*B_r
		P = P_l*P_r
		Q = Q_l*Q_r
		T = B_l*P_l*T_r + B_r*Q_r*T_l

	*/
	currentLastTerm := uint64(16)
	current := calculate(0, currentLastTerm, precision, A, P, B, Q)
	// the marked value is what should be returned which is T/(BQ)
	markValue1 := quo(precision, current.T, mul(precision, current.B, current.Q))
	//now get the next marked value, if the difference isn't already ZERO we need
	// at least one digit of precision to continue
	nextLastTerm := currentLastTerm
	next := current
	var expectedLastTerm uint64
	deltaTerm := uint64(4)
	diffToCompare := decimal.New(1, precision)

	for {
		for {
			next = combineBinarySplitLeftAndRight(precision, next,
				calculate(nextLastTerm, nextLastTerm+deltaTerm, precision, A, P, B, Q))
			nextLastTerm += deltaTerm

			markValue2 := quo(precision, next.T, mul(precision, next.B, next.Q))
			if markValue1.Cmp(markValue2) == 0 {
				//value isn't changing! we're found
				//our target end term just return the value
				return markValue2
			}
			//if not equal one of two things could be happening
			// 1)markValue2 approaching a value away from markValue1 (something not close to markValue1)
			// 2)markValue2 approaching a value toward markValue1 (something close to markValue1)

			//in the 1) case precision should stay the same but scale will change
			// in the 2) case scale & precision should stay the same but the difference should
			// see a reduction is the precision
			//we'll check for the first case since it doesn't require any "real" calculations
			if markValue1.Scale() != markValue2.Scale() {
				//there was a change so save the current state
				current = next

				//next calculate the expectedLastTerm and add 4 to ensure it is always >0
				scaleDiff := stdMath.Abs(float64(markValue1.Scale() - markValue2.Scale()))
				expectedLastTerm = nextLastTerm + uint64(float64(nextLastTerm-currentLastTerm)*float64(precision)/float64(scaleDiff)) + 4
				currentLastTerm = nextLastTerm
				break
			}
			//if not equal take the difference and figure out if we
			// have at least one digit of precision gained
			diff := decimal.WithPrecision(precision).Sub(markValue1, markValue2).Reduce()

			//here's the one case where we need to do a check for
			// something 1E-Precision if equal to or less than
			if diffToCompare.CmpAbs(diff) >= 0 {
				return markValue2
			}

			//we want to have at least 1 digit which really means we
			// need a change in precision of diff of 2 or greater

			precisionChange := int(stdMath.Abs(float64(markValue1.Precision() - diff.Precision())))
			if precisionChange > 1 {
				//we have something that we can use to
				// calculate the true expected last term
				// combine the currentState with this additional state
				// update the currentLastTerm and then calculate expectedLastTerm
				current = next

				//we'll calculate expectedLastTerm but also add 4 to ensure it is always >0
				expectedLastTerm = nextLastTerm + uint64(float64(nextLastTerm-currentLastTerm)*float64(precision)/float64(precisionChange)) + 4
				currentLastTerm = nextLastTerm
				break
			}

			//if for some reason we haven't seen the expected change
			// it could be because the markValue1 and markValue2 are extremely different
			// so we'll breakout and hope the next iteration is better
			// worse case it's not and these continues until the value converges
			// in which case markValue1 and markValue2 will at some point be equal
			if nextLastTerm-currentLastTerm > 16 {
				//save the current state
				current = next

				//and set the expected and current to nextLastTerm
				expectedLastTerm = nextLastTerm
				currentLastTerm = nextLastTerm
				break
			}
		}
		//now we have what we expect to be way closer to the true n
		if currentLastTerm != expectedLastTerm {
			current = combineBinarySplitLeftAndRight(precision,
				current, calculate(currentLastTerm, expectedLastTerm, precision, A, P, B, Q))
		}
		markValue1 = quo(precision, current.T, mul(precision, current.B, current.Q))
		currentLastTerm = expectedLastTerm
		nextLastTerm = currentLastTerm
		next = current
	}
}

func mul(precision int, val1 *decimal.Big, val2 *decimal.Big) *decimal.Big {
	return decimal.WithPrecision(precision).Mul(val1, val2)
}

func quo(precision int, val1 *decimal.Big, val2 *decimal.Big) *decimal.Big {
	return decimal.WithPrecision(precision).Quo(val1, val2)
}

func add(precision int, val1 *decimal.Big, val2 *decimal.Big) *decimal.Big {
	return decimal.WithPrecision(precision).Add(val1, val2)
}

func calculate(start, end uint64, precision int, A func(n uint64) *decimal.Big, P func(n uint64) *decimal.Big,
	B func(n uint64) *decimal.Big, Q func(n uint64) *decimal.Big) apbqBinarySplitState {

	switch end - start {
	case 1:
		n1 := start
		return apbqBinarySplitState{
			B(n1),
			P(n1),
			Q(n1),
			mul(precision, A(n1), P(n1)),
		}
	case 2:
		n1 := start
		n2 := n1 + 1
		A1, A2 := A(n1), A(n2)
		B1, B2 := B(n1), B(n2)
		P1, P2 := P(n1), P(n2)
		Q1, Q2 := Q(n1), Q(n2)
		P12 := mul(precision, P1, P2)
		return apbqBinarySplitState{
			mul(precision, B1, B2),
			P12,
			mul(precision, Q1, Q2),
			//A1*P1*B2*Q2 + B1*A2*P12,
			add(precision, mul(precision, A1, mul(precision, P1, mul(precision, B2, Q2))), mul(precision, B1, mul(precision, A2, P12))),
		}
	case 3:
		n1 := start
		n2 := n1 + 1
		n3 := n2 + 1
		A1, A2, A3 := A(n1), A(n2), A(n3)
		B1, B2, B3 := B(n1), B(n2), B(n3)
		P1, P2, P3 := P(n1), P(n2), P(n3)
		Q1, Q2, Q3 := Q(n1), Q(n2), Q(n3)
		P12 := mul(precision, P1, P2)
		P123 := mul(precision, P12, P3)
		Q23 := mul(precision, Q2, Q3)
		B23 := mul(precision, B2, B3)
		return apbqBinarySplitState{
			mul(precision, B1, B23),
			P123,
			mul(precision, Q1, Q23),
			//A1*P1*B23*Q23 + A2*B1*B3*P12*Q3 + A3*B1*B2*P123
			add(precision,
				mul(precision, A1, mul(precision, P1, mul(precision, B23, Q23))),
				add(precision,
					mul(precision, A2, mul(precision, B1, mul(precision, B3, mul(precision, P12, Q3)))),
					mul(precision, A3, mul(precision, B1, mul(precision, B2, P123))))),
		}
	case 4:
		n1 := start
		n2 := n1 + 1
		n3 := n2 + 1
		n4 := n3 + 1
		A1, A2, A3, A4 := A(n1), A(n2), A(n3), A(n4)
		B1, B2, B3, B4 := B(n1), B(n2), B(n3), B(n4)
		P1, P2, P3, P4 := P(n1), P(n2), P(n3), P(n4)
		Q1, Q2, Q3, Q4 := Q(n1), Q(n2), Q(n3), Q(n4)
		P12 := mul(precision, P1, P2)
		P123 := mul(precision, P12, P3)
		P1234 := mul(precision, P123, P4)
		Q34 := mul(precision, Q3, Q4)
		Q234 := mul(precision, Q2, Q34)
		B23 := mul(precision, B2, B3)
		return apbqBinarySplitState{
			mul(precision, B1, mul(precision, B23, B4)),
			P1234,
			mul(precision, Q1, Q234),
			//A1*P1*B23*B4*Q234 + A2*P12*B1*B3*B4*Q34 + A3*P123*B1*B2*B4*Q4 + A4*P1234*B1*B23
			add(precision,
				mul(precision, A1, mul(precision, P1, mul(precision, B23, mul(precision, B4, Q234)))),
				add(precision,
					mul(precision, A2, mul(precision, P12, mul(precision, B1, mul(precision, B3, mul(precision, B4, Q34))))),
					add(precision,
						mul(precision, A3, mul(precision, P123, mul(precision, B1, mul(precision, B2, mul(precision, B4, Q4))))),
						mul(precision, A4, mul(precision, P1234, mul(precision, B1, B23)))))),
		}
	default:
		// here we have something bigger so we'll do a binary split
		// first find the mid point between the points and create the two side
		// then do the calculations and return the value
		m := uint64((start + end) / 2)
		L := calculate(start, m, precision, A, P, B, Q)
		R := calculate(m, end, precision, A, P, B, Q)

		/*
			generically the following is done
			B = B_l*B_r
			P = P_l*P_r
			Q = Q_l*Q_r
			T = B_l*P_l*T_r + B_r*Q_r*T_l
		*/

		return combineBinarySplitLeftAndRight(precision, L, R)
	}
}

// combineBinarySplitLeftAndRight does the following and return a
// apbqBinarySplitState struct with the B,P,Q,T
// B = B_l*B_r
// P = P_l*P_r
// Q = Q_l*Q_r
// T = B_l*P_l*T_r + B_r*Q_r*T_l
func combineBinarySplitLeftAndRight(precision int, L, R apbqBinarySplitState) apbqBinarySplitState {

	B := mul(precision, L.B, R.B)
	P := mul(precision, L.P, R.P)
	Q := mul(precision, L.Q, R.Q)
	// L.B*L.P*R.T + R.B*R.Q*L.T
	T := add(precision,
		mul(precision, L.B, mul(precision, L.P, R.T)),
		mul(precision, R.B, mul(precision, R.Q, L.T)))

	return apbqBinarySplitState{B, P, Q, T}
}
