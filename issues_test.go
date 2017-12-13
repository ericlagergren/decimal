package decimal

import (
	"testing"
)

func Test_Issue20(t *testing.T) {
	x := New(10240000000000, 0)
	x.Mul(x, New(976563, 9))
	if v, _ := x.Int64(); v != 10000005120 {
		t.Error("error int64: ", v, x.Int(nil).Int64())
	}
}
