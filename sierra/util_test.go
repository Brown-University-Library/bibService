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

func TestTrimPunct(t *testing.T) {
	if trimPunct("one hundred/ ") != "one hundred" {
		t.Errorf("Failed to remove trailing slash")
	}

	if trimPunct("one.") != "one" {
		t.Errorf("Failed to remove trailing period")
	}

	if trimPunct("ct.") != "ct." {
		t.Errorf("Removed trailing period")
	}

	if trimPunct("[hello") != "hello" {
		t.Errorf("Failed to remove square bracket")
	}

	if trimPunct("[hello]") != "hello" {
		t.Errorf("Failed to remove square brackets")
	}

	if trimPunct("[hello [world]]") != "[hello [world]]" {
		t.Errorf("Removed square brackets")
	}
}
