package sierra

import (
	"strings"
)

type Field struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
	Content   string              `json:"content"`
}

func (f Field) Tags() []string {
	tags := []string{}
	for _, sub := range f.Subfields {
		safeAppend(&tags, sub["tag"])
	}
	return tags
}

// Returns true if the field indicates that there are vernacular
// values associated with it. It also returns the MARC field
// where the vernacular values are.
func (f Field) HasVernacular() (bool, string) {
	for _, sub := range f.Subfields {
		if sub["tag"] == "6" {
			return true, sub["content"]
		}
	}
	return false, ""
}

// Returns true if the field contains vernacular values for another
// (target) field.
//
// A field is considered to have vernacular values for another field if
// it has a subfield with tag "6" where the content is for the target.
// The Target tyically comes in the form "NNN-nn" where "NNN" is the MARC
// field and "nn" is a sequence value (e.g. "700-02")
func (f Field) IsVernacularForTag6(target string) bool {
	for _, sub := range f.Subfields {
		if sub["tag"] == "6" {
			if strings.HasPrefix(sub["content"], target) {
				return true
			}
		}
	}
	return false
}

func (f Field) ValuesForTag6(subfields []string) []string {
	values := []string{}
	for _, sub := range f.Subfields {
		if in(subfields, sub["tag"]) {
			safeAppend(&values, sub["content"])
		}
	}
	return values
}

func (f Field) getSubfieldsValues(subfields []string) []string {
	values := []string{}
	// We walk through the subfields in the Field because it is important
	// to preserve the order of the values returned according to the order
	// in which they are listed on the data, not on the spec.
	for _, fieldSub := range f.Subfields {
		for _, specSub := range subfields {
			if fieldSub["tag"] == specSub {
				safeAppend(&values, fieldSub["content"])
			}
		}
	}
	return values
}
