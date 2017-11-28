// Package debug provides simple routines for debugging continued fractions.
package debug

import (
	"bytes"
	"math"
)

type Term struct {
	A string
	B string
}

// TODO(eric): drop this once 1.10 becomes our minimum supported version
func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

func Dump(t []Term) string {
	max := len(t)
	if max >= 25 {
		max = 25
	}

	var length int
	for _, m := range t[:max] {
		length += len(m.B) + len(" + ")
	}

	const ddots = "⋱"
	if max < len(t) {
		length += len(ddots)
	}

	var (
		buf    bytes.Buffer
		lpad   int
		spaces = bytes.Repeat([]byte{' '}, length)
		dashes = bytes.Repeat([]byte{'-'}, length)
	)
	for _, m := range t[:max] {
		a, b := m.A, m.B

		buf.Write(spaces[:lpad])

		neg := a[0] == '-'
		if neg {
			a = a[1:]
		}

		half := int(round(float64(len(b)+length)/2.0) + round(float64(len(a))/2))
		buf.Write(spaces[:half])
		buf.WriteString(a)
		buf.Write(spaces[:half])
		buf.WriteByte('\n')

		buf.Write(spaces[:lpad])
		buf.WriteString(b)
		if neg {
			buf.WriteString(" - ")
		} else {
			buf.WriteString(" + ")
		}
		buf.Write(dashes[:length])
		buf.WriteByte('\n')

		lpad += len(b) + len(" + ")
		length -= len(b) + len(" + ")
	}

	if max < len(t) {
		buf.Write(spaces[:lpad])
		buf.WriteString(t[max].B)
		buf.WriteString(" + ")
		buf.WriteString(ddots)
	}
	return buf.String()
}
