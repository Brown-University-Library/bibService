package sierra

import (
	"strings"
)

type MarcField struct {
	FieldTag  string              `json:"fieldTag"`
	MarcTag   string              `json:"marcTag"`
	Ind1      string              `json:"ind1"`
	Ind2      string              `json:"ind2"`
	Subfields []map[string]string `json:"subfields"`
	Content   string              `json:"content"`
}

func (f MarcField) String() string {
	return strings.Join(f.Strings(), " ")
}

func (f MarcField) Strings() []string {
	if f.Content != "" {
		return []string{f.Content}
	}
	values := []string{}
	for _, subfield := range f.Subfields {
		values = append(values, subfield["content"])
	}
	return values
}

func (f MarcField) StringsTrim() []string {
	if f.Content != "" {
		return []string{trimPunct(f.Content)}
	}
	values := []string{}
	for _, subfield := range f.Subfields {
		values = append(values, trimPunct(subfield["content"]))
	}
	return values
}

func (f MarcField) StringFor(tag string) string {
	return strings.Join(f.StringsFor(tag), " ")
}

func (f MarcField) StringsFor(tag string) []string {
	values := []string{}
	for _, subfield := range f.Subfields {
		if subfield["tag"] == tag {
			values = append(values, subfield["content"])
		}
	}
	return values
}

// Returns true if the field indicates that there are vernacular
// values associated with it. It also returns the MARC field
// where the vernacular values are.
func (f MarcField) HasVernacular() (bool, string) {
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
func (f MarcField) IsVernacularFor(target string) bool {
	for _, sub := range f.Subfields {
		if sub["tag"] == "6" {
			if strings.HasPrefix(sub["content"], target) {
				return true
			}
		}
	}
	return false
}

func (f MarcField) Values(subsWanted []string) MarcField {
	newField := MarcField{MarcTag: f.MarcTag}

	// We walk through the subfields in the Field because it is important
	// to preserve the order of the values returned according to the order
	// in which they are listed on the data, not on the spec.
	for _, fieldSub := range f.Subfields {
		for _, sub := range subsWanted {
			if fieldSub["tag"] == sub {
				content := fieldSub["content"]
				if content != "" {
					newField.Subfields = append(newField.Subfields, fieldSub)
				}
			}
		}
	}
	return newField
}