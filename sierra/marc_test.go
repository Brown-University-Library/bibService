package sierra

import (
	"testing"
)

func TestPubYear008X(t *testing.T) {
	test1 := "760629c19749999ne tr pss o   0   a0eng  cas   "
	year, ok := pubYear008(test1, 15)
	if !ok || year != 1974 {
		t.Errorf("Failed on %s (%v, %v)", test1, ok, year)
	}

	test2 := "061108c200u9999nyuar ss 0 0eng ccas a "
	year, ok = pubYear008(test2, 15)
	if !ok || year != 2005 {
		t.Errorf("Failed on %s (%v, %v)", test2, ok, year)
	}

	// Eventually we want to handle this.
	test3 := "061108duuuu2002nyuar ss 0 0eng ccas a "
	year, ok = pubYear008(test3, 15)
	if ok {
		t.Errorf("Failed on %s (%v, %v)", test3, ok, year)
	}
}
