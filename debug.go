package decimal

const debug = false

func (d *Decimal) validate() {
	if !debug {
		panic("validate called but debug is false")
	}
}
