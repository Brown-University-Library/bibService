package sierra

import (
	"testing"
)

func TestContent(t *testing.T) {
	a := map[string]string{"tag": "a", "content": "aaa"}
	b := map[string]string{"tag": "a", "content": "bbb"}
	field := MarcField{MarcTag: "041"}
	field.Content = "hello world,"
	field.Subfields = []map[string]string{a, b}

	if field.String() != "hello world," {
		t.Errorf("Failed to pick content over subfields")
	}

	value := field.StringsTrim()[0]
	if value != "hello world" {
		t.Errorf("Failed to pick content over subfields: %#v", value)
	}
}

func TestSubfieldValues(t *testing.T) {
	lang1 := map[string]string{"tag": "a", "content": "eng"}
	lang2 := map[string]string{"tag": "a", "content": "fre"}
	lang3 := map[string]string{"tag": "a", "content": "spa"}
	extra := map[string]string{"tag": "x", "content": "xxx"}
	field := MarcField{MarcTag: "041"}
	field.Subfields = []map[string]string{lang1, extra, lang2, lang3}

	subfields := []string{"a"}
	values := field.Values(subfields)
	if len(values.Subfields) != 3 {
		t.Errorf("Incorrect number of values found: %#v", values)
	}

	langs := values.Strings()
	if langs[0] != "eng" || langs[1] != "fre" || langs[2] != "spa" {
		t.Errorf("Expected value not found: %#v", langs)
	}
}
