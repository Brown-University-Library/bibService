package sierra

import (
	"testing"
)

func TestSafeAppend(t *testing.T) {
	ar1 := []string{"a", "b"}
	safeAppend(&ar1, "c")
	if len(ar1) != 3 {
		t.Errorf("Didn't append value")
	}

	safeAppend(&ar1, "b")
	if len(ar1) != 3 {
		t.Errorf("Appended duplicated value")
	}
}
