package decimal

import "testing"

func TestMul128(t *testing.T) {
	bad := &Big{compact: 11234567890123457890, exp: -20, precision: 20}
	ref := &Big{compact: 2, exp: -1, precision: 1}

	cmp := bad.Cmp(ref)
	if cmp != -1 {
		t.Errorf("expected %s smaller than %s\n", bad.String(), ref.String())
	}
	cmp = ref.Cmp(bad)
	if cmp != 1 {
		t.Errorf("expected %s larger than %s\n", ref.String(), bad.String())
	}
}
