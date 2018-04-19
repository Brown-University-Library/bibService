package sierra

import (
	"strings"
)

type FieldSpec struct {
	MarcTag   string
	Subfields []string
}

func NewFieldSpecs(spec string) []FieldSpec {
	fieldSpecs := []FieldSpec{}
	for _, token := range strings.Split(spec, ":") {
		length := len(token)
		if length < 3 {
			// not a valid spec
			continue
		}

		fieldSpec := FieldSpec{
			MarcTag:   token[0:3],
			Subfields: []string{},
		}

		if length > 3 {
			// process the subfields in the spec
			for _, c := range token[3:length] {
				fieldSpec.Subfields = append(fieldSpec.Subfields, string(c))
			}
		}
		fieldSpecs = append(fieldSpecs, fieldSpec)
	}
	return fieldSpecs
}
