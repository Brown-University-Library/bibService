package sierra

import (
	"testing"
)

func TestGetSubfieldValues(t *testing.T) {
	lang1 := map[string]string{"content": "eng", "tag": "a"}
	lang2 := map[string]string{"content": "fre", "tag": "a"}
	lang3 := map[string]string{"content": "spa", "tag": "a"}

	field := MarcField{MarcTag: "041"}
	field.Subfields = []map[string]string{lang1, lang2, lang3}

	subfields := []string{"a"}
	values := field.Values(subfields, true)
	if len(values) != 3 {
		t.Errorf("Incorrect number of values found: %#v", values)
	}

	if !in(values, "eng") || !in(values, "fre") || !in(values, "spa") {
		t.Errorf("Expected value not found: %#v", values)
	}
}
