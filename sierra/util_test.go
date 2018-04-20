package sierra

import (
	"testing"
)

func TestSafeAppend(t *testing.T) {
	array := []string{"a", "b"}
	safeAppend(&array, "c")
	if len(array) != 3 {
		t.Errorf("Didn't append value")
	}

	safeAppend(&array, "b")
	if len(array) != 3 {
		t.Errorf("Appended duplicated value")
	}

	safeAppend(&array, "")
	if len(array) != 3 {
		t.Errorf("Appended empty value")
	}
}
