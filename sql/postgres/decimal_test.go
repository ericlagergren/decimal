package postgres

import (
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/ericlagergren/decimal"
)

var (
	maxInt, _  = new(big.Int).SetString(strings.Repeat("9", MaxIntegralDigits*2), 10)
	maxFrac, _ = new(big.Int).SetString(strings.Repeat("9", MaxFractionalDigits*2), 10)
	r          = rand.New(rand.NewSource(1))
)

func randInt(n *big.Int) string { return new(big.Int).Rand(r, n).String() }

func TestDecimal_Value(t *testing.T) {
	n := 100
	if testing.Short() {
		n = 10
	}
	for i := 0; i < n; i++ {
		ip := randInt(maxInt)
		fp := randInt(maxFrac)

		dec, ok := new(decimal.Big).SetString(fmt.Sprintf("%s.%s", ip, fp))
		if !ok {
			t.Fatal(dec.Err())
		}
		d := Decimal{V: dec, Round: i%2 == 0}

		v, err := d.Value()
		if err != nil {
			if d.Round {
				t.Fatalf("#%d: err == %#v when d.Round == true", i, err)
			}
			switch e := err.(*LengthError); e.Part {
			case "integral":
				if len(ip) != e.N {
					t.Fatalf("#%d: reported int len of %d, got %d", i, e.N, len(ip))
				}
			case "fractional":
				if len(fp) != e.N {
					t.Fatalf("#%d: reported frac len of %d, got %d", i, e.N, len(fp))
				}
			default:
				t.Fatalf("#%d: bad part: %q", i, e.Part)
			}
			continue
		}

		vs := v.(string)

		switch parts := strings.Split(vs, "."); len(parts) {
		case 2:
			var e int
			if i := strings.LastIndexAny(vs, "eE"); i > 0 {
				ev, _ := strconv.ParseInt(vs[i+1:], 10, 32)
				e -= int(ev)
				vs = vs[:i]
			}
			if len(parts[1])+e > MaxFractionalDigits {
				t.Fatalf("%#d: frac part too long: %d", i, len(parts[1]))
			}
			fallthrough
		case 1:
			if len(parts[0]) > MaxIntegralDigits {
				t.Fatalf("#%d: int part too long: %d", i, len(parts[0]))
			}
		}
	}
}
