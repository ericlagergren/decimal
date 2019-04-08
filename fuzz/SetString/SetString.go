package fuzz

import (
	"fmt"
	"runtime/debug"

	"github.com/ericlagergren/decimal"
)

func Fuzz(data []byte) int {
	debug.SetTraceback("system")

	d := new(decimal.Big)
	d, ok := d.SetString(string(data))
	if !ok {
		if decimal.Regexp.Match(data) && d.Context.Err() == nil {
			panic(fmt.Sprintf("should work: %q", data))
		}
		return 0
	}
	d2 := new(decimal.Big)
	ds := d.String()
	d2, ok = d2.SetString(ds)
	if !ok {
		panic(fmt.Sprintf("SetString(%q) == nil, false", ds))
	}
	if d.Cmp(d2) != 0 {
		panic(fmt.Sprintf(`
got   : %#v (%q)
wanted: %#v (%q)
`, d2, d2, d, d))
	}
	if !decimal.Regexp.Match(data) {
		panic(fmt.Sprintf("got: %q", data))
	}
	return 1
}
