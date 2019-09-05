package math

import (
	"strconv"
	"testing"

	"github.com/ericlagergren/decimal/v4"
)

func TestAsin(t *testing.T) {
	const N = 100
	eps := decimal.New(1, N)
	diff := decimal.WithPrecision(N)

	for i, tt := range [...]struct {
		x, r string
	}{
		0: {"0", "0"},
		1: {"1.00", "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499314017412671058534"},
		2: {"-1.00", "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499314017412671058534"},
		3: {"0.5", "0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331046724708903528447"},
		4: {"-0.50", "-0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331046724708903528447"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			z := decimal.WithPrecision(N)

			x, _ := new(decimal.Big).SetString(tt.x)
			r, _ := new(decimal.Big).SetString(tt.r)

			Asin(z, x)
			if z.Cmp(r) != 0 || diff.Sub(r, z).CmpAbs(eps) > 0 {
				t.Errorf(`#%d: Asin(%s)
wanted: %s
got   : %s
diff  : %s
`, i, x, r, z, diff)
			}
		})
	}
}
