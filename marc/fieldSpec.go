package marc

import (
	"strings"
)

// Represents a field specification
type FieldSpec struct {
	MarcTag   string   // MARC tag for the spec
	Subfields []string // Subfields to use
}

// Creates a new FieldSpec from a string. The string is meant to be
// in the form "nnnabc" where "nnn" represents the MARC tag and "abc"
// represent the subfields to include. There could be zero or more subfields.
func NewFieldSpec(spec string) (FieldSpec, bool) {
	length := len(spec)
	if length < 3 {
		// not a valid spec
		return FieldSpec{}, false
	}

	fieldSpec := FieldSpec{
		MarcTag:   spec[0:3],
		Subfields: []string{},
	}

	if length > 3 {
		// process the subfields in the spec
		for _, c := range spec[3:length] {
			fieldSpec.Subfields = append(fieldSpec.Subfields, string(c))
		}
	}
	return fieldSpec, true
}

// Creates an array of FieldSpecs from the given spec string. Multiple
// specs can be indicated separated by a colon (e.g. "245:505abc")
func NewFieldSpecs(specsStr string) []FieldSpec {
	fieldSpecs := []FieldSpec{}
	for _, specStr := range strings.Split(specsStr, ":") {
		fieldSpec, ok := NewFieldSpec(specStr)
		if ok {
			fieldSpecs = append(fieldSpecs, fieldSpec)
		}
	}
	return fieldSpecs
}
