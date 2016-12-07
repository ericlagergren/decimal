// +build go1.7

package decimal

import "testing"

func TestSetString(t *testing.T) {
	tests := []struct {
		dec string
		s   string
	}{
		{"0", "0"},
		{"00000000000000000000", "0"},
	}
	for _, v := range tests {
		t.Run(v.dec, func(t *testing.T) {
			dec := newbig(t, v.dec)
			s := dec.String()
			if v.s != s {
				t.Fatalf("wanted %s, got %s", v.s, s)
			}
		})
	}
}
