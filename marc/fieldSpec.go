package marc

import (
	"strings"
)

type FieldSpec struct {
	MarcTag   string
	Subfields []string
}

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

func NewFieldSpecs(specs string) []FieldSpec {
	fieldSpecs := []FieldSpec{}
	for _, spec := range strings.Split(specs, ":") {
		fieldSpec, ok := NewFieldSpec(spec)
		if ok {
			fieldSpecs = append(fieldSpecs, fieldSpec)
		}
	}
	return fieldSpecs
}
