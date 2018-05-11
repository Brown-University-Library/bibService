package marc

import (
	"testing"
)

func TestTrimPunct(t *testing.T) {
	if TrimPunct("one hundred/ ") != "one hundred" {
		t.Errorf("Failed to remove trailing slash")
	}

	if TrimPunct("one.") != "one" {
		t.Errorf("Failed to remove trailing period")
	}

	if TrimPunct("ct.") != "ct." {
		t.Errorf("Removed trailing period")
	}

	if TrimPunct("[hello") != "hello" {
		t.Errorf("Failed to remove square bracket")
	}

	if TrimPunct("[hello]") != "hello" {
		t.Errorf("Failed to remove square brackets")
	}

	if TrimPunct("[hello [world]]") != "[hello [world]]" {
		t.Errorf("Removed square brackets")
	}
}
